-- トランザクションレベルのロック（ノンブロッキング）
-- name: TryAcquireAdvisoryXactLock :one
SELECT pg_try_advisory_xact_lock(@lock_key::bigint) AS acquired;

-- セッションレベルのロック（ノンブロッキング）
-- name: TryAcquireAdvisoryLock :one
SELECT pg_try_advisory_lock(@lock_key::bigint) AS acquired;

-- セッションレベルのアドバイザリロックを解除
-- name: ReleaseAdvisoryLock :one
SELECT pg_advisory_unlock(@lock_key::bigint) AS released;
