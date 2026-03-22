-- リフレッシュトークンをテーブルに保存する。
-- m_refresh_token_idが返される。
-- name: SaveRefreshToken :one
INSERT INTO
    m_refresh_token (
        m_user_authorization_id,
        token_hash,
        ip_address,
        device_fingerprint,
        user_agent,
        country_code,
        city_name,
        browser_name,
        device_type,
        expires_at,
        access_token_jti,
        activated_at,
        last_logged_in_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING
    m_refresh_token_id;

-- リフレッシュトークンを更新します。
-- token_hash = token_hash_for_check
-- updated_at < updated_at_for_check
-- expires_at > current_timestamp
-- の条件をすべて満たす場合にのみ更新されます。条件を満たさない場合は、pgx.ErrNoRowsが返されます。
-- リフレッシュトークンに紐づくuser_authorization_idから、それに紐づくuserのpublic_idを返します。
-- name: UpdateRefreshToken :one
WITH updated AS (
    UPDATE
        m_refresh_token
    SET
        token_hash = @new_token_hash,
        expires_at = @new_expires_at,
        ip_address = @new_ip_address,
        device_fingerprint = @new_device_fingerprint,
        user_agent = @new_user_agent,
        country_code = @new_country_code,
        city_name = @new_city_name,
        browser_name = @new_browser_name,
        device_type = @new_device_type,
        updated_at = current_timestamp,
        generation = generation + 1,
        access_token_jti = @new_access_token_jti,
        last_logged_in_at = @last_logged_in_at
    WHERE
        token_hash = @token_hash_for_check -- NOTE: token_hashにunique indexがあるため、updated_at, expires_atにインデックスは張らなくてもよい
        AND m_refresh_token.updated_at < @updated_at_for_check
        AND expires_at > current_timestamp
    RETURNING
        m_user_authorization_id
)
SELECT
    m_user.public_id
FROM
    m_user, updated
WHERE
    m_user.m_user_authorization_id = updated.m_user_authorization_id
LIMIT
    1;

-- userテーブルのpublic_idから、そのリフレッシュトークンの一覧を返します。
-- name: GetRefreshTokens :many
SELECT
    m_refresh_token.public_id,
    m_refresh_token.activated_at,
    m_refresh_token.last_logged_in_at,
    m_refresh_token.user_agent,
    m_refresh_token.device_fingerprint,
    m_refresh_token.ip_address,
    m_refresh_token.country_code,
    m_refresh_token.city_name,
    m_refresh_token.token_hash,
    m_refresh_token.expires_at
FROM
    m_refresh_token
    INNER JOIN m_user ON m_user.m_user_authorization_id = m_refresh_token.m_user_authorization_id
WHERE
    m_user.public_id = $1
ORDER BY
    m_refresh_token.created_at DESC
LIMIT
    $2 OFFSET $3;

-- m_refresh_tokenのpublic_idから、そのレコードを削除します。
-- 削除されたレコードに紐づくjtiをブラックリストに保存します。
-- 削除されたレコードのpublic_idが返されます。
-- name: RemoveRefreshTokenByIDAndSaveJtiBlacklist :one
WITH deleted AS (
    DELETE FROM
        m_refresh_token USING m_user
    WHERE
        m_user.public_id = @user_public_id
        AND m_user.m_user_authorization_id = m_refresh_token.m_user_authorization_id
        AND m_refresh_token.public_id = @refresh_token_public_id
    RETURNING
        m_refresh_token.public_id,
        m_refresh_token.access_token_jti
),
inserted AS (
    INSERT INTO
        t_jti_blacklist (jti, expires_at)
    SELECT
        access_token_jti,
        @expires_at
    FROM
        deleted
    ON CONFLICT DO NOTHING
    RETURNING
        jti
)
SELECT
    deleted.public_id
FROM
    deleted
LIMIT
    1;

-- m_refresh_tokenのtoken_hashから、そのレコードを削除します。
-- 削除されたレコードに紐づくjtiをブラックリストに保存します。
-- 削除されたレコードのpublic_idが返されます。
-- name: RemoveRefreshTokenByTokenHashAndSaveJtiBlacklist :one
WITH deleted AS (
    DELETE FROM
        m_refresh_token USING m_user
    WHERE
        m_user.public_id = @user_public_id
        AND m_user.m_user_authorization_id = m_refresh_token.m_user_authorization_id
        AND m_refresh_token.token_hash = @token_hash
    RETURNING
        m_refresh_token.public_id,
        m_refresh_token.access_token_jti
),
inserted AS (
    INSERT INTO
        t_jti_blacklist (jti, expires_at)
    SELECT
        access_token_jti,
        @expires_at
    FROM
        deleted
    ON CONFLICT DO NOTHING
    RETURNING
        jti
)
SELECT
    deleted.public_id
FROM
    deleted
LIMIT
    1;

-- jtiのブラックリストに追加する。
-- jtiに一意制約があり、重複する場合は何もしない。
-- name: SaveJTIBlacklist :exec
INSERT INTO
    t_jti_blacklist (jti, expires_at)
VALUES
    ($1, $2)
ON CONFLICT DO NOTHING;

-- jtiがブラックリストに存在するか確認する。
-- jtiが存在しない場合はpgx.ErrNoRowsが返される。
-- jtiが存在する場合は、そのexpires_atが返される。
-- name: IsJTIBlacklisted :one
SELECT
    expires_at
FROM
    t_jti_blacklist
WHERE
    jti = $1
LIMIT
    1;

-- expires_atが過ぎたjtiのブラックリストを削除します。
-- name: CleanupExpiredJTIBlacklist :exec
DELETE FROM
    t_jti_blacklist
WHERE
    expires_at < $1;
