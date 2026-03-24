-- ユーザーを作成する。
-- name: InsertUser :one
INSERT INTO
    m_user (
        m_user_authorization_id,
        display_name,
        language_code,
        daily_screen_time_seconds,
        joined_at,
        public_id
    )
SELECT
    m_user_authorization.m_user_authorization_id,
    @display_name,
    @language_code,
    @daily_screen_time_seconds,
    @joined_at,
    @public_id
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
    m_user_screen_time_range.screen_time_range_end,
    m_user_screen_time_range.public_id
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
    m_user_screen_time_range.public_id AS screen_time_range_id,
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
    @leave_reason_code AS leave_reason_code,
    deleted.public_id AS public_id
FROM
    deleted;
