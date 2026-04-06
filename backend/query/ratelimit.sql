-- name: UpsertRatelimit :one
INSERT INTO t_ratelimit(
    user_public_id,
    consumed_quota
)
VALUES (@user_id, @quota)
ON CONFLICT (user_public_id) DO UPDATE SET
    consumed_quota = CASE
        WHEN t_ratelimit.updated_at AT TIME ZONE 'America/Los_Angeles' < date_trunc('day', current_timestamp AT TIME ZONE 'America/Los_Angeles') THEN EXCLUDED.consumed_quota
        ELSE t_ratelimit.consumed_quota + EXCLUDED.consumed_quota
    END,
    updated_at = current_timestamp
RETURNING t_ratelimit_id, consumed_quota;
