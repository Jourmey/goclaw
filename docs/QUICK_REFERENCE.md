# Agent Loop 快速参考

## 🎯 最重要的三个文件

### 1️⃣ Pipeline 主循环
**`internal/pipeline/pipeline.go:52`** - `Run()` 方法

```
for 迭代次数 = 0 to MaxIterations {
    for each stage in [think, prune, tool, observe, checkpoint] {
        stage.Execute()
        if AbortRun: 中止所有
        if BreakLoop: 完成迭代
    }
    if AbortRun or BreakLoop: 退出迭代
}
```

### 2️⃣ 入口适配器
**`internal/agent/loop_pipeline_adapter.go:16-40`** - `runViaPipeline()`

所有 Agent 从这里进入 v3 Pipeline。构建依赖注入 → 创建 Pipeline → 运行。

### 3️⃣ Think→Tool→Observe 核心
- **Think** (`think_stage.go:30-100`): 调用 LLM，获取工具调用
- **Tool** (`tool_stage.go:27-66`): 执行工具，检查死循环
- **Observe** (`observe_stage.go:20-68`): 收集结果，积累最终答案

## 🔑 5 个关键概念

| 概念 | 文件 | 说明 |
|------|------|------|
| **Pipeline** | `pipeline/pipeline.go` | 8-stage 引擎 |
| **RunState** | `pipeline/run_state.go` | 单次 run 的可变状态 |
| **Stage** | `pipeline/stage.go` | 每个 stage 的接口 + 流程控制信号 |
| **MessageBuffer** | `pipeline/message_buffer.go` | 三层消息缓冲（system/history/pending） |
| **ToolLoopState** | `agent/toolloop.go` | 死循环检测 |

## 🔄 典型 Run 流程

```
Agent.Run(ctx, RunRequest)
  ↓
Loop.runViaPipeline(ctx, RunRequest)
  ↓
Pipeline.NewDefaultPipeline(deps)
  ↓
Pipeline.Run(ctx, state)
  ├─ Setup: ContextStage
  │  └─ 加载工作空间、上下文文件、记忆
  │
  ├─ Iteration Loop (最多 MaxIterations 次):
  │  ├─ ThinkStage
  │  │  └─ 调用 LLM(messages + tools)
  │  │     ├─ 如果有工具: 返回 ToolCalls
  │  │     └─ 如果无工具: 返回最终答案 → BreakLoop
  │  │
  │  ├─ PruneStage
  │  │  └─ 消息剪枝、紧凑化
  │  │
  │  ├─ ToolStage
  │  │  ├─ 执行工具（并行 I/O + 顺序结果）
  │  │  ├─ 检查死循环、预算
  │  │  └─ 如果满足退出条件: BreakLoop
  │  │
  │  ├─ ObserveStage
  │  │  └─ 收集工具副作用、积累最终答案
  │  │
  │  └─ CheckpointStage
  │     └─ 每 5 次迭代保存状态
  │
  └─ Finalize: FinalizeStage
     ├─ 保存最终消息
     └─ 清理资源
  ↓
返回 RunResult (FinalContent, ToolCallCount, IterationCount, Usage)
```

## 💾 关键数据结构

### RunRequest
```go
{
    SessionKey:     "sess_123",     // 会话 ID
    UserID:         "user_456",     // 用户 ID
    Input:          "用户输入文本",  // 最新用户消息
    ModelOverride:  "claude-3-5",   // 可选：覆盖模型
    MaxIterations:  30,             // 可选：迭代上限
}
```

### RunState（贯穿整个 pipeline）
```go
{
    RunID:       "run_123",
    UserID:      "user_456",
    AgentID:     "agent_789",
    Input:       RunInput,
    Model:       "claude-3-5",
    Provider:    LLMProvider,

    Messages:    MessageBuffer,    // 三层消息缓冲
    Think:       ThinkState,       // LastResponse, OverflowRetries
    Tool:        ToolState,        // TotalToolCalls, LoopKilled
    Observe:     ObserveState,     // FinalContent, BlockReplies

    Iteration:   0,                // 当前迭代次数
    ExitCode:    Continue,         // 流程控制：Continue/BreakLoop/AbortRun
}
```

### 消息结构
```go
type Message struct {
    Role:       "user" | "assistant" | "tool",
    Content:    string,             // 文本内容
    ToolCalls:  []ToolCall,         // 仅当 Role="assistant"
    ToolResult: string,             // 仅当 Role="tool"
}
```

## 🛑 三个流程控制信号

| 信号 | 含义 | 谁设置 | 效果 |
|------|------|--------|------|
| **Continue** | 继续下一个 stage | Pipeline（默认） | 正常流动 |
| **BreakLoop** | 完成迭代 | ThinkStage、ToolStage | 有序退出迭代循环 |
| **AbortRun** | 中止运行 | ToolStage、各 Stage | 立即终止所有 stage |

**示例：**
- ThinkStage 无工具调用 → `BreakLoop` → 迭代结束，返回最终答案
- ToolStage 死循环/预算超限 → `BreakLoop` → 停止迭代，返回目前为止的答案
- LLM 调用错误 → Error → 整个 Run 失败

## 🧠 死循环检测阈值

在 `agent/toolloop.go` 顶部常数定义：

```go
const (
    toolLoopWarning  = 3      // 相同工具+参数+结果 出现 3 次 → 警告
    toolLoopCritical = 5      // 出现 5 次 → 杀死

    readOnlyStreakWarning  = 8    // 连续 8 个只读工具 → 警告
    readOnlyStreakCritical = 12   // 连续 12 个 → 杀死（卡住模式）

    readOnlyExplorationWarning  = 24   // 探索模式：24 → 警告
    readOnlyExplorationCritical = 36   // 36 → 杀死

    sameResultWarning  = 4    // 不同参数但结果相同 → 警告
    sameResultCritical = 6    // 出现 6 次 → 杀死
)
```

## 🔌 Pipeline 回调注入点（关键扩展点）

在 `loop_pipeline_adapter.go:43` 中的 `buildPipelineDeps()` 构建：

```go
deps := pipeline.PipelineDeps{
    // ContextStage
    ResolveWorkspace:    func() { /* 返回工作目录 */ },
    LoadContextFiles:    func() { /* 加载 .goclawmd 等 */ },

    // ThinkStage
    BuildFilteredTools:  func() { /* 权限过滤工具列表 */ },
    CallLLM:             func() { /* 调用 LLM 提供商 */ },

    // ToolStage
    ExecuteToolCall:     func() { /* 执行单个工具 */ },
    ExecuteToolRaw:      func() { /* 并行执行（原始） */ },
    ProcessToolResult:   func() { /* 处理结果（顺序） */ },

    // 其他
    PruneMessages:       func() { /* 历史剪枝 */ },
    CompactMessages:     func() { /* LLM 压缩 */ },
    SaveMessages:        func() { /* 持久化 */ },
}
```

## 📊 工具执行模式

### 顺序执行（单工具或默认）
```
for each tool in toolCalls {
    result := ExecuteToolCall(tool)
    state.Append(result)
}
```

### 并行执行（多工具优化）
```
Phase 1: 并行 I/O
  for i in 0..N {
    go results[i] = ExecuteToolRaw(tools[i])
  }
  wait all results

Phase 2: 顺序处理
  for i in 0..N {
    processed := ProcessToolResult(results[i])
    state.Append(processed)  // 保持顺序
  }
```

## 🎬 调试技巧

### 查看 Pipeline 流程
在 `pipeline.go:52` 设置断点，跟踪每个 stage 的执行和结果信号。

### 查看消息构建
在 `message_buffer.go` 观察：
- `system` - 系统提示（固定）
- `history` - 会话历史（在 Prune stage 被修改）
- `pending` - 本次迭代的新消息

### 死循环诊断
在 `toolloop.go:recordResult()` 查看：
- `toolLoopMatches()` - 检测重复（同工具+参数+结果）
- `readOnlyStreakCheck()` - 检测只读卡住
- 日志：`slog.Warn("tool_loop.*")`

### Token 统计
在 `run_state.go` 查看 `Usage` 字段，`think_stage.go` 中累积：
- `CompletionTokens` - LLM 生成
- `PromptTokens` - 输入
- `CacheCreationTokens` / `CacheReadTokens` - 缓存命中

## 🚀 常见扩展点

1. **添加新的 Stage**
   - 实现 `Stage` 接口（Name() + Execute()）
   - 在 `NewDefaultPipeline()` 中注册
   - 例：内存保存 stage、分析 stage 等

2. **自定义工具过滤**
   - 修改 `BuildFilteredTools` 回调
   - 基于权限、预算等动态过滤

3. **自定义 LLM 调用**
   - 修改 `CallLLM` 回调
   - 支持本地模型、自定义格式等

4. **自定义消息压缩**
   - 修改 `CompactMessages` 回调
   - 或在 `PruneStage` 前自定义剪枝逻辑

## 📝 代码导航速查

| 需求 | 查看文件 |
|------|---------|
| 理解整体流程 | `pipeline/pipeline.go` |
| 添加 LLM 调用逻辑 | `pipeline/think_stage.go` |
| 修改工具执行 | `pipeline/tool_stage.go` |
| 调整死循环阈值 | `agent/toolloop.go` 顶部常数 |
| 改进消息剪枝 | `pipeline/prune_stage.go` + callbacks |
| 跟踪 Agent 生命周期 | `agent/loop_pipeline_adapter.go` |
| Token 统计和预算 | `pipeline/run_state.go` + `think_stage.go` |
| 消息持久化 | `pipeline/finalize_stage.go` |
