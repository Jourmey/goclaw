-- Add extended agent fields for v3.x features

ALTER TABLE agents
    ADD COLUMN IF NOT EXISTS frontmatter TEXT,
    ADD COLUMN IF NOT EXISTS emoji VARCHAR(10) DEFAULT '',
    ADD COLUMN IF NOT EXISTS agent_description TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS thinking_level VARCHAR(20) DEFAULT 'default',
    ADD COLUMN IF NOT EXISTS max_tokens INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS self_evolve BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS skill_evolve BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS skill_nudge_interval INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reasoning_config JSONB,
    ADD COLUMN IF NOT EXISTS workspace_sharing JSONB,
    ADD COLUMN IF NOT EXISTS chatgpt_oauth_routing JSONB,
    ADD COLUMN IF NOT EXISTS shell_deny_groups JSONB,
    ADD COLUMN IF NOT EXISTS kg_dedup_config JSONB;
