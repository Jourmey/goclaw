# Agent Loop 代码执行跟踪

这是一个典型的 Agent Run 从入口到完成的完整代码跟踪。

## 调用栈全景

```
HTTP POST /chat         (内部/gateway/methods/chat.go)
  ↓
gateway.ChatMethod.Run(ctx, msg SessionChatReq)
  ├─ 解析 sessionKey → 获取 Session
  ├─ 获取 Agent
  └─ Agent.Run(ctx, input)           ← 核心入口
       ↓
    Loop.Run(ctx, RunRequest)         (internal/agent/loop_run.go)
       └─ return l.runViaPipeline()   ← 第一行代码
           ↓
        runViaPipeline(ctx, req)      (internal/agent/loop_pipeline_adapter.go:16)
           ├─ convertRunInput(&req)
           ├─ bridgeRS := &runState{}
           ├─ l.buildPipelineDeps()   ← 依赖注入
           ├─ pipeline.NewDefaultPipeline(deps)
           └─ return p.Run(ctx, state)
               ↓
            Pipeline.Run(ctx, state)  (internal/pipeline/pipeline.go:52)
```

## 详细代码跟踪

### 1. 入口：`internal/agent/loop_run.go`

```go
// Loop.Run() 是外部调用入口
func (l *Loop) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
    return l.runViaPipeline(ctx, req)  // 直接委托
}
```

### 2. 适配器：`internal/agent/loop_pipeline_adapter.go:16-40`

```go
func (l *Loop) runViaPipeline(ctx context.Context, req RunRequest) (*RunResult, error) {
    // 1️⃣ 转换输入格式
    input := convertRunInput(&req)
    
    // 2️⃣ 创建状态桥接器（在 agent loop 和 pipeline 之间共享状态）
    bridgeRS := &runState{}
    
    // 3️⃣ 构建依赖注入（关键！所有 stage 回调都在这里注入）
    deps := l.buildPipelineDeps(&req, bridgeRS)
    
    // 4️⃣ 处理模型/提供商覆盖
    model := l.model
    if req.ModelOverride != "" {
        model = req.ModelOverride
    }
    provider := l.provider
    if req.ProviderOverride != nil {
        provider = req.ProviderOverride
    }
    
    // 5️⃣ 创建 Pipeline（标准 8-stage）
    p := pipeline.NewDefaultPipeline(deps)
    
    // 6️⃣ 创建 RunState（初始化）
    state := pipeline.NewRunState(input, nil, model, provider)
    
    // 7️⃣ 执行 Pipeline
    pResult, err := p.Run(ctx, state)
    if err != nil {
        return nil, err
    }
    
    // 8️⃣ 转换结果格式
    return convertRunResult(pResult), nil
}
```

### 3. Pipeline 初始化：`internal/pipeline/pipeline.go:32-50`

```go
func NewDefaultPipeline(deps PipelineDeps) *Pipeline {
    d := &deps
    memFlush := NewMemoryFlushStage(d)
    
    // 标准 8-stage：
    setup := []Stage{
        NewContextStage(d),  // 准备阶段
    }
    
    iteration := []Stage{
        NewThinkStage(d),       // LLM 调用
        NewPruneStage(d, memFlush),     // 消息剪枝
        NewToolStage(d),        // 工具执行
        NewObserveStage(d),     // 结果收集
        NewCheckpointStage(d),  // 状态保存
    }
    
    finalize := []Stage{
        NewFinalizeStage(d),    // 清理返回
    }
    
    return NewPipeline(setup, iteration, finalize, deps)
}
```

### 4. Pipeline 主循环：`internal/pipeline/pipeline.go:52-130`

```go
func (p *Pipeline) Run(ctx context.Context, state *RunState) (*RunResult, error) {
    start := time.Now()
    
    // =========== SETUP PHASE ===========
    for _, stage := range p.setup {
        if err := stage.Execute(ctx, state); err != nil {
            return nil, fmt.Errorf("setup %s: %w", stage.Name(), err)
        }
    }
    
    // 升级 context（contextStage 可能注入额外的 context 值）
    if state.Ctx != nil {
        ctx = state.Ctx
    }
    
    // =========== ITERATION LOOP ===========
    for state.Iteration = 0; state.Iteration < p.Deps.Config.MaxIterations; state.Iteration++ {
        // 执行迭代中的所有 stage
        for _, stage := range p.iteration {
            // 执行 Stage
            if err := stage.Execute(ctx, state); err != nil {
                return nil, fmt.Errorf("iter %d %s: %w", state.Iteration, stage.Name(), err)
            }
            
            // 检查 AbortRun（立即退出所有 stage）
            if swr, ok := stage.(StageWithResult); ok && swr.Result() == AbortRun {
                state.ExitCode = AbortRun
                break  // 退出 stage 循环
            }
        }
        
        // 检查全局退出
        if state.ExitCode == AbortRun {
            break  // 退出迭代循环
        }
        
        // 检查 BreakLoop（有序完成）
        for _, stage := range p.iteration {
            if swr, ok := stage.(StageWithResult); ok && swr.Result() == BreakLoop {
                state.ExitCode = BreakLoop
                break
            }
        }
        if state.ExitCode == BreakLoop {
            break  // 退出迭代循环
        }
        
        // 检查 context 取消
        if ctx.Err() != nil {
            state.ExitCode = AbortRun
            break
        }
    }
    
    // =========== FINALIZE PHASE ===========
    for _, stage := range p.finalize {
        if err := stage.Execute(ctx, state); err != nil {
            return nil, fmt.Errorf("finalize %s: %w", stage.Name(), err)
        }
    }
    
    // =========== 返回结果 ===========
    return state.ToResult(), nil
}
```

### 5. ThinkStage 详细流程：`internal/pipeline/think_stage.go:29-150`

```go
func (s *ThinkStage) Execute(ctx context.Context, state *RunState) error {
    s.result = Continue
    
    // --------- Step 1: 预算进度条 ---------
    s.maybeInjectNudge(state)
    // → 在 70% 和 90% 迭代时向消息注入预算提醒
    
    // --------- Step 2: 构建工具列表 ---------
    var toolDefs []providers.ToolDefinition
    if s.deps.BuildFilteredTools != nil {
        var err error
        toolDefs, err = s.deps.BuildFilteredTools(state)
        if err != nil {
            return fmt.Errorf("build tools: %w", err)
        }
    }
    // toolDefs: [{Name: "write_file", Args: {...}}, ...]
    
    // --------- Step 3: 构建 LLM 请求 ---------
    req := providers.ChatRequest{
        Messages: state.Messages.All(),  // 系统 + 历史 + 待处理
        Tools:    toolDefs,              // 工具定义
        Model:    state.Model,           // 模型名
        Options: map[string]any{
            providers.OptMaxTokens: s.deps.Config.MaxTokens,
        },
    }
    
    // --------- Step 4: 调用 LLM ---------
    resp, err := s.deps.CallLLM(ctx, state, req)
    // resp 结构：
    //   Content: "我先读取文件..."
    //   ToolCalls: [
    //     {Name: "read_file", Args: {"path": "/etc/hosts"}},
    //     {Name: "write_file", Args: {"path": "/tmp/out.txt", "content": "..."}},
    //   ]
    //   Thinking: "我需要..."
    
    // --------- Step 5: 处理错误和重试 ---------
    if err != nil {
        // 错误 A: Context Overflow（输入太长）
        if isContextOverflowErr(err) {
            if state.Think.OverflowRetries > 0 {
                return fmt.Errorf("context overflow after compaction: %w", err)
            }
            state.Think.OverflowRetries++
            // 紧急压缩：用 LLM 压缩历史
            if s.deps.CompactMessages != nil {
                originalLen := len(state.Messages.History())
                compacted, compactErr := s.deps.CompactMessages(
                    ctx, state.Messages.History(), state.Model)
                if compactErr == nil {
                    state.Messages.ReplaceHistory(compacted)
                    slog.Info("emergency_compaction_triggered",
                        "run_id", state.RunID,
                        "original_msgs", originalLen,
                        "compacted_msgs", len(compacted))
                    return nil  // 继续（重试本迭代）
                }
            }
        }
        
        // 错误 B: 截断错误（工具定义太长）
        if isTruncationErr(err) && state.Think.TruncRetries < maxTruncRetries {
            state.Think.TruncRetries++
            // 移除最长工具定义，重试
            return nil  // 继续（重试本迭代）
        }
        
        return fmt.Errorf("llm call failed: %w", err)
    }
    
    // --------- Step 6: 设置流程控制 ---------
    if len(resp.ToolCalls) == 0 {
        // 最终答案（无工具调用）
        s.result = BreakLoop  // ← 关键信号：完成迭代
    }
    
    // --------- Step 7: 保存响应到状态 ---------
    state.Think.LastResponse = resp
    
    // --------- Step 8: 发出块状回复（如有内容） ---------
    if resp.Content != "" {
        s.deps.EmitBlockReply(resp.Content)
    }
    
    return nil
}
```

**ThinkStage 流程图：**
```
Enter ThinkStage
  ↓
[步骤 1] 注入预算提示 (70%, 90%)
  ↓
[步骤 2] 获取工具列表
  ↓
[步骤 3] 构建 ChatRequest
  ↓
[步骤 4] 调用 LLM
  ├─ 成功 → 保存响应
  │  ├─ 有工具? → Continue（下一个 stage）
  │  └─ 无工具? → BreakLoop（完成迭代）
  │
  └─ 失败
     ├─ Context Overflow? → 紧急压缩 → 重试本迭代
     ├─ 截断错误? → 移除最长工具 → 重试本迭代
     └─ 其他? → 返回错误（中止 Run）
```

### 6. ToolStage 详细流程：`internal/pipeline/tool_stage.go:26-150`

```go
func (s *ToolStage) Execute(ctx context.Context, state *RunState) error {
    s.result = Continue
    
    // --------- 检查是否有工具调用 ---------
    resp := state.Think.LastResponse
    if resp == nil || len(resp.ToolCalls) == 0 {
        return nil  // ThinkStage 已设置 BreakLoop
    }
    
    toolCalls := resp.ToolCalls  // [{Name, Args}, ...]
    
    // --------- 选择执行路径 ---------
    // 并行路径（多工具优化）
    if len(toolCalls) > 1 && s.deps.ExecuteToolRaw != nil && s.deps.ProcessToolResult != nil {
        return s.executeParallel(ctx, state, toolCalls)
    }
    
    // 顺序路径（默认）
    for i, tc := range toolCalls {
        slog.Debug("executing_tool",
            "iteration", state.Iteration,
            "index", i,
            "tool", tc.Name,
            "args", tc.Args)
        
        // 执行工具
        msgs, err := s.deps.ExecuteToolCall(ctx, state, tc)
        if err != nil {
            return fmt.Errorf("execute tool %s: %w", tc.Name, err)
        }
        
        // 追加结果消息
        for _, msg := range msgs {
            state.Messages.AppendPending(msg)
        }
        
        state.Tool.TotalToolCalls++
        
        // 死循环检测（内部检查）
        if state.Tool.LoopKilled {
            s.result = BreakLoop  // ← 死循环：完成迭代
            slog.Warn("tool_loop.killed",
                "iteration", state.Iteration,
                "total_calls", state.Tool.TotalToolCalls)
            return nil
        }
    }
    
    // --------- 检查全局退出条件 ---------
    s.checkExitConditions(state)
    // → 检查预算、迭代限制等
    
    return nil
}
```

**并行执行子流程（`executeParallel`）：**
```go
func (s *ToolStage) executeParallel(ctx, state, toolCalls) error {
    // Phase 1: 并行 I/O（无状态）
    results := make([]rawResult, len(toolCalls))
    var wg sync.WaitGroup
    for i, tc := range toolCalls {
        wg.Add(1)
        go func(idx int, toolCall) {
            defer wg.Done()
            rawData, err := s.deps.ExecuteToolRaw(ctx, toolCall)
            results[idx] = rawResult{
                tc: toolCall,
                rawData: rawData,
                err: err,
            }
        }(i, tc)
    }
    wg.Wait()  // 等待所有工具完成
    
    // Phase 2: 顺序结果处理（有状态）
    for i, result := range results {
        if result.err != nil {
            return fmt.Errorf("tool %s: %w", result.tc.Name, result.err)
        }
        
        // 顺序处理（保持结果顺序）
        msg, err := s.deps.ProcessToolResult(
            ctx, state, result.rawData, result.tc)
        if err != nil {
            return fmt.Errorf("process result %d: %w", i, err)
        }
        
        state.Messages.AppendPending(msg)
        state.Tool.TotalToolCalls++
    }
    
    s.checkExitConditions(state)
    return nil
}
```

**ToolStage 执行时间线：**
```
迭代 #0:
  ├─ ThinkStage: 调用 LLM → 返回 [read_file, write_file]
  ├─ ToolStage:
  │  ├─ Phase 1: 并行执行
  │  │  ├─ [并发] read_file("file.txt")
  │  │  └─ [并发] write_file("output.txt")
  │  ├─ 等待两个任务完成
  │  └─ Phase 2: 顺序处理结果
  │     ├─ 处理 read_file 结果 → AppendPending()
  │     └─ 处理 write_file 结果 → AppendPending()
  │
  └─ ObserveStage: 收集副作用

迭代 #1:
  ├─ ThinkStage: 调用 LLM → 返回 [] (无工具)
  ├─ ThinkStage 设置 BreakLoop
  └─ ... 迭代终止
```

### 7. ObserveStage 详细流程：`internal/pipeline/observe_stage.go:19-68`

```go
func (s *ObserveStage) Execute(_ context.Context, state *RunState) error {
    // --------- Step 1: 排干注入通道 ---------
    // 工具执行产生的副作用消息（如 subagent 结果）
    if s.deps.DrainInjectCh != nil {
        for _, msg := range s.deps.DrainInjectCh() {
            state.Messages.AppendPending(msg)
        }
    }
    
    resp := state.Think.LastResponse
    if resp == nil {
        return nil
    }
    
    // --------- Step 2: 追踪块状回复 ---------
    // 工具迭代期间的中间回复（有内容 + 有工具调用）
    if resp.Content != "" && len(resp.ToolCalls) > 0 {
        state.Observe.BlockReplies++
        state.Observe.LastBlockReply = resp.Content
        // → 这些回复会被流式发送给用户
    }
    
    // --------- Step 3: 最终答案积累 ---------
    // 无工具调用时（最终答案）
    if len(resp.ToolCalls) == 0 {
        state.Observe.FinalContent = resp.Content
        state.Observe.FinalThinking = resp.Thinking
        // → 这是要返回给用户的最终回复
    }
    
    // --------- Step 4: 图像收集 ---------
    // 助手生成的最终图像（可能跨多个迭代）
    if len(resp.Images) > 0 {
        for _, img := range resp.Images {
            if img.Partial {
                continue  // 跳过流式部分帧
            }
            state.Observe.AssistantImages = append(
                state.Observe.AssistantImages, img)
        }
        // 清空以防重复计数
        resp.Images = nil
    }
    
    return nil
}
```

### 8. FinalizeStage 详细流程：`internal/pipeline/finalize_stage.go`

```go
func (s *FinalizeStage) Execute(ctx context.Context, state *RunState) error {
    // --------- Step 1: 最后一个 Commit ---------
    // 将 pending 消息提交到 history
    state.Messages.Commit()
    
    // --------- Step 2: 保存消息到数据库 ---------
    if s.deps.SaveMessages != nil {
        if err := s.deps.SaveMessages(ctx, state.Messages.All()); err != nil {
            return fmt.Errorf("save messages: %w", err)
        }
    }
    
    // --------- Step 3: 清理资源 ---------
    // 关闭临时连接、释放内存等
    
    return nil
}
```

## 消息流动全景

```
初始状态：
  system: ["You are GoClaw..."]
  history: [
    {role: user, content: "执行任务 X"},
    {role: assistant, content: "我会..."},
    {role: tool, content: "结果 Y"},
  ]
  pending: []

迭代 #0:

  ContextStage: 无变化
  
  ThinkStage:
    ├─ BuildFilteredTools() → [tool1, tool2, ...]
    ├─ CallLLM(system + history + pending)
    └─ 返回：
         Content: "我先读取文件"
         ToolCalls: [{Name: read_file, Args: ...}]
    └─ EmitBlockReply("我先读取文件") → 流式发送给用户
  
  ToolStage:
    ├─ ExecuteToolCall(read_file)
    └─ AppendPending({role: tool, content: "文件内容..."})
  
  ObserveStage:
    └─ state.Observe.BlockReplies++ (追踪块状回复)

迭代 #1:

  ThinkStage:
    ├─ CallLLM(system + history + pending)
    │  （注意：消息现在包含上一迭代的工具调用结果）
    └─ 返回：
         Content: "好的，任务完成"
         ToolCalls: [] ← 空！
         
    └─ 设置 s.result = BreakLoop ← 完成迭代
  
  ObserveStage:
    └─ state.Observe.FinalContent = "好的，任务完成"

FinalizeStage:
  ├─ Messages.Commit() → pending 追加到 history
  └─ SaveMessages() → 保存到数据库

返回 RunResult:
  FinalContent: "好的，任务完成"
  ToolCallCount: 1
  IterationCount: 2
  Usage: {PromptTokens: 1234, CompletionTokens: 567}
```

## 关键决策点

### 1. ThinkStage: 何时设置 BreakLoop？

```go
if len(resp.ToolCalls) == 0 {
    s.result = BreakLoop  // ← 关键决策
}
```

**触发条件：** LLM 返回无工具调用（最终答案）
**效果：** Pipeline 停止迭代，进入 FinalizeStage

### 2. ToolStage: 何时设置 BreakLoop？

```go
if state.Tool.LoopKilled {
    s.result = BreakLoop  // ← 死循环检测
}

s.checkExitConditions(state)
// → 检查预算、迭代限制等
```

**触发条件：**
- 死循环检测（read-only streak 过长）
- Token 预算超限
- 迭代次数超限

**效果：** 有序停止，返回目前为止的答案

### 3. 任何 Stage: 何时返回 Error？

```go
return fmt.Errorf("...")  // ← 中止整个 Run
```

**触发条件：**
- LLM 调用真正失败（非截断/overflow）
- 工具执行异常
- 数据库 I/O 错误

**效果：** Pipeline 立即终止，返回错误给调用者

## 性能特性

### 并行执行收益

假设有 3 个工具，每个需要 100ms：

**顺序执行：** 100 + 100 + 100 = 300ms
**并行执行：** max(100, 100, 100) = 100ms（3 倍加速）

```
顺序：  [Tool1: ====]
        [Tool2:      ====]
        [Tool3:           ====]
        总时间: ============

并行：  [Tool1: ====]
        [Tool2: ====]  ← 同时运行
        [Tool3: ====]
        总时间: ====
```

### Token 预算管理

在 `think_stage.go` 中跟踪：
```go
Usage: {
    PromptTokens: 5000,          // 历史 + 工具定义
    CompletionTokens: 2000,      // LLM 生成
    CacheReadTokens: 1000,       // 命中缓存
    CacheCreationTokens: 0,      // 创建缓存
}

RemainingBudget = MaxTokens - (PromptTokens + CompletionTokens)
```

在 70% 和 90% 迭代时注入预算提醒。

## 错误恢复

### 1. Context Overflow（输入过长）

```
LLM 返回 "context length exceeded"
  ↓
检测 isContextOverflowErr(err)
  ↓
调用 CompactMessages() 用 LLM 压缩历史
  ↓
替换 state.Messages.History()
  ↓
return nil (重试本迭代)
  ↓
重新调用 LLM（消息更短）
```

### 2. 截断错误（工具定义过长）

```
LLM 返回 "tool definitions too long"
  ↓
检测 isTruncationErr(err) && TruncRetries < 3
  ↓
移除最长的工具定义
  ↓
return nil (重试本迭代)
  ↓
重新调用 LLM（工具更少）
```

### 3. 死循环（同工具+参数+结果）

```
ToolStage.checkExitConditions()
  ↓
toolLoopState.recordResult()
  ↓
检测：同工具 × N 次 + 同参数 + 同结果
  ↓
如果 N >= 5: state.Tool.LoopKilled = true
  ↓
ToolStage 返回 BreakLoop
  ↓
有序停止迭代（不是崩溃）
```

