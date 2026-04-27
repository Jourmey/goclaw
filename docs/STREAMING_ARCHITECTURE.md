# 流式输出架构

## 简答

**流式输出有两个层次，都被正确触发：**

| 层次 | 触发点 | 路径 | 事件类型 |
|------|--------|------|---------|
| **1. Token 级别流式** | `makeCallLLM()` | LLM 提供商 → callback → emit | `ChatEventChunk` / `ChatEventThinking` |
| **2. 块状回复** | `ThinkStage:150-151` | LLM 完整响应 → EmitBlockReply | `AgentEventBlockReply` |

---

## 完整流式输出链路

### 1️⃣ Token 级别流式输出（实时 token 流）

**位置：** `internal/agent/loop_pipeline_callbacks.go:216-311` (makeCallLLM)

```go
func (l *Loop) makeCallLLM(req *RunRequest, emitRun func(AgentEvent)) func(...) (*providers.ChatResponse, error) {
    return func(ctx context.Context, state *pipeline.RunState, chatReq providers.ChatRequest) (*providers.ChatResponse, error) {
        // 第 265-283 行：流式路径
        if req.Stream {  // ← 流式标志
            resp, err = provider.ChatStream(ctx, chatReq, func(chunk providers.StreamChunk) {
                // 每个 token/完成事件到来时触发

                // 1a. Thinking token（推理模型特定）
                if chunk.Thinking != "" {
                    emitRun(AgentEvent{
                        Type:    protocol.ChatEventThinking,  // ← 思考块
                        AgentID: l.id,
                        RunID:   req.RunID,
                        Payload: map[string]string{"content": chunk.Thinking},
                    })
                }

                // 1b. Content token（文本输出）
                if chunk.Content != "" {
                    emitRun(AgentEvent{
                        Type:    protocol.ChatEventChunk,     // ← 内容块
                        AgentID: l.id,
                        RunID:   req.RunID,
                        Payload: map[string]string{"content": chunk.Content},
                    })
                }
            })
        }

        // 第 284-306 行：非流式路径（后置发送）
        } else {
            resp, err = provider.Chat(ctx, chatReq)  // 同步调用，等待完整响应
        }

        // 非流式模式下，手动发送事件
        if !req.Stream && err == nil && resp != nil {
            if resp.Thinking != "" {
                emitRun(AgentEvent{
                    Type: protocol.ChatEventThinking,
                    ... payload: resp.Thinking,
                })
            }
            if resp.Content != "" {
                emitRun(AgentEvent{
                    Type: protocol.ChatEventChunk,
                    ... payload: resp.Content,
                })
            }
        }
    }
}
```

**触发机制：**
```
Client 请求 req.Stream = true
  ↓
CallLLM (Loop.makeCallLLM)
  ├─ provider.ChatStream(ctx, req, callback)  ← 开启实时流
  └─ callback() 每个 token 调用
     ├─ emitRun(ChatEventThinking)  ← 思考 token（Claude 3.5 Sonnet 等）
     └─ emitRun(ChatEventChunk)     ← 内容 token（所有模型）
        ↓
        Gateway 事件总线
        ├─ 转换为 WS 消息
        └─ 发送给客户端 (real-time token 流)
```

### 2️⃣ 块状回复（工具迭代期间的中间回复）

**位置：** `internal/pipeline/think_stage.go:128-154`

```go
func (s *ThinkStage) Execute(ctx context.Context, state *RunState) error {
    // ... LLM 调用完成，获得 resp ...

    // 第 132-135 行：最终答案（无工具调用）
    if len(resp.ToolCalls) == 0 {
        s.result = BreakLoop  // 完成迭代
        return nil  // 不发出块状回复（避免重复）
    }

    // 第 137-146 行：工具迭代（有工具调用）
    assistantMsg := providers.Message{
        Role:       "assistant",
        Content:    resp.Content,
        Thinking:   resp.Thinking,
        ToolCalls: resp.ToolCalls,  // ← 工具调用信息
    }
    state.Messages.AppendPending(assistantMsg)  // 追加到消息历史

    // 第 148-152 行：发出块状回复 ← 关键！
    // 注释："Emit block.reply for intermediate assistant content during tool iterations.
    //         Non-streaming channels (Zalo, Discord, WhatsApp) need this for delivery."
    if resp.Content != "" && s.deps.EmitBlockReply != nil {
        s.deps.EmitBlockReply(resp.Content)  // ← 发送给客户端
    }

    return nil
}
```

**什么时候发出：**
- ✅ **有工具调用** + 有中间文本 → 发出块状回复（让用户知道 Agent 在思考）
- ❌ **无工具调用** → 跳过（避免重复，最终答案由 FinalizeStage 处理）

**示例：**
```
迭代 0:
  LLM 返回: "我现在要读取文件..." + ToolCalls: [read_file]
  ThinkStage: EmitBlockReply("我现在要读取文件...")  ← 流式发送给用户
  然后执行工具

迭代 1:
  LLM 返回: "好的，任务完成" + ToolCalls: []  (空)
  ThinkStage: 不发出块状回复 (result = BreakLoop)
  FinalizeStage: 发出最终答案
```

---

## 两层流式的关系

```
Timeline:

时间线 →

迭代 #0:

  ThinkStage:
    ├─ CallLLM(Stream=true)  [makeCallLLM]
    │  ├─ provider.ChatStream()
    │  ├─ [Token 1] → emitRun(ChatEventChunk: "我") ← 第1层：实时流
    │  ├─ [Token 2] → emitRun(ChatEventChunk: "现") ← 第1层：实时流
    │  ├─ [Token 3] → emitRun(ChatEventChunk: "在") ← 第1层：实时流
    │  └─ [完成] → resp = {Content: "我现在要...", ToolCalls: [read_file]}
    │
    └─ EmitBlockReply(resp.Content)  [think_stage:150]  ← 第2层：块状回复
       (用于非流式 channel，或作为块状汇总)

时间线 →

迭代 #1:

  ToolStage:
    ├─ 执行 read_file
    └─ 追加工具结果到消息

  ThinkStage:
    ├─ CallLLM(Stream=true)  [makeCallLLM]
    │  ├─ [Token 1] → emitRun(ChatEventChunk: "好") ← 第1层：实时流
    │  └─ [完成] → resp = {Content: "好的，任务完成", ToolCalls: []}
    │
    └─ (无 EmitBlockReply，因为 ToolCalls 为空)  ← 第2层：跳过

时间线 →

FinalizeStage:
  └─ 发出最终答案 + 其他元数据
```

---

## 三种事件类型

### 1. `ChatEventChunk` - Token 流（来自 provider）

```go
// makeCallLLM 中发出
emitRun(AgentEvent{
    Type: protocol.ChatEventChunk,
    Payload: {"content": "单个 token 或多个字符"},
})
```

**特性：**
- 来自 LLM provider 的实时流
- 每个 token/chunk 立即发送
- 用户看到实时打字效果
- 网关 dedup 检查会处理重复

### 2. `ChatEventThinking` - 思考块（来自 provider）

```go
// makeCallLLM 中发出（仅当 chunk.Thinking != ""）
emitRun(AgentEvent{
    Type: protocol.ChatEventThinking,
    Payload: {"content": "Claude 的内部思考..."},
})
```

**特性：**
- 仅当模型返回思考块时发送
- Claude 3.5 Sonnet、o3 等推理模型特定
- 可选的中间推理过程可见化

### 3. `AgentEventBlockReply` - 块状回复（来自 ThinkStage）

```go
// think_stage.go 中发出
cb.emitBlockReply(resp.Content)  // EmitBlockReply 回调
```

**特性：**
- 工具迭代期间的中间回复
- 用于非流式 channel（Zalo, Discord, WhatsApp）
- 或作为流式 token 的块状汇总

---

## 调用链总结

```
┌─────────────────────────────────────────────────────────┐
│ Client 请求: req.Stream = true/false                   │
└──────────────────┬──────────────────────────────────────┘
                   ↓
         ┌─────────────────────┐
         │  Agent.Run()        │
         │  (internal/agent)   │
         └──────────┬──────────┘
                    ↓
      ┌─────────────────────────────────────┐
      │ Pipeline.Run()                      │
      │ (internal/pipeline)                 │
      │                                     │
      ├─ Setup: ContextStage               │
      │                                     │
      ├─ Iteration Loop:                   │
      │  ├─ ThinkStage  ← 流式输出第 1 层  │
      │  │  ├─ CallLLM()                   │
      │  │  │  └─ provider.ChatStream()    │
      │  │  │     ├─ Token → ChatEventChunk│
      │  │  │     └─ Think → ChatEventThinking
      │  │  │                               │
      │  │  └─ EmitBlockReply() ← 第 2 层  │
      │  │     (工具迭代中间回复)          │
      │  │                                 │
      │  ├─ PruneStage                     │
      │  ├─ ToolStage                      │
      │  ├─ ObserveStage                   │
      │  └─ CheckpointStage                │
      │                                    │
      └─ Finalize: FinalizeStage          │
         (最终答案发送)                     │
└─────────────────────────────────────────┘
```

---

## 流式 vs 非流式的区别

### 流式模式 (req.Stream = true)

```go
// makeCallLLM:265-283
provider.ChatStream(ctx, chatReq, func(chunk providers.StreamChunk) {
    // 实时 token 回调
    if chunk.Content != "" {
        emitRun(AgentEvent{Type: ChatEventChunk, ...})  // ← 立即发送
    }
})
```

**行为：**
- 客户端实时接收 token
- 用户看到逐字打字效果
- 降低感知延迟

### 非流式模式 (req.Stream = false)

```go
// makeCallLLM:284-306
resp, err = provider.Chat(ctx, chatReq)  // 等待完整响应

// 然后手动发送
if resp.Content != "" {
    emitRun(AgentEvent{Type: ChatEventChunk, Payload: resp.Content})  // ← 一次性发送
}
```

**行为：**
- 等待完整 LLM 响应
- 一次性发送全部内容
- 更简单但延迟更高

---

## 关键代码位置

| 功能 | 文件 | 行号 | 说明 |
|------|------|------|------|
| **Token 流（Stream=true）** | `loop_pipeline_callbacks.go` | 265-283 | ChatStream callback 发送 chunk |
| **Token 流（Stream=false）** | `loop_pipeline_callbacks.go` | 289-306 | Chat 同步，手动发送事件 |
| **块状回复** | `think_stage.go` | 150-151 | EmitBlockReply() 发送中间回复 |
| **流程控制** | `think_stage.go` | 132-135 | 无工具调用时跳过块状回复 |
| **消息追加** | `think_stage.go` | 137-146 | 工具迭代消息追加到历史 |

---

## 网关层处理

流式事件到达网关后的处理（伪代码）：

```go
// internal/gateway/methods/chat.go (推测)

for event := range agent.EventChannel() {
    switch event.Type {
    case ChatEventThinking:
        // 转换为 WS thinking 帧
        ws.SendFrame(protocol.FrameTypeEvent, event)

    case ChatEventChunk:
        // 转换为 WS chunk 帧
        ws.SendFrame(protocol.FrameTypeEvent, event)

    case AgentEventBlockReply:
        // 块状回复（非流式 channel 或块状汇总）
        ws.SendFrame(protocol.FrameTypeEvent, event)

    case AgentEventCompleted:
        // 运行完成，最终答案
        // 网关 dedup 检查：是否与之前的 ChatEventChunk 重复
        if !isDuplicate(event.FinalContent) {
            ws.SendFrame(protocol.FrameTypeEvent, event)
        }
    }
}
```

**关键点：**
- Dedup 检查防止最终答案与 token 流重复
- 块状回复用于非流式 channel 补偿
- 所有事件都使用同一事件总线

---

## 事件去重（Dedup）

网关有内置去重机制（protocol/events.go 推测）：

```
场景 1: 流式输出
  ThinkStage (ChatEventChunk): "我现在要..."  ← 发出
  ObserveStage (BlockReplies): "我现在要..." ← 重复，被去重掉
  FinalizeStage (FinalContent): "我现在要..." ← 重复，被去重掉

场景 2: 非流式输出
  ThinkStage (no chunk events)
  ObserveStage (BlockReplies): "我现在要..."  ← 发出
  FinalizeStage (FinalContent): "完整答案"  ← 新内容，发出
```

---

## 总结

✅ **流式输出通过两个独立的机制触发：**

1. **Token 级别流式**（`makeCallLLM`）
   - 位置：`loop_pipeline_callbacks.go:265-283` 或 `:289-306`
   - 时机：LLM 流式调用期间（每个 token）
   - 事件：`ChatEventChunk` / `ChatEventThinking`

2. **块状回复**（`ThinkStage`）
   - 位置：`think_stage.go:150-151`
   - 时机：工具迭代完成，有中间文本
   - 事件：`AgentEventBlockReply`

**关键点：**
- ✅ `think_stage.go:150-151` **确实** 触发流式输出（块状回复）
- ✅ 但它是"块状"而非"token 级"流式
- ✅ Token 级实时流来自 `makeCallLLM` 的 provider callback
- ✅ 两层都通过 `emitRun()` 发送到事件总线，最终到网关 → 客户端
