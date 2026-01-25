-- トランザクションレベルのロック（トランザクション終了時に自動解放）
-- name: AcquireAdvisoryXactLock :exec
SELECT pg_advisory_xact_lock($1::bigint);

-- トランザクションレベルのロック（ノンブロッキング）
-- name: TryAcquireAdvisoryXactLock :one
SELECT pg_try_advisory_xact_lock($1::bigint) AS acquired;
