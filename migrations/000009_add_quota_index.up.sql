-- Partial index for quota checker: efficiently counts traces per user in time windows.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_traces_quota
ON traces (user_id, created_at DESC)
WHERE user_id IS NOT NULL;
