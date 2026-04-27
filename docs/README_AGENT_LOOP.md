# 🤖 Agent Loop 完整学习指南

本目录包含 GoClaw Agent Loop（思→行→观循环）的全面文档。

## 📚 文档导航

### 快速入门
| 文件 | 用途 | 推荐阅读时间 |
|------|------|-----------|
| **QUICK_REFERENCE.md** | ✨ 5 分钟快速上手 | 5 min |
| **AGENT_LOOP_GUIDE.md** | 📖 完整架构指南 | 20 min |
| **CODE_TRACE.md** | 🔍 代码执行跟踪 | 30 min |

### 学习路径

#### 🚀 新手路径
1. 先读 `QUICK_REFERENCE.md` 的"最重要的三个文件"
2. 了解"5 个关键概念"表
3. 跟踪"典型 Run 流程"图
4. 查看"三个流程控制信号"
5. 进阶：读 `AGENT_LOOP_GUIDE.md` 的"思→行→观循环细节"

#### 👨‍💻 开发者路径
1. 读 `AGENT_LOOP_GUIDE.md` 的"核心文件详解"
2. 研究 `CODE_TRACE.md` 的"详细代码跟踪"
3. 在 IDE 中打开对应代码，设置断点
4. 关注"关键决策点"和"错误恢复"

#### 🔧 维护者路径
1. 理解"8-Stage Pipeline 架构"
2. 熟悉"Pipeline Deps 回调"（扩展点）
3. 掌握"死循环检测"机制
4. 关注"Token 预算管理"
5. 研究"并行执行"的优化

## 🎯 核心概念一页纸

### Pipeline = 状态机

```
状态: RunState {
    Messages: [系统提示, 历史, 待处理]
    Think: {LastResponse, OverflowRetries}
    Tool: {TotalToolCalls, LoopKilled}
    Observe: {FinalContent, BlockReplies}
}

Stage 是函数: (ctx, state) → error
Stage 可以返回信号: Continue / BreakLoop / AbortRun

Pipeline 是循环:
    for i < MaxIterations {
        for each stage {
            stage.Execute(ctx, state)
            if stage.Result() == AbortRun { break }
        }
        if state.ExitCode == BreakLoop { break }
    }
```

### Think→Tool→Observe 循环

**Think Stage：** LLM 看到 messages + tools → 返回 {Content, ToolCalls}
- 无 ToolCalls → BreakLoop（最终答案）
- 有 ToolCalls → Continue（下一个 stage）

**Tool Stage：** 执行 ToolCalls（并行 I/O + 顺序结果）
- 检测死循环、预算 → 如果出问题 → BreakLoop
- 否则 → Continue

**Observe Stage：** 收集工具结果、积累最终答案
- 如果 ToolCalls 为空 → 保存最终答案
- 否则 → 继续迭代

### 死循环检测 = 三层防护

1. **Tool Loop** (3/5 阈值)：同工具 + 同参数 + 同结果
2. **Read-Only Streak** (8/12 阈值)：连续非变工具调用
3. **Same Result** (4/6 阈值)：不同参数但相同结果

在 `toolloop.go:75-150` 实现。

## 🗂️ 代码地图

### 关键文件速查

```
internal/
├── agent/
│   ├── loop_run.go                       ← 入口 Agent.Run()
│   ├── loop_pipeline_adapter.go:16       ← 适配器 runViaPipeline()
│   ├── loop_pipeline_callbacks.go        ← 所有 stage 回调实现
│   ├── toolloop.go                       ← 死循环检测
│   ├── resolver.go                       ← Agent 工厂
│   └── types.go                          ← Loop, RunRequest 类型
│
└── pipeline/
    ├── pipeline.go:52                    ← 主循环 Pipeline.Run()
    ├── stage.go                          ← Stage 接口 + 信号定义
    ├── think_stage.go:30                 ← LLM 调用 + 截断重试
    ├── tool_stage.go:27                  ← 工具执行 + 死循环检查
    ├── observe_stage.go:20               ← 结果收集
    ├── prune_stage.go                    ← 消息剪枝
    ├── context_stage.go                  ← 工作空间准备
    ├── checkpoint_stage.go               ← 状态保存
    ├── finalize_stage.go                 ← 清理返回
    ├── message_buffer.go                 ← 三层消息缓冲
    ├── run_state.go                      ← RunState 定义
    └── deps.go                           ← PipelineDeps 回调集合
```

### 一次典型运行的代码路径

```
HTTP /chat → gateway/methods/chat.go
           → Agent.Run(ctx, input)
           → Loop.runViaPipeline(ctx, req)           [adapter]
           → Pipeline.Run(ctx, state)                [engine]
           ├─ ContextStage.Execute()                 [setup]
           ├─ [Iteration Loop]
           │  ├─ ThinkStage.Execute()                [think]
           │  ├─ PruneStage.Execute()                [prune]
           │  ├─ ToolStage.Execute()                 [act]
           │  │  └─ ExecuteToolCall() × N             [tool callbacks]
           │  ├─ ObserveStage.Execute()              [observe]
           │  └─ CheckpointStage.Execute()           [checkpoint]
           └─ FinalizeStage.Execute()                [finalize]
           → return RunResult
           → [客户端接收最终答案]
```

## 💡 常见问题快速答案

### Q: Agent 如何决定何时停止？
**A:** `ThinkStage` 检查 LLM 返回的 `ToolCalls` 是否为空。
- 空 → 设置 `BreakLoop` → 迭代结束
- 非空 → 继续 → 执行工具

### Q: 如何处理 LLM 输入过长？
**A:** `ThinkStage` 的"Context Overflow"错误恢复（第 62-79 行）：
- 调用 `CompactMessages()` 用 LLM 压缩历史
- 重试本迭代（`return nil` 不设置结果）
- 最多 1 次

### Q: 工具执行如何并行？
**A:** `ToolStage.executeParallel()` 做两阶段：
1. **Phase 1：** 所有工具并行执行（`ExecuteToolRaw()`）
2. **Phase 2：** 顺序处理结果（`ProcessToolResult()`）

### Q: 死循环如何触发警告？
**A:** `toolloop.go` 的 `recordResult()` 检查三种模式：
- 同工具 × 3 次 + 同参数 + 同结果 → 警告
- × 5 次 → 杀死迭代

### Q: 如何添加新的 Stage？
**A:** 三步：
1. 实现 `Stage` 接口（`Name()` + `Execute()`）
2. 在 `NewDefaultPipeline()` 中注册
3. 通过 `RunState` 读写状态

### Q: 消息如何组织？
**A:** `MessageBuffer` 三层：
- `system`：系统提示（固定）
- `history`：会话历史（在 PruneStage 修改）
- `pending`：本次迭代新消息（ThinkStage 后追加）

LLM 看到：`All() = system + history + pending`

### Q: 如何跟踪一次 Run 的所有日志？
**A:** 在日志中查询 `RunID`（唯一标识符）：
```bash
grep "run_id=abc123" goclaw.log
# 显示这次 run 的所有 stage 日志
```

## 🧪 测试路径

### 单元测试
```bash
go test -v ./internal/pipeline/...      # Stage 单元测试
go test -v ./internal/agent/toolloop*   # 死循环检测
```

### 集成测试
```bash
go test -v -tags integration ./tests/integration/
```

### 覆盖率
```bash
go test -cover ./internal/pipeline/
go test -cover ./internal/agent/
```

## 📊 性能调优

### 1. 消息压缩
在 `ThinkStage` 因 Context Overflow 时触发（自动）。
也可在 `PruneStage` 中主动调用 `CompactMessages()`。

### 2. 工具并行
默认 `ToolStage` 检查 `len(toolCalls) > 1` 自动启用。
需要实现 `ExecuteToolRaw` + `ProcessToolResult` 回调。

### 3. Token 预算
在 70% 和 90% 迭代时注入预算提醒（自动）。
可通过 `PipelineConfig.MaxTokens` 调整。

### 4. 死循环检测阈值
编辑 `toolloop.go` 顶部常数调整敏感度。

## 🔗 相关文档

- **`docs/25-database-migrations.md`** - 数据库 schema
- **`docs/26-migrations-checklist.md`** - 迁移操作指南
- **`CLAUDE.md`** - 项目约定
- **`internal/pipeline/README.md`** - Pipeline 模块 README（如有）

## 🎓 推荐学习资源

1. **理解状态机：** 读 `CODE_TRACE.md` 的"消息流动全景"
2. **掌握并行执行：** 研究 `tool_stage.go` 的 `executeParallel()`
3. **调试技巧：** 在 IDE 中设置断点于 `Pipeline.Run()` 的主循环
4. **死循环检测：** 追踪 `toolloop.go` 中三个阈值的触发条件
5. **Token 管理：** 理解 `ThinkStage` 的 Context Overflow 恢复

## 📝 笔记模板

使用这个模板记录你的学习笔记：

```markdown
## [时间] Agent Loop 学习笔记

### 今天学到的三个要点
1. 
2. 
3. 

### 代码片段（有趣的发现）
- 文件：
- 行号：
- 解释：

### 疑问（待解决）
- 
```

---

**最后的心法：**

> Agent Loop 不是一个 `for` 循环。  
> 它是一个**可插拔的 Stage 工厂**，  
> 通过**不可变的 RunState**  
> 和**可控的流程信号**（Continue/BreakLoop/AbortRun）  
> 驱动的状态机。

这使得：
- ✅ 每个 Stage 可独立测试
- ✅ 新增 Stage 无需修改 Pipeline 核心
- ✅ 调试时消息清晰可追踪
- ✅ 并发执行安全（Stage 顺序执行）

---

**版本：** v1.0  
**最后更新：** 2026-04-27  
**维护者：** Claude  
**反馈或问题：** 查看 `QUICK_REFERENCE.md` 中的"代码导航速查"
