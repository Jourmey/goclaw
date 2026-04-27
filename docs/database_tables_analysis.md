# GoClaw 数据库表分析与瘦身建议

**总计：65 张表**

## 📊 核心模块分类

### ⭐ P0 - 核心基础（不可移除）

#### 租户与用户系统 (2 表)
- `tenants` - 租户表
- `tenant_users` - 租户用户表

#### LLM 与 Agent 核心 (11 表)
- `llm_providers` - LLM 提供商配置
- `agents` - Agent 定义
- `agent_shares` - Agent 共享
- `agent_context_files` - Agent 上下文文件
- `user_context_files` - 用户级上下文文件
- `user_agent_profiles` - 用户 Agent 配置
- `user_agent_overrides` - 用户覆写配置
- `sessions` - 会话记录（对话历史）
- `agent_config_permissions` - Agent 配置权限
- `system_configs` - 系统配置
- `config_secrets` - 配置加密秘钥

#### 基础工具系统 (2 表)
- `builtin_tools` - 内置工具定义
- `builtin_tool_tenant_configs` - 租户级工具配置

**小计：15 表 (23%)**

---

### 🔥 P1 - 重要功能（轻度使用可考虑移除）

#### 技能系统 (4 表) - **建议保留**
- `skills` - 技能定义
- `skill_agent_grants` - Agent 级技能授权
- `skill_user_grants` - 用户级技能授权
- `skill_tenant_configs` - 租户级技能配置

**理由：技能系统是 GoClaw 的核心差异化功能**

#### 内存系统 (3 表) - **建议简化**
- `memory_documents` - 记忆文档
- `memory_chunks` - 记忆片段（向量检索）
- `embedding_cache` - Embedding 缓存

**瘦身建议：**
- ✅ 保留 `memory_documents` 和 `memory_chunks`（核心 RAG）
- ❌ 移除 `embedding_cache`（性能优化非必需，重新计算即可）

#### API 密钥系统 (1 表) - **建议保留**
- `api_keys` - API 密钥管理

**小计：8 表 (12%)**

---

### 🟡 P2 - 增值功能（中等复杂度，可移除）

#### MCP 服务器桥接 (5 表) - **建议移除**
- `mcp_servers` - MCP 服务器配置
- `mcp_agent_grants` - Agent 级 MCP 授权
- `mcp_user_grants` - 用户级 MCP 授权
- `mcp_access_requests` - MCP 访问请求
- `mcp_user_credentials` - MCP 用户凭证

**理由：MCP 桥接是高级功能，普通用户使用率低**

#### Knowledge Vault (2 表) - **建议移除**
- `vault_documents` - Vault 文档
- `vault_links` - Wikilink 关系

**理由：与 memory 系统功能重叠，可合并到 memory 体系**

#### Secure CLI 二进制管理 (3 表) - **建议移除**
- `secure_cli_binaries` - 安全 CLI 二进制
- `secure_cli_agent_grants` - Agent 级授权
- `secure_cli_user_credentials` - 用户凭证

**理由：安全沙箱功能，非核心用户场景较少**

#### Hooks 系统 (4 表) - **建议移除**
- `hooks` - Hook 定义
- `hook_agents` - Hook-Agent 关联
- `hook_executions` - Hook 执行日志
- `tenant_hook_budget` - Hook 预算管理

**理由：扩展性功能，可通过外部 webhook 替代**

**小计：14 表 (22%)**

---

### 🔴 P3 - 高级特性（强烈建议移除）

#### Team 协作系统 (7 表) - **强烈建议移除**
- `agent_teams` - Agent 团队
- `agent_team_members` - 团队成员
- `team_tasks` - 团队任务
- `team_task_comments` - 任务评论
- `team_task_events` - 任务事件
- `team_task_attachments` - 任务附件
- `team_user_grants` - 用户团队权限

**理由：团队协作是企业级功能，增加大量复杂度**

#### Agent 进化系统 (3 表) - **强烈建议移除**
- `agent_evolution_metrics` - 进化指标
- `agent_evolution_suggestions` - 进化建议
- `episodic_summaries` - 片段式总结（记忆巩固）

**理由：实验性自适应功能，实际效果有限**

#### 知识图谱 (3 表) - **强烈建议移除**
- `kg_entities` - 知识图谱实体
- `kg_relations` - 知识图谱关系
- `kg_dedup_candidates` - 去重候选

**理由：复杂度高，与向量检索功能重叠**

#### Agent 心跳监控 (2 表) - **强烈建议移除**
- `agent_heartbeats` - Agent 心跳
- `heartbeat_run_logs` - 心跳运行日志

**理由：运维监控功能，非核心业务**

#### Subagent 子任务系统 (2 表) - **考虑保留**
- `subagent_tasks` - 子 Agent 任务
- `agent_links` - Agent 链接（委托关系）

**理由：虽然是高级功能，但 delegate 工具依赖这些表**

#### Tracing 追踪系统 (2 表) - **强烈建议移除**
- `traces` - 调用链追踪
- `spans` - Span 记录

**理由：OTel 追踪是可选调试功能，生产环境不需要**

#### Channel 渠道系统 (3 表) - **考虑保留部分**
- `channel_instances` - 渠道实例（Telegram/Discord/WhatsApp）
- `channel_pending_messages` - 待发送消息队列
- `channel_contacts` - 渠道联系人

**理由：如果不需要多渠道接入，可全部移除**

#### Pairing 配对系统 (2 表) - **建议移除**
- `pairing_requests` - 配对请求
- `paired_devices` - 已配对设备

**理由：设备配对是桌面版功能，Web 版不需要**

**小计：24 表 (37%)**

---

### 📈 P4 - 监控统计（可选）

#### 日志与统计 (4 表)
- `activity_logs` - 活动日志
- `usage_snapshots` - 使用量快照
- `cron_jobs` - 定时任务
- `cron_run_logs` - 定时任务运行日志

**瘦身建议：**
- ✅ 保留 `cron_jobs`（如果需要定时任务功能）
- ❌ 移除 `activity_logs`, `usage_snapshots`, `cron_run_logs`（可用日志系统替代）

**小计：4 表 (6%)**

---

## 🎯 瘦身方案建议

### 方案 A：激进瘦身（保留 20-25 张表，减少 60%+）

**移除模块：**
1. ❌ Team 协作系统 (7 表)
2. ❌ 知识图谱 (3 表)
3. ❌ Agent 进化 (3 表)
4. ❌ MCP 桥接 (5 表)
5. ❌ Vault (2 表)
6. ❌ Secure CLI (3 表)
7. ❌ Hooks (4 表)
8. ❌ Tracing (2 表)
9. ❌ Pairing (2 表)
10. ❌ Channel 系统 (3 表)
11. ❌ 心跳监控 (2 表)
12. ❌ 日志统计 (3 表: activity_logs, usage_snapshots, cron_run_logs)
13. ❌ Embedding cache (1 表)

**保留：**
- ✅ 租户/用户 (2)
- ✅ Agent 核心 (11)
- ✅ 技能系统 (4)
- ✅ 内存系统 (2: documents + chunks)
- ✅ 工具系统 (2)
- ✅ API 密钥 (1)
- ✅ Subagent (2: 如需 delegate 功能)
- ✅ Cron (1: 如需定时任务)

**预计剩余：25 张表**

---

### 方案 B：温和瘦身（保留 35-40 张表，减少 40%）

在方案 A 基础上，额外保留：
- ✅ Channel 系统 (3 表) - 如果需要 Telegram/Discord 接入
- ✅ Vault (2 表) - 如果需要 Wikilink 知识库
- ✅ Cron 日志 (1 表) - 如果需要任务历史
- ✅ Activity logs (1 表) - 如果需要审计

**预计剩余：32-35 张表**

---

## 🔧 实施步骤

### 第一阶段：移除最小价值模块（减少 15 表）

```bash
# 1. Team 协作系统
rm -rf internal/teams/
# 删除 7 张表相关代码

# 2. 知识图谱
rm -rf internal/knowledgegraph/
# 删除 3 张表相关代码

# 3. Agent 进化
rm -rf internal/consolidation/
# 删除 3 张表相关代码

# 4. Tracing
rm -rf internal/tracing/
# 删除 2 张表相关代码
```

### 第二阶段：移除中等价值模块（减少 14 表）

```bash
# 5. MCP 桥接
# 移除 internal/mcp/ 相关 5 张表

# 6. Secure CLI
# 移除 3 张表

# 7. Hooks
# 移除 4 张表

# 8. Pairing
# 移除 2 张表
```

### 第三阶段：清理监控统计（减少 6 表）

```bash
# 9. Vault → 合并到 memory
# 10. Channel → 按需保留
# 11. 清理日志表
```

---

## 📊 预期收益

| 方案 | 移除表数 | 剩余表数 | 减少比例 | 预计代码减少 |
|------|---------|---------|---------|------------|
| 激进 | 40 表 | 25 表 | 62% | 50,000+ 行 |
| 温和 | 30 表 | 35 表 | 46% | 30,000+ 行 |

---

## ⚠️ 风险评估

**低风险移除（优先）：**
- Agent 进化、KG、Tracing、Pairing、心跳监控

**中风险移除（需评估）：**
- Team 系统（如果有团队协作需求）
- Channel 系统（如果有多渠道需求）
- MCP 桥接（如果有 VSCode 插件集成）

**高风险移除（慎重）：**
- Subagent/delegate（影响 agent 协作）
- Memory 系统（影响 RAG）
- Skills 系统（核心功能）
