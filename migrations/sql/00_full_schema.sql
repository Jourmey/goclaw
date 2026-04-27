-- GoClaw Full Database Schema
-- Total tables: 34
-- Combined from all migrations

-- Table: llm_providers
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE llm_providers (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name          VARCHAR(50) NOT NULL UNIQUE,
    display_name  VARCHAR(255),
    provider_type VARCHAR(30) NOT NULL DEFAULT 'openai_compat',
    api_base      TEXT,
    api_key       TEXT,
    enabled       BOOLEAN NOT NULL DEFAULT true,
    settings      JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Table: agents
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE agents (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_key             VARCHAR(100) NOT NULL UNIQUE,
    display_name          VARCHAR(255),
    owner_id              VARCHAR(255) NOT NULL,
    provider              VARCHAR(50) NOT NULL DEFAULT 'openrouter',
    model                 VARCHAR(200) NOT NULL,
    context_window        INT NOT NULL DEFAULT 200000,
    max_tool_iterations   INT NOT NULL DEFAULT 20,
    workspace             TEXT NOT NULL DEFAULT '.',
    restrict_to_workspace BOOLEAN NOT NULL DEFAULT true,
    tools_config          JSONB NOT NULL DEFAULT '{}',
    sandbox_config        JSONB,
    subagents_config      JSONB,
    memory_config         JSONB,
    compaction_config     JSONB,
    context_pruning       JSONB,
    other_config          JSONB NOT NULL DEFAULT '{}',
    is_default            BOOLEAN NOT NULL DEFAULT false,
    agent_type            VARCHAR(20) NOT NULL DEFAULT 'open',
    status                VARCHAR(20) DEFAULT 'active',
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

-- Table: agent_shares
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE agent_shares (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id   UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id    VARCHAR(255) NOT NULL,
    role       VARCHAR(20) NOT NULL DEFAULT 'user',
    granted_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_id, user_id)
);

-- Table: agent_context_files
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE agent_context_files (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id   UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    file_name  VARCHAR(255) NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_id, file_name)
);

-- Table: user_context_files
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE user_context_files (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id   UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id    VARCHAR(255) NOT NULL,
    file_name  VARCHAR(255) NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_id, user_id, file_name)
);

-- Table: user_agent_profiles
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE user_agent_profiles (
    agent_id      UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id       VARCHAR(255) NOT NULL,
    workspace     TEXT,
    first_seen_at TIMESTAMPTZ DEFAULT NOW(),
    last_seen_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (agent_id, user_id)
);

-- Table: user_agent_overrides
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE user_agent_overrides (
    id       UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id  VARCHAR(255) NOT NULL,
    provider VARCHAR(50),
    model    VARCHAR(200),
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_id, user_id)
);

-- Table: sessions
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE sessions (
    id                            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    session_key                   VARCHAR(500) NOT NULL UNIQUE,
    agent_id                      UUID REFERENCES agents(id),
    user_id                       VARCHAR(255),
    messages                      JSONB NOT NULL DEFAULT '[]',
    summary                       TEXT,
    model                         VARCHAR(200),
    provider                      VARCHAR(50),
    channel                       VARCHAR(50),
    input_tokens                  BIGINT NOT NULL DEFAULT 0,
    output_tokens                 BIGINT NOT NULL DEFAULT 0,
    compaction_count              INT NOT NULL DEFAULT 0,
    memory_flush_compaction_count INT NOT NULL DEFAULT 0,
    memory_flush_at               BIGINT DEFAULT 0,
    label                         VARCHAR(500),
    spawned_by                    VARCHAR(200),
    spawn_depth                   INT NOT NULL DEFAULT 0,
    created_at                    TIMESTAMPTZ DEFAULT NOW(),
    updated_at                    TIMESTAMPTZ DEFAULT NOW()
);

-- Table: memory_documents
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE memory_documents (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id   UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id    VARCHAR(255),
    path       VARCHAR(500) NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    hash       VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: embedding_cache
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE embedding_cache (
    hash      VARCHAR(64) NOT NULL,
    provider  VARCHAR(50) NOT NULL,
    model     VARCHAR(200) NOT NULL,
    embedding vector(1536),
    dims      INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (hash, provider, model)
);

-- Table: skills
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE skills (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name         VARCHAR(255) NOT NULL,
    slug         VARCHAR(255) NOT NULL UNIQUE,
    description  TEXT,
    owner_id     VARCHAR(255) NOT NULL,
    visibility   VARCHAR(10) NOT NULL DEFAULT 'private',
    version      INT NOT NULL DEFAULT 1,
    status       VARCHAR(20) NOT NULL DEFAULT 'active',
    frontmatter  JSONB NOT NULL DEFAULT '{}',
    file_path    TEXT NOT NULL,
    file_size    BIGINT NOT NULL DEFAULT 0,
    file_hash    VARCHAR(64),
    embedding    vector(1536),
    tags         TEXT[],
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Table: skill_agent_grants
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE skill_agent_grants (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id       UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    agent_id       UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    pinned_version INT NOT NULL,
    granted_by     VARCHAR(255) NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(skill_id, agent_id)
);

-- Table: skill_user_grants
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE skill_user_grants (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id   UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    user_id    VARCHAR(255) NOT NULL,
    granted_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(skill_id, user_id)
);

-- Table: cron_jobs
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE cron_jobs (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id         UUID REFERENCES agents(id),
    user_id          TEXT,
    name             VARCHAR(255) NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT true,
    schedule_kind    VARCHAR(10) NOT NULL,
    cron_expression  VARCHAR(100),
    interval_ms      BIGINT,
    run_at           TIMESTAMPTZ,
    timezone         VARCHAR(50),
    payload          JSONB NOT NULL,
    delete_after_run BOOLEAN NOT NULL DEFAULT false,
    next_run_at      TIMESTAMPTZ,
    last_run_at      TIMESTAMPTZ,
    last_status      VARCHAR(20),
    last_error       TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

-- Table: cron_run_logs
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE cron_run_logs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    job_id        UUID NOT NULL REFERENCES cron_jobs(id) ON DELETE CASCADE,
    agent_id      UUID REFERENCES agents(id),
    status        VARCHAR(20) NOT NULL,
    summary       TEXT,
    error         TEXT,
    duration_ms   INT,
    input_tokens  INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    ran_at        TIMESTAMPTZ DEFAULT NOW(),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Table: pairing_requests
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE pairing_requests (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    code       VARCHAR(8) NOT NULL UNIQUE,
    sender_id  VARCHAR(200) NOT NULL,
    channel    VARCHAR(255) NOT NULL,
    chat_id    VARCHAR(200) NOT NULL,
    account_id VARCHAR(100) NOT NULL DEFAULT 'default',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: paired_devices
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE paired_devices (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    sender_id VARCHAR(200) NOT NULL,
    channel   VARCHAR(255) NOT NULL,
    chat_id   VARCHAR(200) NOT NULL,
    paired_by VARCHAR(100) NOT NULL DEFAULT 'operator',
    paired_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(sender_id, channel)
);

-- Table: traces
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE traces (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    agent_id            UUID,
    user_id             VARCHAR(255),
    session_key         TEXT,
    run_id              TEXT,
    start_time          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time            TIMESTAMPTZ,
    duration_ms         INT,
    name                TEXT,
    channel             VARCHAR(50),
    input_preview       TEXT,
    output_preview      TEXT,
    total_input_tokens  INT DEFAULT 0,
    total_output_tokens INT DEFAULT 0,
    total_cost          NUMERIC(12,6) DEFAULT 0,
    span_count          INT DEFAULT 0,
    llm_call_count      INT DEFAULT 0,
    tool_call_count     INT DEFAULT 0,
    status              VARCHAR(20) DEFAULT 'running',
    error               TEXT,
    metadata            JSONB,
    tags                TEXT[],
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: spans
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE spans (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    trace_id       UUID NOT NULL,
    parent_span_id UUID,
    agent_id       UUID,
    span_type      VARCHAR(20) NOT NULL,
    name           TEXT,
    start_time     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time       TIMESTAMPTZ,
    duration_ms    INT,
    status         VARCHAR(20) DEFAULT 'running',
    error          TEXT,
    level          VARCHAR(10) DEFAULT 'DEFAULT',
    model          VARCHAR(200),
    provider       VARCHAR(50),
    input_tokens   INT,
    output_tokens  INT,
    total_cost     NUMERIC(12,8),
    finish_reason  VARCHAR(50),
    model_params   JSONB,
    tool_name      VARCHAR(200),
    tool_call_id   VARCHAR(100),
    input_preview  TEXT,
    output_preview TEXT,
    metadata       JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: mcp_servers
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE mcp_servers (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name          VARCHAR(255) NOT NULL UNIQUE,
    display_name  VARCHAR(255),
    transport     VARCHAR(50) NOT NULL,            -- "stdio", "sse", "streamable-http"
    command       TEXT,                             -- stdio: command to spawn
    args          JSONB DEFAULT '[]',               -- stdio: command arguments
    url           TEXT,                             -- sse/http: server URL
    headers       JSONB DEFAULT '{}',               -- sse/http: HTTP headers
    env           JSONB DEFAULT '{}',               -- stdio: environment variables
    api_key       TEXT,                             -- encrypted (AES-256-GCM)
    tool_prefix   VARCHAR(50),                      -- optional prefix for tool names
    timeout_sec   INT DEFAULT 60,
    settings      JSONB NOT NULL DEFAULT '{}',
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_by    VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Table: mcp_agent_grants
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE mcp_agent_grants (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    server_id        UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    agent_id         UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    enabled          BOOLEAN NOT NULL DEFAULT true,
    tool_allow       JSONB,                         -- ["tool1", "tool2"] (null = all)
    tool_deny        JSONB,                         -- ["dangerous_tool"]
    config_overrides JSONB,
    granted_by       VARCHAR(255) NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(server_id, agent_id)
);

-- Table: mcp_user_grants
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE mcp_user_grants (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    server_id        UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    user_id          VARCHAR(255) NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT true,
    tool_allow       JSONB,
    tool_deny        JSONB,
    granted_by       VARCHAR(255) NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(server_id, user_id)
);

-- Table: mcp_access_requests
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE mcp_access_requests (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    server_id     UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    agent_id      UUID REFERENCES agents(id) ON DELETE CASCADE,
    user_id       VARCHAR(255),
    scope         VARCHAR(10) NOT NULL,             -- "agent" or "user"
    status        VARCHAR(20) NOT NULL DEFAULT 'pending', -- "pending", "approved", "rejected"
    reason        TEXT,
    tool_allow    JSONB,                            -- requested tool subset (null = all)
    requested_by  VARCHAR(255) NOT NULL,
    reviewed_by   VARCHAR(255),
    reviewed_at   TIMESTAMPTZ,
    review_note   TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Table: custom_tools
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE custom_tools (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name            VARCHAR(100) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    parameters      JSONB NOT NULL DEFAULT '{}',
    command         TEXT NOT NULL,
    working_dir     TEXT DEFAULT '',
    timeout_seconds INT DEFAULT 60,
    env             BYTEA,                               -- encrypted env vars (AES-256-GCM)
    agent_id        UUID REFERENCES agents(id) ON DELETE CASCADE,
    enabled         BOOLEAN DEFAULT TRUE,
    created_by      VARCHAR(255) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Table: channel_instances
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE channel_instances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name            VARCHAR(100) NOT NULL UNIQUE,
    display_name    VARCHAR(255) DEFAULT '',
    channel_type    VARCHAR(50) NOT NULL,
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    credentials     BYTEA,
    config          JSONB DEFAULT '{}',
    enabled         BOOLEAN DEFAULT true,
    created_by      VARCHAR(255) DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Table: config_secrets
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE config_secrets (
    key         VARCHAR(100) PRIMARY KEY,
    value       BYTEA NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Table: group_file_writers
-- Source: migrations/000001_init_schema.up.sql
CREATE TABLE group_file_writers (
    agent_id     UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    group_id     VARCHAR(255) NOT NULL,
    user_id      VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    username     VARCHAR(255),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (agent_id, group_id, user_id)
);

-- Table: activity_logs
-- Source: migrations/000015_agent_budget.up.sql
CREATE TABLE activity_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    actor_type  VARCHAR(20) NOT NULL,
    actor_id    VARCHAR(255) NOT NULL,
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id   VARCHAR(255),
    details     JSONB,
    ip_address  VARCHAR(45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: secure_cli_binaries
-- Source: migrations/000020_secure_cli_and_api_keys.up.sql
CREATE TABLE secure_cli_binaries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    binary_name     TEXT NOT NULL,                          -- display name: "gh", "gcloud"
    binary_path     TEXT,                                   -- resolved absolute path (nullable, auto-resolved at runtime)
    description     TEXT NOT NULL DEFAULT '',
    encrypted_env   BYTEA NOT NULL,                         -- AES-256-GCM encrypted JSON: {"GH_TOKEN":"xxx"}
    deny_args       JSONB NOT NULL DEFAULT '[]',            -- regex patterns: ["auth\\s+", "ssh-key"]
    deny_verbose    JSONB NOT NULL DEFAULT '[]',            -- verbose flag patterns: ["--verbose", "-v"]
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    tips            TEXT NOT NULL DEFAULT '',                -- hint injected into TOOLS.md context
    agent_id        UUID REFERENCES agents(id) ON DELETE CASCADE,  -- null = global (all agents)
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_by      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Table: api_keys
-- Source: migrations/000020_secure_cli_and_api_keys.up.sql
CREATE TABLE api_keys (
    id            UUID PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    prefix        VARCHAR(8)   NOT NULL,              -- first 8 chars for display identification
    key_hash      VARCHAR(64)  NOT NULL UNIQUE,       -- SHA-256 hex digest
    scopes        TEXT[]       NOT NULL DEFAULT '{}',  -- e.g. {'operator.admin','operator.read'}
    expires_at    TIMESTAMPTZ,                         -- NULL = never expires
    last_used_at  TIMESTAMPTZ,
    revoked       BOOLEAN      NOT NULL DEFAULT false,
    created_by    VARCHAR(255),                        -- user ID who created the key
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Table: tenant_users
-- Source: migrations/000027_tenant_foundation.up.sql
CREATE TABLE tenant_users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id      VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    role         VARCHAR(20) NOT NULL DEFAULT 'member',
    metadata     JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

-- Table: builtin_tool_tenant_configs
-- Source: migrations/000027_tenant_foundation.up.sql
CREATE TABLE builtin_tool_tenant_configs (
    tool_name  VARCHAR(100) NOT NULL REFERENCES builtin_tools(name) ON DELETE CASCADE,
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    enabled    BOOLEAN,
    settings   JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tool_name, tenant_id)
);

-- Table: skill_tenant_configs
-- Source: migrations/000027_tenant_foundation.up.sql
CREATE TABLE skill_tenant_configs (
    skill_id   UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    enabled    BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (skill_id, tenant_id)
);

-- Table: mcp_user_credentials
-- Source: migrations/000027_tenant_foundation.up.sql
CREATE TABLE mcp_user_credentials (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id  UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    user_id    VARCHAR(255) NOT NULL,
    api_key    TEXT,
    headers    BYTEA,
    env        BYTEA,
    tenant_id  UUID NOT NULL REFERENCES tenants(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(server_id, user_id, tenant_id)
);

