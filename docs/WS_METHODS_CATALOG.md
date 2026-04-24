# WebSocket RPC 方法完整目录

**总计：127 个方法，10,585 行代码**

---

## ⚡ 重要说明：这些都是 WebSocket RPC 方法

### 📡 接口类型

| 类型 | 说明 | 格式 |
|------|------|------|
| **这份目录中的所有方法** | ✅ WebSocket RPC | `{"method": "chat.send", "params": {...}}` |
| **HTTP REST API** | ❌ 另外的系统 | `POST /v1/agents`, `GET /v1/sessions/{id}` |

### 🔄 两套独立系统

```
WebSocket RPC (这份文档)
├── 连接: ws://localhost:18790/ws
├── 方法: 127 个方法
├── 处理器: /internal/gateway/methods/
├── 路由: gateway.MethodRouter
└── 用途: 实时对话、Agent 管理（WS 客户端专用）

HTTP REST API (另一套系统)
├── 连接: http://localhost:18790/v1/*
├── 端点: /v1/agents, /v1/sessions, /v1/skills 等
├── 处理器: /internal/http/
├── 路由: http.ServeMux
└── 用途: 后端服务集成（标准 REST API）
```

### 💡 如何选择

**用 WebSocket RPC 当：**
- ✅ 构建 Web 前端（UI 实时通信）
- ✅ 构建桌面 App（Wails）
- ✅ 需要实时流式输出
- ✅ 需要保持连接状态

**用 HTTP REST API 当：**
- ✅ 构建后端服务
- ✅ 第三方集成
- ✅ 脚本自动化
- ✅ 无状态调用

---

## 📊 分类统计

| 优先级 | 数量 | 功能区域 | 代码行数 |
|--------|------|---------|---------|
| **Phase 1 - CRITICAL** | 29 | 核心对话、Agent 管理、配置、会话 | ~3,000 |
| **Phase 2 - NEEDED** | 35 | 技能、定时任务、团队、频道、配对 | ~4,000 |
| **Phase 2+ - ADVANCED** | 43 | 心跳、权限、工作区、API 密钥 | ~2,000 |
| **Phase 3 - NICE TO HAVE** | 20 | 浏览器、TTS、Hooks、日志、QR 码 | ~1,585 |

---

## 🔴 PHASE 1 - CRITICAL METHODS (29)

### 系统 (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `connect` | `MethodConnect` | WebSocket 握手，验证 token/API key，设置 locale/tenantId/role |
| `health` | `MethodHealth` | 健康检查，返回服务器状态、数据库连接、客户端列表 |
| `status` | `MethodStatus` | 获取服务器统计：Agent 数量、会话数、客户端数 |

### Agent (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `agent` | `MethodAgent` | 获取当前 Agent ID 和运行状态 |
| `agent.wait` | `MethodAgentWait` | 阻塞式等待 Agent 状态（未来特性） |
| `agent.identity.get` | `MethodAgentIdentityGet` | 获取 Agent 身份信息（名称、emoji、头像、描述）来自 IDENTITY.md |

### Chat (5) ⭐ **最重要**

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `chat.send` | `MethodChatSend` | 发送消息并执行 Agent 循环；支持流式输出、媒体、消息注入、自动中止 |
| `chat.history` | `MethodChatHistory` | 获取会话对话历史，返回带有签名的文件 URLs |
| `chat.abort` | `MethodChatAbort` | 中止正在运行的 Agent 调用（按 runId 或 sessionKey） |
| `chat.inject` | `MethodChatInject` | 注入消息到会话历史（不运行 Agent） |
| `chat.session.status` | `MethodChatSessionStatus` | 获取会话运行状态、活跃 runId、当前活动（phase/tool/iteration） |

### Agents Management (7)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `agents.list` | `MethodAgentsList` | 列出所有 Agent（管理员=全部，用户=仅可访问） |
| `agents.create` | `MethodAgentsCreate` | 创建新 Agent（指定 provider/model/identity） |
| `agents.update` | `MethodAgentsUpdate` | 修改 Agent 设置 |
| `agents.delete` | `MethodAgentsDelete` | 删除 Agent |
| `agents.files.list` | `MethodAgentsFileList` | 列出 Agent bootstrap 文件（SOUL.md、IDENTITY.md、USER.md 等） |
| `agents.files.get` | `MethodAgentsFileGet` | 获取 Agent 文件内容 |
| `agents.files.set` | `MethodAgentsFileSet` | 写入/更新 Agent 文件（持久化到 DB） |

### Config (5)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `config.get` | `MethodConfigGet` | 获取完整配置（隐藏密钥） |
| `config.apply` | `MethodConfigApply` | 替换整个配置（JSON5），提取密钥到 DB，广播变更事件 |
| `config.patch` | `MethodConfigPatch` | 合并部分配置更新（JSON5），持久化 |
| `config.schema` | `MethodConfigSchema` | 返回配置 JSON Schema（用于 UI 表单生成） |
| `config.defaults` | `MethodConfigDefaults` | 返回只读默认值 + Agent 默认值覆盖（无密钥） |

### Sessions (6)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `sessions.list` | `MethodSessionsList` | 分页会话列表（管理员=全部，用户=仅自己的） |
| `sessions.preview` | `MethodSessionsPreview` | 获取会话历史 + 摘要，带签名文件 URLs |
| `sessions.patch` | `MethodSessionsPatch` | 更新会话元数据（标签、模型、自定义字段） |
| `sessions.delete` | `MethodSessionsDelete` | 删除会话 |
| `sessions.reset` | `MethodSessionsReset` | 清空会话历史 |
| `sessions.compact` | `MethodSessionsCompact` | 截断历史到最后 N 条消息 |

---

## 🟠 PHASE 2 - NEEDED METHODS (35)

### Skills (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `skills.list` | `MethodSkillsList` | 列出所有技能（含状态/可见性/租户启用覆盖） |
| `skills.get` | `MethodSkillsGet` | 获取完整技能内容（按名称） |
| `skills.update` | `MethodSkillsUpdate` | 更新技能元数据（仅所有者或管理员） |

### Cron (8)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `cron.list` | `MethodCronList` | 列出定时任务（用户=仅自己的，管理员=全部），含禁用标志 |
| `cron.create` | `MethodCronCreate` | 创建新任务（schedule、message、delivery channel、stateless 标志） |
| `cron.update` | `MethodCronUpdate` | 修改任务字段 |
| `cron.delete` | `MethodCronDelete` | 删除任务 |
| `cron.toggle` | `MethodCronToggle` | 启用/禁用任务 |
| `cron.status` | `MethodCronStatus` | 获取全局 cron 调度器状态 |
| `cron.run` | `MethodCronRun` | 手动触发任务 |
| `cron.runs` | `MethodCronRuns` | 列出过去的任务执行历史 |

### Channels (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `channels.list` | `MethodChannelsList` | 获取启用的频道（Telegram、Discord、Slack、Zalo、WhatsApp 等） |
| `channels.status` | `MethodChannelsStatus` | 获取详细的频道连接状态 |
| `channels.toggle` | `MethodChannelsToggle` | 启用/禁用频道（未实现） |

### Device Pairing (6)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `device.pair.request` | `MethodPairingRequest` | 请求设备配对（sender_id/channel/chatId），返回审批码 |
| `device.pair.approve` | `MethodPairingApprove` | 批准待审批的配对 |
| `device.pair.deny` | `MethodPairingDeny` | 拒绝待审批的配对 |
| `device.pair.list` | `MethodPairingList` | 显示所有活跃配对 |
| `device.pair.revoke` | `MethodPairingRevoke` | 停用配对 |
| `browser.pairing.status` | `MethodBrowserPairingStatus` | 浏览器配对状态（未实现） |

### Exec Approval (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `exec.approval.list` | `MethodApprovalsList` | 待审批命令列表（ID/command/agentId/createdAt） |
| `exec.approval.approve` | `MethodApprovalsApprove` | 允许命令执行（一次性或始终） |
| `exec.approval.deny` | `MethodApprovalsDeny` | 拒绝命令 |

### Usage & Quota (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `usage.get` | `MethodUsageGet` | 分页 Token 使用记录（按会话/模型/provider） |
| `usage.summary` | `MethodUsageSummary` | 按 Agent 聚合 Token 使用量 |
| `quota.usage` | `MethodQuotaUsage` | 每用户配额消费量（未配置时返回禁用） |

### Generic Send (1)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `send` | `MethodSend` | 发布出站消息到 channel/chatId |

---

## 🟡 PHASE 2+ - ADVANCED METHODS (43)

### Heartbeat Monitoring (8)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `heartbeat.get` | `MethodHeartbeatGet` | 获取 Agent 心跳配置 |
| `heartbeat.set` | `MethodHeartbeatSet` | 写入 HEARTBEAT.md 配置 |
| `heartbeat.toggle` | `MethodHeartbeatToggle` | 启用/禁用心跳 |
| `heartbeat.test` | `MethodHeartbeatTest` | 立即触发心跳检查 |
| `heartbeat.logs` | `MethodHeartbeatLogs` | 获取心跳执行历史日志 |
| `heartbeat.checklist.get` | `MethodHeartbeatChecklistGet` | 获取预检清单 |
| `heartbeat.checklist.set` | `MethodHeartbeatChecklistSet` | 更新清单项 |
| `heartbeat.targets` | `MethodHeartbeatTargets` | 列出告警目标（teams/channels） |

### Config Permissions / RBAC (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `config.permissions.list` | `MethodConfigPermissionsList` | 列出 Agent/configType 的 RBAC 权限 |
| `config.permissions.grant` | `MethodConfigPermissionsGrant` | 分配权限给用户 |
| `config.permissions.revoke` | `MethodConfigPermissionsRevoke` | 撤销权限 |

### Channel Instances (5)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `channels.instances.list` | `MethodChannelInstancesList` | 列出所有频道实例（凭证隐藏） |
| `channels.instances.get` | `MethodChannelInstancesGet` | 获取实例详情（凭证隐藏） |
| `channels.instances.create` | `MethodChannelInstancesCreate` | 添加新频道实例 |
| `channels.instances.update` | `MethodChannelInstancesUpdate` | 修改实例 |
| `channels.instances.delete` | `MethodChannelInstancesDelete` | 删除实例 + 清缓存 |

### Agent Links / Delegation (4)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `agents.links.list` | `MethodAgentsLinksList` | 列出委托链接（from/to/all 方向） |
| `agents.links.create` | `MethodAgentsLinksCreate` | 添加委托链接 |
| `agents.links.update` | `MethodAgentsLinksUpdate` | 修改链接 |
| `agents.links.delete` | `MethodAgentsLinksDelete` | 删除链接 |

### API Key Management (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `api_keys.list` | `MethodAPIKeysList` | 列出 API 密钥（非管理员=仅自己的） |
| `api_keys.create` | `MethodAPIKeysCreate` | 生成新密钥（含 scopes + 可选过期时间） |
| `api_keys.revoke` | `MethodAPIKeysRevoke` | 停用 API 密钥 |

### Voices / TTS (2)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `voices.list` | `MethodVoicesList` | 获取 ElevenLabs 语音列表（按租户缓存） |
| `voices.refresh` | `MethodVoicesRefresh` | 清缓存并获取最新语音列表 |

### Team Core (6)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `teams.list` | `MethodTeamsList` | 列出所有 teams（管理员=全部，成员=仅自己的） |
| `teams.create` | `MethodTeamsCreate` | 创建新 team |
| `teams.get` | `MethodTeamsGet` | 获取 team + 成员（非管理员检查成员资格） |
| `teams.delete` | `MethodTeamsDelete` | 删除 team + 清 Agent 缓存 |
| `teams.update` | `MethodTeamsUpdate` | 更新 team 元数据 |
| `teams.known_users` | `MethodTeamsKnownUsers` | 列出 team 中参与过的用户 |

### Team Members (2)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `teams.members.add` | `MethodTeamsMembersAdd` | 添加用户/组到 team |
| `teams.members.remove` | `MethodTeamsMembersRemove` | 移除成员 |

### Team Tasks (13)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `teams.tasks.list` | `MethodTeamsTaskList` | 分页任务列表（状态/频道过滤） |
| `teams.tasks.get` | `MethodTeamsTaskGet` | 获取任务详情（含 comments/events/attachments + 签名 URLs） |
| `teams.tasks.get-light` | `MethodTeamsTaskGetLight` | 轻量级获取任务（无 comments/events） |
| `teams.tasks.approve` | `MethodTeamsTaskApprove` | 标记任务为已批准（含可选反馈） |
| `teams.tasks.reject` | `MethodTeamsTaskReject` | 标记任务为已拒绝 |
| `teams.tasks.comment` | `MethodTeamsTaskComment` | 添加任务注释 |
| `teams.tasks.comments` | `MethodTeamsTaskComments` | 列出任务注释 |
| `teams.tasks.events` | `MethodTeamsTaskEvents` | 列出任务状态变更事件 |
| `teams.tasks.create` | `MethodTeamsTaskCreate` | 创建新任务（从 Agent 通过 dispatch tracker 路由） |
| `teams.tasks.delete` | `MethodTeamsTaskDelete` | 删除单个任务 |
| `teams.tasks.delete-bulk` | `MethodTeamsTaskDeleteBulk` | 批量删除任务 |
| `teams.tasks.assign` | `MethodTeamsTaskAssign` | 重新分配任务给其他成员 |
| `teams.tasks.active-by-session` | `MethodTeamsTaskActiveBySession` | 获取聊天会话的活跃任务 |

### Team Workspace (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `teams.workspace.list` | `MethodTeamsWorkspaceList` | 列出 team 工作区文件 |
| `teams.workspace.read` | `MethodTeamsWorkspaceRead` | 读取 team 工作区文件（symlink 边界检查） |
| `teams.workspace.delete` | `MethodTeamsWorkspaceDelete` | 删除 team 工作区文件 |

### Team Events & Scopes (2)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `teams.events.list` | `MethodTeamsEventsList` | 列出 team 审计/活动事件 |
| `teams.scopes` | `MethodTeamsScopes` | 列出成员 scopes（team/chat） |

---

## 🟢 PHASE 3+ - NICE TO HAVE (20)

### Hooks / Webhooks (7)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `hooks.list` | `MethodHooksList` | 列出 Agent hooks（按 event/scope 过滤） |
| `hooks.create` | `MethodHooksCreate` | 创建 webhook hook |
| `hooks.update` | `MethodHooksUpdate` | 修改 hook |
| `hooks.delete` | `MethodHooksDelete` | 删除 hook |
| `hooks.toggle` | `MethodHooksToggle` | 启用/禁用 hook |
| `hooks.test` | `MethodHooksTest` | 干运行 hook（不写审计） |
| `hooks.history` | `MethodHooksHistory` | 列出 hook 执行历史 |

### Logs (1)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `logs.tail` | `MethodLogsTail` | 启动/停止实时日志跟踪（action=start/stop、level） |

### TTS / Text-to-Speech (6)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `tts.status` | `MethodTTSStatus` | TTS 状态（未实现） |
| `tts.enable` | `MethodTTSEnable` | 启用 TTS（未实现） |
| `tts.disable` | `MethodTTSDisable` | 禁用 TTS（未实现） |
| `tts.convert` | `MethodTTSConvert` | 文本转语音（未实现） |
| `tts.setProvider` | `MethodTTSSetProvider` | 设置 TTS provider（未实现） |
| `tts.providers` | `MethodTTSProviders` | 列出可用 TTS providers（未实现） |

### Browser Automation (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `browser.act` | `MethodBrowserAct` | 浏览器操作（未实现） |
| `browser.snapshot` | `MethodBrowserSnapshot` | 浏览器快照（未实现） |
| `browser.screenshot` | `MethodBrowserScreenshot` | 浏览器截图（未实现） |

### Platform QR / Integration (3)

| 方法名 | 常量 | 功能说明 |
|--------|------|---------|
| `zalo.personal.qr.start` | `MethodZaloPersonalQRStart` | Zalo 个人二维码扫描（未实现） |
| `zalo.personal.contacts` | `MethodZaloPersonalContacts` | Zalo 联系人列表（未实现） |
| `whatsapp.qr.start` | `MethodWhatsAppQRStart` | WhatsApp 二维码扫描（未实现） |

---

## 🔒 安全特性

### 通用处理器签名
```go
func (m *<Type>Methods) handle<Action>(
    ctx context.Context,
    client *gateway.Client,
    req *protocol.RequestFrame
)
```

### 安全防护

| 防护类型 | 实现方式 | 方法示例 |
|---------|---------|---------|
| **所有权检查** | `requireSessionOwner()`, `canSeeAll()` | chat.send, sessions.* |
| **角色检查** | `requireMinRole()` RBAC | config.apply, agents.delete |
| **输入验证** | JSON unmarshal + i18n 错误 | 所有方法 |
| **速率限制** | `m.rateLimiter.Allow(key)` | chat.send |
| **凭证隐藏** | `maskInstance()` | channels.instances.* |
| **路径遍历保护** | `resolveWorkspacePath()` symlink 检查 | teams.workspace.read |
| **密钥管理** | 提取到 `config_secrets` 表 | config.apply, config.patch |

### 上下文传播
```go
store.WithUserID(ctx)        // 用户隔离
store.WithTenantID(ctx)      // 租户隔离
store.WithLocale(ctx)        // 错误消息 i18n
store.WithAgentType(ctx)     // Agent 类型
```

### 缓存失效
- Team/Agent 变更 → 路由缓存清除
- Config 变更 → `bus.TopicConfigChanged` 事件广播
- Channel 实例 → `protocol.EventCacheInvalidate`

---

## 📋 使用场景矩阵

| 场景 | 必需方法 | 推荐方法 | 可选 |
|-----|---------|---------|------|
| **基础聊天** | connect, chat.send, chat.history | sessions.* | - |
| **Agent 管理** | agents.list/create/update/delete | agents.files.*, config.* | agent_links |
| **团队协作** | teams.list/create/get | teams.tasks.*, teams.members.* | teams.workspace |
| **自动化** | cron.* | heartbeat.* | hooks.* |
| **集成** | channels.list | channels.instances.* | zalo.*, whatsapp.* |
| **监控** | health, status | usage.*, quota.* | logs.tail |

---

## 💾 精简建议

### 保留必需（Phase 1）
✅ 所有 29 个 Phase 1 方法 - 核心功能

### 推荐保留（Phase 2）
✅ Chat、Sessions、Config、Agents - 基础功能
✅ Skills、Cron、Channels、Teams.core - 常用功能

### 可延迟（Phase 2+ 高级）
⚠️ Agent Links、Heartbeat、Config Permissions - 高级功能
⚠️ Team 任务/工作区 - 仅 20% 用户需要

### 建议移除或模块化（Phase 3）
❌ Browser 自动化 - 未实现，5% 用户
❌ TTS - 可选功能，6% 用户
❌ Zalo/WhatsApp QR - 平台特定，<5% 用户
❌ Hooks - 高级自动化，<10% 用户
❌ Logs.tail - 调试用途

**估计可减少：1,450-2,000 行代码（13-19%）**

---

## 🌐 WebSocket vs HTTP API 对照

### 常见功能的两种调用方式

| 功能 | WebSocket RPC | HTTP REST API |
|------|---------------|---------------|
| **发送消息** | `chat.send` | `POST /v1/agents/{id}/summon` |
| **获取历史** | `chat.history` | `GET /v1/sessions/{id}/messages` |
| **列出 Agent** | `agents.list` | `GET /v1/agents` |
| **创建 Agent** | `agents.create` | `POST /v1/agents` |
| **列出会话** | `sessions.list` | `GET /v1/sessions` |
| **删除会话** | `sessions.delete` | `DELETE /v1/sessions/{id}` |
| **列出技能** | `skills.list` | `GET /v1/skills` |
| **更新配置** | `config.apply` | `PUT /v1/config` |

### 协议对比

**WebSocket RPC 请求示例：**
```json
{
  "id": "123",
  "method": "chat.send",
  "params": {
    "sessionKey": "abc",
    "message": "Hello",
    "mediaItems": []
  }
}
```

**HTTP REST API 请求示例：**
```bash
curl -X POST http://localhost:18790/v1/agents/{id}/summon \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'
```

---

## 📁 实现位置

### WebSocket 实现（这份目录）
```
/Users/zhengjm/product/claw/goclaw/internal/gateway/methods/
├── chat.go (19KB)                      — 发送、历史、中止、注入、状态
├── agents.go + agents_*.go (30KB)      — Agent CRUD + 文件管理
├── sessions.go (10KB)                  — 会话 CRUD
├── config.go (11KB)                    — 配置管理
├── teams_*.go (65KB)                   — Team CRUD、任务、工作区
├── skills.go (5KB)                     — 技能管理
├── cron.go (10KB)                      — 定时任务
├── hooks.go (17KB)                     — Webhook 管理
├── heartbeat.go (15KB)                 — Agent 健康检查
└── access.go (2KB)                     — 权限检查工具
```

### HTTP REST API 实现（另一套系统）
```
/Users/zhengjm/product/claw/goclaw/internal/http/
├── agents.go (25KB)                    — Agent CRUD + 导入导出
├── sessions.go                         — 会话管理
├── skills.go + skills_upload.go        — 技能管理
├── teams_*.go                          — Team 相关
├── knowledge_graph_handlers.go         — 知识图谱
└── ... 其他特定功能
```

**共享依赖（两套系统都用）：**
```
/Users/zhengjm/product/claw/
├── internal/agent/                    — Agent 循环引擎
├── internal/store/                    — 数据库操作
├── internal/pipeline/                 — Agent 执行管道
├── internal/bus/                      — 事件总线
└── pkg/protocol/                      — 协议定义
```

---

## 🔌 连接方式

### WebSocket 连接（推荐用于前端）
```javascript
// JavaScript 客户端
const ws = new WebSocket('ws://localhost:18790/ws');

ws.send(JSON.stringify({
  id: '1',
  method: 'connect',
  params: {
    token: 'your-token',
    userId: 'user-123',
    locale: 'en'
  }
}));

ws.send(JSON.stringify({
  id: '2',
  method: 'chat.send',
  params: {
    sessionKey: 'session-abc',
    message: 'Hello'
  }
}));
```

### HTTP REST API（推荐用于后端）
```bash
# 设置 token
TOKEN="your-api-key"

# 创建 Agent
curl -X POST http://localhost:18790/v1/agents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Agent",
    "model": "claude-3-sonnet"
  }'

# 列出会话
curl -X GET http://localhost:18790/v1/sessions \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📚 使用场景

| 客户端类型 | 推荐协议 | 原因 |
|----------|--------|------|
| **Web 前端 (React/Vue)** | WebSocket | ✅ 实时更新、流式输出、双向通信 |
| **桌面 App (Wails)** | WebSocket | ✅ 实时状态管理、嵌入式连接 |
| **移动 App** | WebSocket | ✅ 实时消息、长连接 |
| **后端服务** | HTTP REST | ✅ 无状态、易于集成、标准接口 |
| **CLI 工具** | HTTP REST | ✅ 简单、可脚本化 |
| **第三方集成** | HTTP REST | ✅ 标准化、易于文档 |

---

## 📊 代码量统计

| 部分 | 文件数 | 代码行数 | 说明 |
|------|--------|--------|------|
| **WebSocket RPC** | 48 | ~10,585 | 这份目录 |
| **HTTP REST API** | 30+ | ~8,000 | 另一套系统 |
| **共享逻辑** | - | ~5,000 | Agent、Store、Pipeline |

---

**生成日期：2026-04-24**
**GoClaw 版本：v3.x**
