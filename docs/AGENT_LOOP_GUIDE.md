# GoClaw Agent Loop 核心代码指南

## 📍 核心位置

### 主入口
- **`internal/agent/resolver.go`** - Agent resolver 工厂，创建 Loop 实例
- **`internal/agent/loop_run.go`** - Agent Run 执行入口
- **`internal/agent/loop_pipeline_adapter.go:runViaPipeline()` (Line 16)** - 所有 Agent 都通过这里进入 v3 Pipeline

## 🔄 8-Stage Pipeline 架构

所有 Agent 运行都使用同一个 8 阶段管道。位置：`internal/pipeline/`

### Pipeline 流程图

```
┌─────────────────────────────────────────────────────────────┐
│ SETUP (执行一次)                                              │
├─────────────────────────────────────────────────────────────┤
│ ✓ ContextStage       - 准备工作空间、加载上下文文件、注入记忆  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ ITERATION LOOP (最多 MaxIterations 次)                        │
├─────────────────────────────────────────────────────────────┤
│ 1. ThinkStage        - 调用 LLM，获取 Tool Calls/Final Answer│
│    ├─ 构建工具定义列表                                        │
│    ├─ 调用 LLM（处理截断重试）                              │
│    └─ 检查是否有 Tool Calls → 无则 BreakLoop               │
│                            ↓                               │
│ 2. PruneStage        - 消息剪枝、历史紧凑、内存刷新          │
│                            ↓                               │
│ 3. ToolStage         - 执行 Tool Calls（并行 I/O + 顺序结果）│
│    ├─ 并行执行多个工具                                       │
│    ├─ 顺序处理结果（保持状态安全）                          │
│    └─ 检查死循环、预算、退出条件                            │
│                            ↓                               │
│ 4. ObserveStage      - 收集工具侧效应，积累最终内容          │
│                            ↓                               │
│ 5. CheckpointStage   - 定期保存状态（每5次迭代）           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ FINALIZE (执行一次)                                           │
├─────────────────────────────────────────────────────────────┤
│ ✓ FinalizeStage      - 保存消息、清理资源、返回最终结果       │
└─────────────────────────────────────────────────────────────┘
```

## 📂 核心文件详解

### Pipeline 核心（`internal/pipeline/`）

| 文件 | 行数 | 功能 |
|------|------|------|
| **pipeline.go** | ~120 | Pipeline 主循环引擎 |
| **stage.go** | ~35 | Stage 接口定义（Continue/BreakLoop/AbortRun） |
| **think_stage.go** | ~150+ | LLM 调用 + 截断重试 + 预算进度条 |
| **tool_stage.go** | ~150+ | 工具执行（并行 I/O + 顺序结果） + 死循环检查 |
| **observe_stage.go** | ~70 | 工具副作用收集 + 最终内容积累 |
| **context_stage.go** | - | 工作空间解析、上下文文件加载 |
| **prune_stage.go** | - | 消息剪枝、紧凑化、内存刷新 |
| **checkpoint_stage.go** | - | 定期状态保存 |
| **finalize_stage.go** | - | 最终处理、资源清理 |

### Agent 适配层（`internal/agent/`）

| 文件 | 功能 |
|------|------|
| **loop_pipeline_adapter.go** (Line 16-40) | `runViaPipeline()` 入口 + Deps 构建 |
| **loop_pipeline_callbacks.go** | Pipeline 的所有回调实现 |
| **toolloop.go** | 死循环检测（read-only streak、same-result） |
| **resolver.go** | Agent 工厂、依赖注入 |
| **types.go** | Loop、RunRequest、RunResult 类型定义 |

### 关键数据结构（`internal/agent/loop_types.go`）

```go
type Loop struct {
    id                 UUID              // Agent ID
    agentKey          string             // Agent key (logs/paths)
    model             string             // LLM model
    provider          providers.Provider // LLM provider
    maxIterations     int                // 迭代次数限制
    contextWindow     int                // 模型上下文窗口
    compactionCfg     *CompactionConfig  // 消息紧凑策略
    // ... 其他配置
}

type RunRequest struct {
    SessionKey       string             // 会话ID
    UserID           string             // 用户ID
    Input            string             // 用户输入
    ModelOverride    string             // 可选模型覆盖
    MaxIterations    int                // 可选迭代限制
}

type RunResult struct {
    FinalContent     string             // 最终回复
    ToolCallCount    int                // 工具调用总数
    IterationCount   int                // 迭代次数
    Usage            *TokenUsage        // Token 统计
}
```

## 🎯 思→行→观 循环细节

### Think Stage（`think_stage.go:30-120`）

```go
func (s *ThinkStage) Execute(ctx context.Context, state *RunState) error {
    // 1. 预算进度条提示（70%/90% 迭代时注入 nudge）
    s.maybeInjectNudge(state)

    // 2. 构建工具定义列表（权限过滤）
    toolDefs, err := s.deps.BuildFilteredTools(state)

    // 3. 构建 ChatRequest（消息历史 + 工具列表）
    req := providers.ChatRequest{
        Messages: state.Messages.All(),
        Tools:    toolDefs,
        Model:    state.Model,
        Options:  {MaxTokens: ...},
    }

    // 4. 调用 LLM（处理 Context Overflow 重试）
    resp, err := s.deps.CallLLM(ctx, state, req)
    if isContextOverflowErr(err) && state.Think.OverflowRetries == 0 {
        // 紧急压缩 → 重试迭代
        compacted := s.deps.CompactMessages(state.Messages.History())
        state.Messages.ReplaceHistory(compacted)
        return nil // Continue (retry)
    }

    // 5. 截断重试逻辑（最多 3 次）
    if isTruncationErr(err) && state.Think.TruncRetries < maxTruncRetries {
        // 截断最长工具定义 → 重试迭代
        return nil // Continue (retry)
    }

    // 6. 设置流程控制
    if len(resp.ToolCalls) == 0 {
        s.result = BreakLoop  // 无工具调用 → 最终答案
    }
}
```

**关键点：**
- LLM 返回格式：`{Content: "...", ToolCalls: [{Name, Args}, ...], Thinking: "..."}`
- 无工具调用 = 最终答案，设置 BreakLoop
- 截断重试：移除最长工具 → 继续迭代

### Tool Stage（`tool_stage.go:27-66`）

```go
func (s *ToolStage) Execute(ctx context.Context, state *RunState) error {
    // 1. 检查是否有工具调用
    if len(resp.ToolCalls) == 0 {
        return nil  // 由 ThinkStage 设置 BreakLoop
    }

    // 2. 选择执行路径
    if len(toolCalls) > 1 && s.deps.ExecuteToolRaw != nil {
        // 并行路径：I/O 并行 + 结果顺序处理
        return s.executeParallel(ctx, state, toolCalls)
    } else {
        // 顺序路径：逐个执行
        for _, tc := range toolCalls {
            msgs, err := s.deps.ExecuteToolCall(ctx, state, tc)
            state.Messages.AppendPending(msg)
            state.Tool.TotalToolCalls++

            if state.Tool.LoopKilled {  // 死循环检测
                s.result = BreakLoop
                return nil
            }
        }
    }

    // 3. 检查退出条件
    s.checkExitConditions(state)
    // ├─ 迭代预算用尽 → AbortRun
    // ├─ Read-only streak 过长 → inject warning / AbortRun
    // ├─ 工具调用次数超限 → AbortRun
    // └─ Token 预算超限 → AbortRun
}
```

**并行执行路径（`executeParallel`）：**
1. **Phase 1：并行 I/O** - 所有工具同时执行（`ExecuteToolRaw`）
2. **Phase 2：顺序结果处理** - 按调用顺序处理结果（`ProcessToolResult`）

这样保证结果序列化（避免状态竞争）同时最大化 I/O 并行度。

### Observe Stage（`observe_stage.go:20-68`）

```go
func (s *ObserveStage) Execute(_ context.Context, state *RunState) error {
    // 1. 排干注入通道（工具副作用消息）
    for _, msg := range s.deps.DrainInjectCh() {
        state.Messages.AppendPending(msg)
    }

    resp := state.Think.LastResponse

    // 2. 追踪块状回复（工具迭代期间的中间回复）
    if resp.Content != "" && len(resp.ToolCalls) > 0 {
        state.Observe.BlockReplies++
        state.Observe.LastBlockReply = resp.Content
    }

    // 3. 最终答案积累（无工具调用时）
    if len(resp.ToolCalls) == 0 {
        state.Observe.FinalContent = resp.Content
        state.Observe.FinalThinking = resp.Thinking
    }

    // 4. 助手生成的最终图像积累
    for _, img := range resp.Images {
        if !img.Partial {
            state.Observe.AssistantImages = append(...)
        }
    }
}
```

## 🔀 流程控制信号

Pipeline 用三个信号控制流程（`pipeline/stage.go:12-16`）：

```go
const (
    Continue   = 0  // 继续下一个 stage
    BreakLoop  = 1  // 退出迭代循环（正常完成）
    AbortRun   = 2  // 中止整个运行（错误/杀死）
)
```

**Pipeline 流程逻辑（`pipeline.go:67-99`）：**
```go
for state.Iteration = 0; state.Iteration < MaxIterations; state.Iteration++ {
    for _, stage := range iteration {
        if err := stage.Execute(ctx, state); err != nil {
            return nil, err  // 错误中止
        }
        if stage.Result() == AbortRun {
            state.ExitCode = AbortRun
            break  // 立即中止所有 stage
        }
    }

    if state.ExitCode == AbortRun {
        break  // 中止外层循环
    }

    // 检查 BreakLoop 信号
    if anyStage.Result() == BreakLoop {
        state.ExitCode = BreakLoop
        break  // 有序退出迭代
    }
}
```

## 🧠 死循环检测（`internal/agent/toolloop.go`）

### 检测策略

| 类型 | 阈值（Warning/Critical） | 说明 |
|------|--------------------------|------|
| **Tool Loop** | 3 / 5 | 同工具 + 同参数 + 同结果 |
| **Read-only Streak** | 8 / 12 (stuck) | 连续不变工具调用 |
| **Read-only Exploration** | 24 / 36 (探索) | 探索新文件不算被卡 |
| **Same Result** | 4 / 6 | 不同参数但相同结果 |

### 实现（`toolloop.go:75-150`）

```go
type toolLoopState struct {
    history        []toolCallRecord  // 最近 30 个调用
    readOnlyStreak int               // 连续不变调用数
    seenReadArgs   map[string]bool   // 追踪唯一性
}

func (s *toolLoopState) record(toolName string, args map[string]any) {
    h := hashToolCall(toolName, args)
    s.history = append(s.history, toolCallRecord{
        toolName: toolName,
        argsHash: h,
    })
}

func (s *toolLoopState) recordResult(resultHash string) {
    // 检测：同工具 + 同参 + 同结果
    if toolLoopMatches(s.history) {
        return ToolLoopWarning / ToolLoopCritical
    }
}
```

## 📊 消息缓冲模型（`internal/pipeline/message_buffer.go`）

```go
type MessageBuffer struct {
    // 三层缓冲：
    system      []Message   // 系统提示（固定）
    history     []Message   // 会话历史（剪枝/压缩）
    pending     []Message   // 待提交（当前迭代）
}

func (mb *MessageBuffer) All() []Message {
    // 构建最终消息列表（用于 LLM 调用）
    return append(append(system, history...), pending...)
}

func (mb *MessageBuffer) Commit() {
    // 将 pending 追加到 history，清空 pending
    history = append(history, pending...)
    pending = nil
}
```

## 🔌 Pipeline Deps 回调（`internal/pipeline/deps.go`）

Agent Loop 通过依赖注入提供这些回调：

```go
type PipelineDeps struct {
    // Context Stage
    ResolveWorkspace   func(*RunState) string
    LoadContextFiles   func(ctx context.Context, state *RunState) ([]Message, error)

    // Think Stage
    BuildFilteredTools func(*RunState) ([]ToolDefinition, error)
    CallLLM            func(ctx context.Context, state *RunState, req ChatRequest) (*Response, error)

    // Prune Stage
    PruneMessages      func(ctx context.Context, msgs []Message) ([]Message, error)
    CompactMessages    func(ctx context.Context, msgs []Message, model string) ([]Message, error)

    // Tool Stage
    ExecuteToolCall    func(ctx context.Context, state *RunState, tc ToolCall) ([]Message, error)
    ExecuteToolRaw     func(ctx context.Context, tc ToolCall) (any, error)  // 并行
    ProcessToolResult  func(ctx context.Context, state *RunState, rawData any, tc ToolCall) ([]Message, error)

    // Finalize Stage
    SaveMessages       func(ctx context.Context, msgs []Message) error
}
```

所有回调由 `Loop.buildPipelineDeps()` (loop_pipeline_adapter.go:43) 实现并传入。

## 🚀 执行入口示例

```go
// 1. 通过 Resolver 创建 Agent Loop
deps := &ResolverDeps{
    AgentStore: ...,
    ProviderStore: ...,
    Tools: ...,
    // 其他依赖...
}
resolver := agent.NewResolver(deps)
agentLoop, err := resolver.ResolveAgent(ctx, agentID)

// 2. 发起 Run
req := agent.RunRequest{
    SessionKey:    "sess_123",
    UserID:        "user_456",
    Input:         "请执行任务 X",
    MaxIterations: 30,
}

result, err := agentLoop.Run(ctx, req)
// → runViaPipeline()
// → Pipeline.Run() 执行 8 stage
// → 返回 RunResult
```

## 📌 关键源码行号速查

| 功能 | 位置 |
|------|------|
| Pipeline 主循环 | `internal/pipeline/pipeline.go:52-110` |
| Think Stage | `internal/pipeline/think_stage.go:30-100+` |
| Tool Stage | `internal/pipeline/tool_stage.go:27-66` |
| Observe Stage | `internal/pipeline/observe_stage.go:20-68` |
| 并行执行 | `internal/pipeline/tool_stage.go:68-140` |
| Pipeline 适配 | `internal/agent/loop_pipeline_adapter.go:16-40` |
| 死循环检测 | `internal/agent/toolloop.go:75-150` |
| 消息缓冲 | `internal/pipeline/message_buffer.go` |

---

**最后一个理解的关键点：**

Agent Loop 从来不是 `goto loop`。它是通过一个**通用 Pipeline 引擎**驱动的，每个阶段都是**可插拔的 Stage**，所有状态都存在 **RunState** 中。这使得：
- ✅ 易于测试（每个 stage 独立）
- ✅ 易于扩展（添加新 stage）
- ✅ 易于诊断（每个 stage 日志清晰）
- ✅ 并发安全（Pipeline 序列化 stage 执行）
