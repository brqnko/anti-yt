-- ユーザーを作成する。
-- name: InsertUser :one
INSERT INTO
    m_user (
        m_user_authorization_id,
        display_name,
        language_code,
        daily_screen_time_seconds,
        joined_at,
        public_id,
        recent_playlist_ids
    )
SELECT
    m_user_authorization.m_user_authorization_id,
    @display_name,
    @language_code,
    @daily_screen_time_seconds,
    @joined_at,
    @public_id,
    '{}'::bigint[]
FROM
    m_user_authorization
WHERE
    m_user_authorization.public_id = @user_authorization_public_id
LIMIT
    1
RETURNING
    m_user_id;

-- ユーザーを更新する。
-- name: UpdateUser :one
UPDATE
    m_user
SET
    display_name = @display_name,
    language_code = @language_code,
    daily_screen_time_seconds = @daily_screen_time_seconds,
    updated_at = CURRENT_TIMESTAMP
WHERE
    m_user.public_id = @user_public_id
RETURNING
    m_user_id;

-- NOTE: err == nilの場合はlen(param)
-- name: BulkInsertScreenTimeRanges :copyfrom
INSERT INTO
    m_user_screen_time_range (
        m_user_id,
        screen_time_range_start,
        screen_time_range_end
    )
VALUES
    ($1, $2, $3);

-- m_user.public_idのユーザーの視聴制限範囲を取得する。
-- name: ListScreenTimeRanges :many
SELECT
    m_user_screen_time_range.screen_time_range_start,
    m_user_screen_time_range.screen_time_range_end
FROM
    m_user_screen_time_range
WHERE
    m_user_screen_time_range.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_public_id
        LIMIT
            1
    )
ORDER BY
    m_user_screen_time_range.screen_time_range_start;

-- m_user_authorization_idに紐づくh_userとm_userの数を数える
-- name: CountUsersByAuthorization :one
WITH auth AS (
    SELECT
        m_user_authorization_id
    FROM
        m_user_authorization
    WHERE
        m_user_authorization.public_id = $1
    LIMIT
        1
)
SELECT
    (
        SELECT
            COUNT(*)
        FROM
            h_user
        WHERE
            m_user_authorization_id = (
                SELECT
                    auth.m_user_authorization_id
                FROM
                    auth
            )
    ) + (
        SELECT
            COUNT(*)
        FROM
            m_user
        WHERE
            m_user_authorization_id = (
                SELECT
                    auth.m_user_authorization_id
                FROM
                    auth
            )
    ) AS total_count;

-- m_user.public_idから、ユーザーをロッキングリードする。
-- name: GetUserForUpdate :one
SELECT
    m_user.public_id,
    m_user.display_name,
    m_user.language_code,
    m_user.joined_at,
    m_user.daily_screen_time_seconds
FROM
    m_user
WHERE
    m_user.public_id = @user_public_id
LIMIT
    1
FOR UPDATE;

-- m_user.public_idから、ユーザーのプロファイルとスクリーン時間制限範囲を取得する。
-- name: GetUserProfile :many
SELECT
    m_user.display_name,
    m_user.language_code,
    m_user.joined_at,
    m_user.daily_screen_time_seconds,
    m_user_screen_time_range.screen_time_range_start,
    m_user_screen_time_range.screen_time_range_end
FROM
    m_user
    LEFT JOIN m_user_screen_time_range ON m_user_screen_time_range.m_user_id = m_user.m_user_id
WHERE
    m_user.public_id = @user_public_id
ORDER BY
    m_user_screen_time_range.screen_time_range_start;

-- m_user.m_user_idから、そのユーザーのスクリーン時間の範囲制限を削除する
-- name: DeleteScreenTimeRangesByUserID :exec
DELETE FROM
    m_user_screen_time_range
WHERE
    m_user_screen_time_range.m_user_id = $1;

-- m_userをh_userに移動します。
-- name: ArchiveUser :exec
WITH deleted AS (
    DELETE FROM
        m_user
    WHERE
        m_user.public_id = @user_public_id
    RETURNING
        m_user.m_user_id,
        m_user.m_user_authorization_id,
        m_user.display_name,
        m_user.language_code,
        m_user.daily_screen_time_seconds,
        m_user.joined_at,
        m_user.public_id
)
INSERT INTO
    h_user (
        h_user_id,
        m_user_authorization_id,
        display_name,
        language_code,
        daily_screen_time_seconds,
        joined_at,
        left_at,
        leave_reason_code,
        public_id
    )
SELECT
    deleted.m_user_id AS h_user_id,
    deleted.m_user_authorization_id AS m_user_authorization_id,
    deleted.display_name AS display_name,
    deleted.language_code AS language_code,
    deleted.daily_screen_time_seconds AS daily_screen_time_seconds,
    deleted.joined_at AS joined_at,
    CURRENT_TIMESTAMP AS left_at,
    @leave_reason_code AS leave_reason_code,
    deleted.public_id AS public_id
FROM
    deleted;

-- authorization_public_idに紐づく退会済みユーザーを削除する（再登録用）。
-- 退会済みユーザーが存在しない場合はpgx.ErrNoRowsが返される。
-- name: DeleteLeftUserByAuthorization :one
DELETE FROM h_user
WHERE h_user.m_user_authorization_id = (
    SELECT m_user_authorization.m_user_authorization_id
    FROM m_user_authorization
    WHERE m_user_authorization.public_id = @user_authorization_public_id
    LIMIT 1
)
RETURNING h_user_id;

-- 退会済みユーザーの一覧を取得する。
-- name: ListLeftUsers :many
SELECT h_user_id, m_user_authorization_id, public_id FROM h_user;

-- 退会済みユーザーとその関連データを全て削除する。
-- m_refresh_tokenはm_user_authorizationのCASCADE DELETEで自動削除される。
-- name: PurgeLeftUser :exec
WITH d_playlist_video AS (
    DELETE FROM m_playlist_video
    WHERE m_playlist_video.m_playlist_id IN (SELECT m_playlist.m_playlist_id FROM m_playlist WHERE m_playlist.m_user_id = @h_user_id)
),
d_playlist AS (
    DELETE FROM m_playlist WHERE m_playlist.m_user_id = @h_user_id
),
d_subscription AS (
    DELETE FROM m_user_subscribing_channel WHERE m_user_subscribing_channel.m_user_id = @h_user_id
),
d_screen_time AS (
    DELETE FROM m_user_screen_time_range WHERE m_user_screen_time_range.m_user_id = @h_user_id
),
d_video_watch AS (
    DELETE FROM t_video_watch WHERE t_video_watch.m_user_id = @h_user_id
),
d_video_watched AS (
    DELETE FROM t_video_watched WHERE t_video_watched.m_user_id = @h_user_id
),
d_summary AS (
    DELETE FROM s_monthly_video_watch WHERE s_monthly_video_watch.m_user_id = @h_user_id
),
d_ratelimit AS (
    DELETE FROM t_ratelimit WHERE t_ratelimit.user_public_id = @user_public_id
),
d_h_user AS (
    DELETE FROM h_user WHERE h_user.h_user_id = @h_user_id
)
DELETE FROM m_user_authorization WHERE m_user_authorization.m_user_authorization_id = @authorization_id;
