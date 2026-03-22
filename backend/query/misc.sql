-- トランザクションレベルのロック（ノンブロッキング）
-- name: TryAcquireAdvisoryXactLock :one
SELECT pg_try_advisory_xact_lock($1::bigint) AS acquired;
