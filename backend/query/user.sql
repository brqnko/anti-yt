-- 新しいユーザーを挿入する。
-- name: SaveUser :one
INSERT INTO m_user (m_user_authorization_id, display_name, language_code, daily_screen_time_seconds)
SELECT m_user_authorization.m_user_authorization_id, @display_name, @language_code, @daily_screen_time_seconds FROM m_user_authorization
WHERE m_user_authorization.public_id = @user_authorization_public_id LIMIT 1
RETURNING m_user_id, joined_at, public_id;

-- NOTE: err == nilの場合はlen(param)
-- name: SaveUserScreenTimeRanges :copyfrom
INSERT INTO m_user_screen_time_range (m_user_id, screen_time_range_start, screen_time_range_end)
VALUES ($1, $2, $3);

-- m_user.public_idのユーザーの視聴制限範囲を取得する。
-- name: GetUserScreenTimeRanges :many
SELECT m_user_screen_time_range.screen_time_range_start, m_user_screen_time_range.screen_time_range_end, m_user_screen_time_range.public_id FROM m_user_screen_time_range
WHERE m_user_screen_time_range.m_user_id = (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_public_id)
ORDER BY m_user_screen_time_range.screen_time_range_start;

-- m_user_authorization_idに紐づくh_userとm_userの数を数える
-- name: CountUsersByAuthorization :one
WITH auth AS (
    SELECT m_user_authorization_id
    FROM m_user_authorization
    WHERE m_user_authorization.public_id = $1
    LIMIT 1
)
SELECT
    (SELECT COUNT(*) FROM h_user WHERE m_user_authorization_id = (SELECT auth.m_user_authorization_id FROM auth)) +
    (SELECT COUNT(*) FROM m_user WHERE m_user_authorization_id = (SELECT auth.m_user_authorization_id FROM auth))
    AS total_count;

-- m_user.public_idから、ユーザーのプロファイルを更新する
-- name: UpdateUserProfile :one
UPDATE m_user SET
    display_name = COALESCE(sqlc.narg('new_display_name'), m_user.display_name),
    daily_screen_time_seconds = COALESCE(sqlc.narg('new_daily_screen_time_seconds'), m_user.daily_screen_time_seconds),
    language_code = COALESCE(sqlc.narg('new_language_code'), m_user.language_code),
    updated_at = CURRENT_TIMESTAMP
WHERE m_user.public_id = @user_public_id
RETURNING m_user.m_user_id, m_user.public_id, m_user.joined_at, m_user.display_name, m_user.daily_screen_time_seconds, m_user.language_code;

-- m_user.public_idから、ユーザーのプロファイルを取得する。
-- name: GetUserProfile :one
SELECT m_user.m_user_id, m_user.joined_at, m_user.display_name, m_user.daily_screen_time_seconds, m_user.language_code FROM m_user
WHERE m_user.public_id = @user_public_id
LIMIT 1;

-- m_user.m_user_idから、そのユーザーのスクリーン時間の範囲制限を削除する
-- name: RemoveScreenTimeRangesByUserId :exec
DELETE FROM m_user_screen_time_range WHERE m_user_screen_time_range.m_user_id = $1;

-- m_userをh_userに移動します。
-- name: RemoveUser :exec
WITH deleted AS (
    DELETE FROM m_user WHERE m_user.public_id = @user_public_id
    RETURNING m_user.m_user_id, m_user.m_user_authorization_id, m_user.display_name, m_user.language_code,
    m_user.daily_screen_time_seconds, m_user.joined_at, m_user.public_id
)
INSERT INTO h_user(h_user_id, m_user_authorization_id, display_name, language_code, daily_screen_time_seconds, joined_at, leave_reason_code, public_id) SELECT
deleted.m_user_id AS h_user_id,
deleted.m_user_authorization_id AS m_user_authorization_id,
deleted.display_name AS display_name,
deleted.language_code AS language_code,
deleted.daily_screen_time_seconds AS daily_screen_time_seconds,
deleted.joined_at AS joined_at,
@leave_reason_code AS leave_reason_code,
deleted.public_id AS public_id
FROM deleted;
