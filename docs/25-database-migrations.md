# 数据库迁移系统 (Database Migrations)

## 概述

GoClaw 使用 **golang-migrate** 管理数据库 schema 版本。每个迁移包含两个文件：
- `.up.sql` — 应用迁移（向前）
- `.down.sql` — 回滚迁移（向后）

当前 **Required Schema Version: 55**（见 `internal/upgrade/version.go`）

> **双数据库支持：**
> - **PostgreSQL**：迁移文件在 `migrations/` （标准部署）
> - **SQLite**：全 schema 在 `internal/store/sqlitestore/schema.sql` + 增量补丁在 `schema.go` migrations map（桌面版）
> - 修改 schema 时**两个系统都要更新**，否则桌面版启动会崩溃

---

## 迁移分类与详解

### Phase 1: 核心基础 (000001-000005)

#### `000001_init_schema` — 初始化完整 Schema
**作用：** 项目启动时创建所有基础表结构

**关键表：**
- `llm_providers` — LLM 提供商配置（OpenAI、Anthropic、DashScope 等）
- `agents` — Agent 定义（所有者、模型、workspace、配置）
- `agent_shares` — Agent 共享权限
- `sessions` — 对话会话
- `custom_tools` — 用户自定义工具
- `messages` — 对话消息历史
- `spans` — 执行 trace 记录
- `skills` — 技能定义
- `channel_config` — 频道配置（Telegram、Feishu 等）

**扩展：**
```sql
CREATE EXTENSION "pgcrypto";  -- UUID 生成、加密
CREATE EXTENSION "vector";    -- pgvector 向量存储（内存系统）
```

**UUID v7 生成函数：** 基于时间戳 + 随机字节，确保 Agent 日志中的 KV 排序

---

#### `000005_phase4` — Agent 切换路由
**作用：** 支持多 Agent 切换（handoff）场景

**新表：**
```sql
handoff_routes (
    channel VARCHAR(50),
    chat_id VARCHAR(255),
    from_agent_key,
    to_agent_key,
    reason TEXT
)
```
用途：记录 Agent A → Agent B 的切换历史（Telegram、Feishu 等频道）

---

### Phase 2: 工具与技能系统 (000006-000017)

#### `000006_builtin_tools` — 内置工具注册表
**作用：** 管理系统内置工具的元数据和启用/禁用

**表：**
```sql
builtin_tools (
    name VARCHAR(100) PRIMARY KEY,  -- "web_search", "code_exec"
    display_name, category, enabled,
    settings JSONB                  -- 工具配置
)
```

**用途：**
- 全局启用/禁用工具（如禁用代码执行）
- 工具发现与分类
- 按需加载工具定义

---

#### `000010_agents_md_v2` — Agent Markdown 定义
**作用：** 支持 Agent 自定义 Markdown 上下文（`.md` 文件存储）

**新列：** `agents.agent_md_hash` — 用于检测 Agent 定义更新

---

#### `000015_agent_budget` — Agent 使用预算
**作用：** 限制单个 Agent 的 LLM 调用成本

**新表：**
```sql
agent_budgets (
    agent_id, period, max_cost, reset_at
)
```

**场景：** 防止成本失控（如过度循环调用）

---

#### `000017_system_skills` — 系统技能库
**作用：** 注册系统级技能（与 Agent 独立的技能定义）

**用途：**
- 跨 Agent 共享技能
- 版本控制技能更新
- 权限管理（谁能使用/修改技能）

---

### Phase 3: 安全与认证 (000020, 000026, 000032, 000036)

#### `000020_secure_cli_and_api_keys` — CLI 凭证注入 + API 密钥
**作用：** 支持安全的 CLI 工具集成

**新表：**
```sql
secure_cli_binaries (
    binary_name "gh", "gcloud" etc,
    binary_path,
    encrypted_env BYTEA,        -- AES-256-GCM 加密 JSON
    deny_args JSONB,            -- 黑名单参数（如 "auth\s+", "ssh-key"）
    timeout_seconds
)
```

**流程：**
1. 管理员配置 CLI → 加密 env vars（如 `GH_TOKEN`）
2. Agent 调用 CLI 时自动注入凭证
3. 执行时强制黑名单参数（防止信息泄露）

**API Keys 表：**
```sql
api_keys (
    key_hash VARCHAR(64) UNIQUE,
    scopes TEXT[],              -- e.g. ['operator.admin', 'operator.read']
    expires_at, last_used_at
)
```

---

#### `000026_api_key_user_binding` — API 密钥 ↔ 用户绑定
**作用：** 关联 API 密钥与创建者，支持审计

**新列：** `api_keys.created_by` — 允许追溯谁创建了该密钥

---

#### `000032_secure_cli_user_credentials` — 用户级 CLI 凭证
**作用：** 支持每个用户的独立 CLI 配置

**新表：**
```sql
user_cli_credentials (
    user_id, binary_name,
    encrypted_credential BYTEA  -- AES-256-GCM
)
```

场景：不同用户用不同的 GitHub token 执行 `gh` 命令

---

#### `000036_secure_cli_agent_grants` — Agent ↔ CLI 授权
**作用：** 定义哪个 Agent 可以使用哪个 CLI 工具

**新表：**
```sql
agent_cli_grants (
    agent_id, binary_name, granted_by, granted_at
)
```

**权限层级：**
```
全局 CLI（所有 Agent）
  ↓
租户级 CLI（该租户 Agent）
  ↓
Agent 级 CLI（特定 Agent）
```

---

### Phase 4: 多租户基础 (000027, 000029)

#### `000027_tenant_foundation` — 租户系统初始化
**作用：** 实现真正的多租户隔离

**新表：**
```sql
tenants (
    id UUID PRIMARY KEY,        -- Master UUID: 0193a5b0-7000-7000-8000-000000000001
    name, slug, status
)

tenant_users (
    tenant_id, user_id, role    -- 租户内的用户角色 (admin/member/viewer)
)
```

**ALTER 操作：** 在 30+ 表中添加 `tenant_id` 列（默认 Master UUID）

**关键语义：**
- **Master Tenant** — 系统级资源容器（全局 hooks、全局 CLI）
- **业务租户** — 用户创建的隔离空间
- **NULL tenant_id** — 系统级跨租户资源（如 API keys）

---

#### `000029_system_configs` — 系统配置存储
**作用：** 持久化系统级配置（避免环境变量）

**新表：**
```sql
system_configs (
    key VARCHAR(255) PRIMARY KEY,  -- "feature_flags.extended_thinking"
    value JSONB,
    description TEXT,
    updated_at
)
```

**用途：**
- 特性开关（feature flags）
- 系统级设置同步
- 运行时热更新（无需重启）

---

### Phase 5: 性能优化 (000009, 000030, 000051)

#### `000009_add_quota_index` — 配额查询优化
**作用：** 加速用户配额检查

```sql
CREATE INDEX idx_usage_snapshots_by_agent_time
  ON usage_snapshots(agent_id, bucket_hour DESC)
```

**查询优化：** `SELECT SUM(cost) FROM usage_snapshots WHERE agent_id=? AND bucket_hour > ?`
- 无索引：全表扫描
- 有索引：10x 快速

---

#### `000016_usage_snapshots` — 使用统计快照
**作用：** 分时段聚合 LLM 使用数据（用于计费、限流）

**表：**
```sql
usage_snapshots (
    bucket_hour TIMESTAMPTZ,   -- 按小时分桶
    agent_id, provider, model, channel,
    input_tokens, output_tokens,
    cache_read_tokens, cache_create_tokens, thinking_tokens,
    total_cost,
    request_count, llm_call_count, tool_call_count, error_count
)
```

**聚合流程：**
```
traces/spans 原始数据
  → 小时聚合 worker
  → usage_snapshots（预计算成本）
  → 仪表板显示
```

**好处：** 无需每次都重新计算成本，直接查聚合表 O(1)

---

#### `000030_jsonb_gin_indexes` — JSONB 全文搜索
**作用：** 快速搜索存储在 JSONB 字段中的数据

```sql
CREATE INDEX idx_agents_config_gin ON agents USING GIN(tools_config);
CREATE INDEX idx_spans_metadata_gin ON spans USING GIN(metadata);
```

**场景：** 搜索"所有启用了网络工具的 Agent"
- 无索引：`tools_config @> '{"web_search": true}'` 扫描所有行
- 有 GIN：倒排索引，毫秒级

---

#### `000051_backfill_context_pruning_mode` — 上下文修剪模式回填
**作用：** 迁移现有 Agent 的上下文修剪策略

**操作：**
```sql
UPDATE agents SET context_pruning = '{"mode": "adaptive"}'
WHERE context_pruning IS NULL
```

**语义：** 确保所有 Agent 有显式的修剪策略（无默认 NULL）

---

### Phase 6: 消息与频道 (000012, 000014, 000011)

#### `000012_channel_pending_messages` — 待发送消息队列
**作用：** 支持频道消息重试机制

**表：**
```sql
channel_pending_messages (
    id, channel_type, chat_id, user_id,
    content, status,            -- pending/sent/failed
    retry_count, next_retry_at
)
```

**场景：** Telegram 消息发送失败 → 自动重试 3 次

---

#### `000014_channel_contacts` — 频道联系人映射
**作用：** 管理 Telegram/Feishu/Discord 用户与 GoClaw 用户的映射

**表：**
```sql
channel_contacts (
    channel_type,  -- "telegram", "feishu"
    channel_user_id,
    goclaw_user_id,
    display_name, settings
)
```

**用途：** 支持跨频道用户识别

---

#### `000011_session_profile_metadata` — 会话元数据
**作用：** 存储每个会话的平台特定信息

**新列：** `sessions.profile_metadata JSONB`

**内容示例：**
```json
{
  "channel": "telegram",
  "chat_id": "123456",
  "thread_id": "789",  // Telegram topic thread
  "language": "vi"
}
```

---

### Phase 7: 观察与监控 (000044)

#### `000044_seed_agents_core_task_files` — 核心任务 Agent 文件
**作用：** 为系统 Agent 预装文件和 skill

**操作：** 种植（seed）`context_files` 表，为 Agent 关联预定义的任务文件

**用途：** 确保所有部署的核心 Agent 有一致的初始文件集

---

### Phase 8: Hooks 系统 (000052, 000053, 000054)

#### `000052_agent_hooks` — Agent hooks 核心系统
**作用：** 实现 Agent 行为中间件（tool 前/后拦截）

**表：**
```sql
agent_hooks (
    agent_id UUID REFERENCES agents(id),
    scope VARCHAR(8),           -- 'global' | 'tenant' | 'agent'
    event VARCHAR(32),          -- 'before_tool_call', 'after_tool_call'
    handler_type VARCHAR(16),   -- 'command' | 'http' | 'prompt'
    config JSONB,               -- 处理器特定配置
    matcher VARCHAR(256),       -- 工具名正则（如 "web_.*"）
    if_expr TEXT,               -- CEL 表达式，条件执行
    timeout_ms INT,
    on_timeout VARCHAR(8),      -- 'block' | 'allow'
    enabled BOOL
)
```

**示例：** 所有 API 调用前，验证请求签名
```sql
INSERT INTO agent_hooks VALUES (...)
  scope='global',
  event='before_tool_call',
  handler_type='http',
  matcher='api_.*',
  config='{"url": "https://verify.example.com/check"}'
```

---

#### `000053_agent_hooks_script_and_builtin` — Hooks 脚本模式
**作用：** 支持内联脚本和内置处理器

**新列：**
- `config.script` — 内联脚本（Python/Lua）
- `config.builtin_handler` — 预定义处理器名

---

#### `000054_agent_hooks_name` — Hooks 显示名
**作用：** 为 hooks 添加人类可读名称

**新列：** `agent_hooks.name VARCHAR(255)`

**好处：** UI 显示 "Rate Limit Check" 而不是 UUID

---

### Phase 9: Agent 扩展字段 (000055)

#### `000055_add_agent_extended_fields` — Agent 扩展字段
**作用：** 为 Agent 添加新的可选字段，无需迁移

**新列：**
```sql
ALTER TABLE agents ADD COLUMN extended_fields JSONB DEFAULT '{}'
```

**设计：** JSONB blob 支持向后兼容扩展
- 添加字段：直接写入 `extended_fields.new_field`
- 无需新迁移：避免 schema 膨胀

---

## 迁移执行流程

### 向上迁移（启动）
```bash
./goclaw migrate up
```

**流程：**
1. 检查当前 DB schema 版本（`schema_migrations` 表）
2. 比较 `RequiredSchemaVersion`（55）
3. 顺序执行缺失的 `.up.sql` 文件
4. 更新版本号 → 启动 gateway

**检查版本：**
```sql
SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;
```

---

### 向下迁移（回滚）
```bash
./goclaw migrate down 3  # 回滚最后 3 个迁移
```

**限制：** 某些迁移无法完全回滚（如数据删除）— 仅用于开发/测试

---

## SQLite 与 PostgreSQL 的同步

### PostgreSQL（生产）
- 迁移：`migrations/*.sql`
- 版本：`internal/upgrade/version.go` → `RequiredSchemaVersion`

### SQLite（桌面版）
- 完整 schema：`internal/store/sqlitestore/schema.sql`
- 增量补丁：`internal/store/sqlitestore/schema.go` → `migrations` map

**关键区别：**
- PG：顺序应用迁移
- SQLite：启动时检查版本 → 应用缺失补丁 → 更新 schema version

**修改 schema 时：**
```bash
# 1. 添加 PG 迁移
migrations/000056_new_feature.up.sql
migrations/000056_new_feature.down.sql

# 2. 更新 schema.go
var migrations = map[int]string{
    56: `ALTER TABLE ... ADD COLUMN ...`,
}

# 3. 更新版本
SchemaVersion = 56

# 4. 更新 PG 版本
RequiredSchemaVersion = 56
```

**验证：**
```bash
go build -tags sqliteonly ./...  # 确保 SQLite 构建成功
```

---

## 迁移命名约定

```
NNNNN_descriptive_name.{up,down}.sql
```

- **NNNNN**：序列号，不跳过（000001, 000002, 000003...）
- **descriptive_name**：snake_case，清晰表达意图

**好例子：**
- `000027_tenant_foundation` — 一眼看出是租户系统初始化
- `000016_usage_snapshots` — 新增使用统计快照表

**坏例子：**
- `000028_schema_updates` — 太模糊
- `000029_fix_bug_x` — 应该说是什么 bug

---

## 常见操作

### 添加新迁移
```bash
# 1. 生成新迁移文件
migrate create -ext sql -dir migrations -seq add_new_feature

# 2. 编写 .up.sql
cat > migrations/000056_add_new_feature.up.sql << 'EOF'
CREATE TABLE new_table (...);
EOF

# 3. 编写 .down.sql（回滚）
cat > migrations/000056_add_new_feature.down.sql << 'EOF'
DROP TABLE new_table;
EOF

# 4. 更新版本
sed -i 's/RequiredSchemaVersion uint = 55/RequiredSchemaVersion uint = 56/' internal/upgrade/version.go

# 5. SQLite 同步
# 更新 schema.go + schema.sql
```

### 测试迁移
```bash
# 清空测试 DB
dropdb goclaw_test 2>/dev/null || true
createdb goclaw_test

# 应用所有迁移
DATABASE_URL="postgres://localhost/goclaw_test" ./goclaw migrate up

# 回滚验证
DATABASE_URL="postgres://localhost/goclaw_test" ./goclaw migrate down 1
DATABASE_URL="postgres://localhost/goclaw_test" ./goclaw migrate up
```

---

## 常见问题

**Q: 为什么有 `.down.sql` 但生产环境不回滚？**
A: 生产迁移是单向的（向前兼容）。`.down.sql` 仅用于开发/测试环境快速重置。

**Q: 添加新列时为什么要 DEFAULT？**
A: 现有行必须有值。DEFAULT 避免 NULL 导致的业务逻辑错误。

**Q: SQLite 为什么用增量补丁而不是顺序迁移？**
A: SQLite 缺乏完整的 DDL 支持（如 `DROP COLUMN`）。完整 schema + 补丁组合提供灵活性。

**Q: JSONB 与固定列什么时候用？**
A: 频繁查询 → 固定列 + 索引；灵活扩展 → JSONB（如 `extended_fields`）

