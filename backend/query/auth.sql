-- name: CreateAuthorization :one
INSERT INTO m_user_authorization (issuer, sub)
VALUES ($1, $2)
ON CONFLICT (issuer, sub) DO UPDATE
    SET issuer = EXCLUDED.issuer
RETURNING m_user_authorization_id, public_id, (xmax = 0) AS is_created;;

-- name: CreateRefreshToken :exec
INSERT INTO m_refresh_token (m_user_authorization_id, token_hash, ip_address, device_fingerprint, user_agent,
                             country_code, city_name, browser_name, device_type, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT DO NOTHING;

-- name: GetUserIDByAuthorization :one
SELECT public_id
FROM m_user
WHERE m_user_authorization_id = $1;

-- name: SaveJTIBlacklist :exec
INSERT INTO t_jti_blacklist (jti, expires_at)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: IsJTIBlacklisted :one
SELECT expires_at FROM t_jti_blacklist WHERE jti = $1;

-- name: RemoveRefreshToken :one
DELETE
FROM m_refresh_token
WHERE token_hash = $1
RETURNING token_hash;

-- name: SaveRefreshToken :one
UPDATE m_refresh_token
SET token_hash         = $1,
    expires_at         = $2,
    ip_address         = $3,
    device_fingerprint = $4,
    user_agent         = $5,
    country_code       = $6,
    city_name          = $7,
    browser_name       = $8,
    device_type        = $9,
    updated_at         = current_timestamp,
    generation         = generation + 1
WHERE token_hash = $10
  AND $11 < expires_at
  AND updated_at < $12
  AND device_fingerprint = $4
RETURNING m_user_authorization_id;

-- name: CleanupExpiredJTIBlacklist :exec
DELETE FROM t_jti_blacklist WHERE expires_at < $1;

-- name: GetUserAuthorizationID :one
SELECT m_user_authorization_id FROM m_user WHERE public_id = $1;

-- name: GetUserAllRefreshTokens :many
SELECT public_id, created_at, updated_at, country_code, city_name, browser_name FROM m_refresh_token WHERE m_user_authorization_id = $1;

-- name: RemoveRefreshTokenByID :one
DELETE FROM m_refresh_token WHERE public_id = $1 RETURNING public_id;
