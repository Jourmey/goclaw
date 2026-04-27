# 迁移快速参考

## 所有迁移一览表

| 版本 | 名称 | 主要功能 | 关键表 | 状态 |
|------|------|--------|-------|------|
| **001** | `init_schema` | 完整初始化 schema | `agents`, `sessions`, `messages`, `spans`, `skills`, `llm_providers` | ✅ 核心 |
| **005** | `phase4` | Agent 切换路由 | `handoff_routes` | ✅ 通讯 |
| **006** | `builtin_tools` | 内置工具注册表 | `builtin_tools` | ✅ 工具 |
| **009** | `add_quota_index` | 配额查询优化 | 索引 on `usage_snapshots` | ✅ 性能 |
| **010** | `agents_md_v2` | Agent Markdown 支持 | `agents.agent_md_hash` | ✅ Agent |
| **011** | `session_profile_metadata` | 会话平台元数据 | `sessions.profile_metadata` | ✅ 通讯 |
| **012** | `channel_pending_messages` | 消息重试队列 | `channel_pending_messages` | ✅ 通讯 |
| **014** | `channel_contacts` | 频道用户映射 | `channel_contacts` | ✅ 通讯 |
| **015** | `agent_budget` | Agent 成本预算 | `agent_budgets` | ✅ 监管 |
| **016** | `usage_snapshots` | 使用统计聚合 | `usage_snapshots` | ✅ 计费 |
| **017** | `system_skills` | 系统技能库 | `system_skills`, `skill_versions` | ✅ 技能 |
| **020** | `secure_cli_and_api_keys` | CLI 凭证 + API 密钥 | `secure_cli_binaries`, `api_keys` | ✅ 安全 |
| **026** | `api_key_user_binding` | 密钥创建者追踪 | `api_keys.created_by` | ✅ 安全 |
| **027** | `tenant_foundation` | 多租户基础 | `tenants`, `tenant_users` | ✅ 核心 |
| **029** | `system_configs` | 系统配置存储 | `system_configs` | ✅ 配置 |
| **030** | `jsonb_gin_indexes` | JSONB 全文搜索索引 | GIN 索引 | ✅ 性能 |
| **032** | `secure_cli_user_credentials` | 用户级 CLI 凭证 | `user_cli_credentials` | ✅ 安全 |
| **036** | `secure_cli_agent_grants` | Agent ↔ CLI 授权 | `agent_cli_grants` | ✅ 安全 |
| **044** | `seed_agents_core_task_files` | 核心任务 Agent 文件 | 种植 `context_files` | ✅ Agent |
| **051** | `backfill_context_pruning_mode` | 上下文修剪回填 | `agents.context_pruning` | ✅ 优化 |
| **052** | `agent_hooks` | Hooks 核心系统 | `agent_hooks`, `hook_executions`, `tenant_hook_budget` | ✅ 扩展 |
| **053** | `agent_hooks_script_and_builtin` | Hooks 脚本模式 | `agent_hooks.config` 扩展 | ✅ 扩展 |
| **054** | `agent_hooks_name` | Hooks 显示名 | `agent_hooks.name` | ✅ 扩展 |
| **055** | `add_agent_extended_fields` | Agent 扩展字段 JSONB | `agents.extended_fields` | ✅ 扩展 |

---

## 按功能分类

### 核心系统（必需）
- `001` — 基础 schema
- `027` — 多租户基础
- `029` — 系统配置

### 通讯与频道
- `005` — Agent 切换
- `011` — 会话元数据
- `012` — 消息队列
- `014` — 频道联系人

### 安全与认证
- `020` — CLI + API 密钥
- `026` — 密钥追踪
- `032` — 用户 CLI 凭证
- `036` — Agent 授权

### Agent 与技能
- `006` — 内置工具
- `010` — Agent Markdown
- `015` — 成本预算
- `017` — 系统技能
- `044` — 种植任务文件
- `055` — 扩展字段

### 计费与监管
- `016` — 使用统计
- `009` — 配额索引

### 性能优化
- `030` — JSONB 索引
- `051` — 上下文修剪

### Hooks 系统（可选扩展）
- `052` → `053` → `054` — Hooks 核心 + 脚本 + 名称

---

## 迁移依赖图

```
001 (init_schema)
 ├─→ 005 (handoff_routes)
 ├─→ 006 (builtin_tools)
 ├─→ 009 (quota_index)
 ├─→ 010 (agents_md_v2)
 ├─→ 011 (session_metadata)
 ├─→ 012 (channel_pending_messages)
 ├─→ 014 (channel_contacts)
 ├─→ 015 (agent_budget)
 ├─→ 016 (usage_snapshots) ← 009 (index)
 ├─→ 017 (system_skills)
 ├─→ 020 (secure_cli_and_api_keys)
 ├─→ 026 (api_key_user_binding) ← 020 (api_keys table)
 ├─→ 027 (tenant_foundation) ← 加入 tenant_id 到 30+ 表
 ├─→ 029 (system_configs)
 ├─→ 030 (jsonb_gin_indexes)
 ├─→ 032 (secure_cli_user_credentials) ← 020
 ├─→ 036 (agent_cli_grants) ← 020 + 001 (agents)
 ├─→ 044 (seed_core_task_files)
 ├─→ 051 (backfill_context_pruning)
 ├─→ 052 (agent_hooks) ← 027 (tenant_id)
 ├─→ 053 (agent_hooks_script)
 ├─→ 054 (agent_hooks_name)
 └─→ 055 (agent_extended_fields)

所有迁移必须按序执行（无跳过）
```

---

## 修改 Schema 时的检查清单

### PostgreSQL 部分
- [ ] 创建新迁移文件 `migrations/00NNN_description.{up,down}.sql`
- [ ] 编写 `.up.sql`（应用迁移）
- [ ] 编写 `.down.sql`（回滚，测试用）
- [ ] 更新 `internal/upgrade/version.go` → `RequiredSchemaVersion`
- [ ] 在测试 PG 上验证：`migrate up` 和 `migrate down`

### SQLite 部分（重要！）
- [ ] 更新 `internal/store/sqlitestore/schema.sql`（完整 schema）
- [ ] 在 `internal/store/sqlitestore/schema.go` 中添加增量补丁
- [ ] 更新 `SchemaVersion` 常数
- [ ] 验证构建：`go build -tags sqliteonly ./...`
- [ ] 在桌面版上测试（如有条件）

### 通用
- [ ] 如果添加了用户可见的新列，更新 i18n 字符串
- [ ] 如果新增表，检查是否需要 `tenant_id` 列
- [ ] 运行 `go vet ./...` 和 `go test ./...`
- [ ] 提交前验证迁移号不重复

---

## 常见迁移场景

### 场景 1：添加新表
```sql
-- 000056_new_feature.up.sql
CREATE TABLE new_feature (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_new_feature_agent ON new_feature(agent_id);

-- 000056_new_feature.down.sql
DROP TABLE new_feature;
```

### 场景 2：添加新列到现有表
```sql
-- .up.sql
ALTER TABLE agents ADD COLUMN new_field VARCHAR(255);

-- .down.sql
ALTER TABLE agents DROP COLUMN new_field;
```

### 场景 3：添加索引以优化查询
```sql
-- .up.sql
CREATE INDEX idx_spans_agent_time
  ON spans(agent_id, start_time DESC)
  WHERE status = 'completed';

-- .down.sql
DROP INDEX idx_spans_agent_time;
```

### 场景 4：数据迁移（修改现有数据）
```sql
-- .up.sql
ALTER TABLE agents ADD COLUMN new_enum VARCHAR(20);
UPDATE agents SET new_enum = 'default' WHERE new_enum IS NULL;
ALTER TABLE agents ADD NOT NULL CHECK (new_enum IN ('a', 'b', 'c'));

-- .down.sql
ALTER TABLE agents DROP COLUMN new_enum;
```

---

## 调试迁移问题

### 检查当前版本
```sql
SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;
```

### 查看迁移历史
```sql
SELECT * FROM schema_migrations ORDER BY version;
```

### 清空测试 DB（开发用）
```bash
dropdb goclaw_test 2>/dev/null || true
createdb goclaw_test
DATABASE_URL="postgres://localhost/goclaw_test" ./goclaw migrate up
```

### 回滚单个迁移（开发用）
```bash
DATABASE_URL="postgres://localhost/goclaw_test" ./goclaw migrate down 1
DATABASE_URL="postgres://localhost/goclaw_test" ./goclaw migrate up
```

### PostgreSQL vs SQLite 版本不同步的症状
- ❌ 桌面版启动时崩溃：`schema version mismatch`
- ✅ 解决：确保两个系统的版本号一致

---

## 参考链接

- 完整迁移详解：`docs/25-database-migrations.md`
- Schema 工具：`internal/store/sqlitestore/schema.go`
- PG 迁移：`migrations/`
- 版本控制：`internal/upgrade/version.go`
- SQLite 完整 schema：`internal/store/sqlitestore/schema.sql`

