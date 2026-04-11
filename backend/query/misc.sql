-- トランザクションレベルのロック（ノンブロッキング）
-- name: TryAcquireAdvisoryXactLock :one
SELECT pg_try_advisory_xact_lock($1::bigint) AS acquired;

-- セッションレベルのロック（ノンブロッキング）
-- name: TryAcquireAdvisoryLock :one
SELECT pg_try_advisory_lock($1::bigint) AS acquired;

-- セッションレベルのアドバイザリロックを解除
-- name: ReleaseAdvisoryLock :one
SELECT pg_advisory_unlock($1::bigint) AS released;
