-- Rollback extended agent fields

ALTER TABLE agents
    DROP COLUMN IF EXISTS frontmatter,
    DROP COLUMN IF EXISTS emoji,
    DROP COLUMN IF EXISTS agent_description,
    DROP COLUMN IF EXISTS thinking_level,
    DROP COLUMN IF EXISTS max_tokens,
    DROP COLUMN IF EXISTS self_evolve,
    DROP COLUMN IF EXISTS skill_evolve,
    DROP COLUMN IF EXISTS skill_nudge_interval,
    DROP COLUMN IF EXISTS reasoning_config,
    DROP COLUMN IF EXISTS workspace_sharing,
    DROP COLUMN IF EXISTS chatgpt_oauth_routing,
    DROP COLUMN IF EXISTS shell_deny_groups,
    DROP COLUMN IF EXISTS kg_dedup_config;
