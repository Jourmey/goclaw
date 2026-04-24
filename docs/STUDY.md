# GoClaw 7天精简学习清单

## 目标

7天后你应能做到：

- 说清项目主链路：请求进来 -> Agent处理 -> 模型调用 -> 返回
- 能新增一个简单接口或方法
- 能定位一次常见故障（配置/存储/路由）

## Day 1：启动链路（看懂程序怎么起来）

- 看：`main.go`、`cmd/root.go`、`cmd/gateway.go`
- 任务：
  - 画出启动流程图（函数调用级别）
  - 记录“启动必需配置”有哪些
- 产出：一页“启动流程笔记”

## Day 2：网关入口（请求怎么进来）

- 看：`internal/gateway/`、`internal/http/`、`pkg/protocol/`
- 任务：
  - 找到 WS 和 HTTP 的入口注册点
  - 选一个请求路径，跟踪到 handler
- 产出：一条完整请求链路（文件路径列表）

## Day 3：Agent 主流程（核心业务）

- 看：`internal/agent/`、`internal/pipeline/`
- 任务：
  - 梳理 pipeline 各阶段职责（context/history/prompt/think/act...）
  - 标记“输入/输出”在何处变化
- 产出：一张“阶段职责表”

## Day 4：模型与工具（能力来源）

- 看：`internal/providers/`、`internal/providerresolve/`、`internal/tools/`
- 任务：
  - 识别 provider 注册和调用入口
  - 找一个 tool 的注册 + 执行路径
- 产出：一条“模型调用链”+ 一条“工具调用链”

## Day 5：存储与配置（稳定运行基础）

- 看：`internal/config/`、`internal/store/`、`internal/store/pg/`、`migrations/`
- 任务：
  - 搞清配置加载优先级（文件/env）
  - 跟一条核心数据读写（如 session/agent）
- 产出：一页“配置&存储速查表”

## Day 6：做一个小改动（从看懂到能改）

- 只在核心层做小改动（别碰 UI/desktop/channels）：
  - 例如新增一个轻量 HTTP endpoint，或给现有响应加字段
- 任务：
  - 改代码 -> 本地跑通 -> 自测
- 产出：一个可运行的小功能

## Day 7：复盘与精简边界

- 复盘本周所有笔记
- 形成你自己的“当前不学清单”：
  - `ui/web/`
  - `ui/desktop/`
  - `internal/channels/`
  - `internal/memory/`、`internal/consolidation/`、`internal/knowledgegraph/`、`internal/vault/`
  - `internal/mcp/`、`internal/sandbox/`、`internal/tts/`
- 产出：你的“核心学习地图 v1”

## 每天固定节奏（建议）

- 30分钟：读代码（只读主链路）
- 20分钟：画图/写笔记
- 10分钟：跑一次最小验证（启动或接口调用）
