-- API key owner binding for identity enforcement
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS owner_id VARCHAR(255);
COMMENT ON COLUMN api_keys.owner_id IS 'User who owns this key. When set, auth via this key forces user_id = owner_id.';
CREATE INDEX IF NOT EXISTS idx_api_keys_owner_id ON api_keys(owner_id) WHERE owner_id IS NOT NULL;

-- Drop unused legacy tables
DROP TABLE IF EXISTS handoff_routes;
DROP TABLE IF EXISTS delegation_history;
