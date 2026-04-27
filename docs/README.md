# GoClaw Database Schema

**Version:** Schema v55
**Total Tables:** 34
**Last Updated:** Generated from migration files

## 📚 Files

- `00_full_schema.sql` - Complete schema definition (all 34 tables)
- `TABLE_INDEX.md` - Alphabetical table index with metadata
- `README.md` - This file

## 🗂️ Table Categories

### Core Multi-Tenant (2)
- `mcp_user_credentials` (000027_tenant_foundation.up.sql)
- `tenant_users` (000027_tenant_foundation.up.sql)

### Agent Management (6)
- `agent_context_files` (000001_init_schema.up.sql)
- `agent_shares` (000001_init_schema.up.sql)
- `agents` (000001_init_schema.up.sql)
- `user_agent_overrides` (000001_init_schema.up.sql)
- `user_agent_profiles` (000001_init_schema.up.sql)
- `user_context_files` (000001_init_schema.up.sql)

### Sessions & Memory (2)
- `memory_documents` (000001_init_schema.up.sql)
- `sessions` (000001_init_schema.up.sql)

### LLM & Providers (1)
- `llm_providers` (000001_init_schema.up.sql)

### Skills & Tools (6)
- `builtin_tool_tenant_configs` (000027_tenant_foundation.up.sql)
- `custom_tools` (000001_init_schema.up.sql)
- `skill_agent_grants` (000001_init_schema.up.sql)
- `skill_tenant_configs` (000027_tenant_foundation.up.sql)
- `skill_user_grants` (000001_init_schema.up.sql)
- `skills` (000001_init_schema.up.sql)

### Security & Access (2)
- `api_keys` (000020_secure_cli_and_api_keys.up.sql)
- `secure_cli_binaries` (000020_secure_cli_and_api_keys.up.sql)

### Scheduling & Cron (2)
- `cron_jobs` (000001_init_schema.up.sql)
- `cron_run_logs` (000001_init_schema.up.sql)

### Pairing & Devices (2)
- `paired_devices` (000001_init_schema.up.sql)
- `pairing_requests` (000001_init_schema.up.sql)

### MCP Integration (4)
- `mcp_access_requests` (000001_init_schema.up.sql)
- `mcp_agent_grants` (000001_init_schema.up.sql)
- `mcp_servers` (000001_init_schema.up.sql)
- `mcp_user_grants` (000001_init_schema.up.sql)

### Channels & Communication (1)
- `channel_instances` (000001_init_schema.up.sql)

### Configuration (1)
- `config_secrets` (000001_init_schema.up.sql)

### Collaboration (1)
- `group_file_writers` (000001_init_schema.up.sql)

### Tracing & Monitoring (3)
- `activity_logs` (000015_agent_budget.up.sql)
- `spans` (000001_init_schema.up.sql)
- `traces` (000001_init_schema.up.sql)

### Cache & Embeddings (1)
- `embedding_cache` (000001_init_schema.up.sql)

