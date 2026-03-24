-- m_user.public_idから、そのユーザーが今日視聴していた合計時間(seconds)と設定している制限時間を返す。
-- その日に一本も動画を視聴していない場合は0を返します。
-- name: GetDailyWatchSummary :one
SELECT
    (
        SELECT
            m_user.daily_screen_time_seconds
        FROM
            m_user
        WHERE
            m_user.public_id = @public_id
    )::int AS daily_limit_seconds,
    COALESCE(
        EXTRACT(
            EPOCH
            FROM
                SUM(
                    LEAST(t_video_watch.watch_end_at, CURRENT_TIMESTAMP) - t_video_watch.watch_start_at
                )
        )::bigint,
        0
    )::int AS today_watch_total
FROM
    t_video_watch
WHERE
    t_video_watch.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @public_id
        LIMIT
            1
    )
    AND CURRENT_DATE <= t_video_watch.watch_start_at
    AND t_video_watch.watch_start_at < CURRENT_DATE + INTERVAL '1 day';

-- name: ListWatchHistory :many
SELECT
    m_video.public_id AS video_id,
    m_video.external_title AS external_video_title,
    m_video.external_thumbnail_url AS external_video_thumbnail_url,
    m_video.external_length_seconds AS external_video_length_seconds,
    t_video_watch.watch_position_seconds,
    t_video_watch.watch_start_at AS watched_at,
    m_channel.public_id AS channel_id,
    m_channel.external_display_name AS external_channel_display_name,
    m_channel.external_icon_url AS external_channel_icon_url
FROM
    t_video_watch
    INNER JOIN m_video ON m_video.m_video_id = t_video_watch.m_video_id
    INNER JOIN m_channel ON m_channel.m_channel_id = m_video.m_channel_id
WHERE
    t_video_watch.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_id
        LIMIT
            1
    )
    AND (
        sqlc.narg('cursor')::uuid IS NULL
        OR t_video_watch.public_id < sqlc.narg('cursor')::uuid
    )
ORDER BY
    t_video_watch.public_id DESC
LIMIT
    @query_limit;

-- name: UpsertWatchHeartbeat :exec
WITH resolved_user AS (
    SELECT
        m_user_id
    FROM
        m_user
    WHERE
        m_user.public_id = @user_public_id
    LIMIT
        1
),
resolved_video AS (
    SELECT
        m_video_id
    FROM
        m_video
    WHERE
        m_video.public_id = @video_public_id
    LIMIT
        1
),
close_stale AS (
    -- heartbeatが途絶えた放置セッションをクローズ
    UPDATE
        t_video_watch
    SET
        watch_end_at = t_video_watch.updated_at,
        updated_at = CURRENT_TIMESTAMP
    WHERE
        m_user_id = (SELECT m_user_id FROM resolved_user)
        AND watch_end_at > CURRENT_TIMESTAMP
        AND t_video_watch.updated_at < CURRENT_TIMESTAMP - INTERVAL '2 minutes'
    RETURNING
        t_video_watch_id
),
latest_active AS (
    -- ユーザーの現在アクティブなレコードを取得（タイムアウトしていないもの）
    SELECT
        t_video_watch_id,
        m_video_id
    FROM
        t_video_watch
    WHERE
        m_user_id = (SELECT m_user_id FROM resolved_user)
        AND watch_end_at > CURRENT_TIMESTAMP
        AND updated_at > CURRENT_TIMESTAMP - INTERVAL '2 minutes'
    ORDER BY
        watch_start_at DESC
    LIMIT
        1
),
close_old AS (
    -- 別の動画なら終了させる
    UPDATE
        t_video_watch
    SET
        watch_end_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    FROM
        latest_active
    WHERE
        t_video_watch.t_video_watch_id = latest_active.t_video_watch_id
        AND latest_active.m_video_id != (SELECT m_video_id FROM resolved_video)
    RETURNING
        t_video_watch.t_video_watch_id
),
update_same AS (
    -- 同じ動画ならポジションを更新
    UPDATE
        t_video_watch
    SET
        watch_position_seconds = @watch_position_seconds,
        updated_at = CURRENT_TIMESTAMP
    FROM
        latest_active
    WHERE
        t_video_watch.t_video_watch_id = latest_active.t_video_watch_id
        AND latest_active.m_video_id = (SELECT m_video_id FROM resolved_video)
    RETURNING
        t_video_watch.t_video_watch_id
),
do_insert AS (
    -- update_same が空 = 同じ動画のアクティブセッションがない -> 新規INSERT
    INSERT INTO
        t_video_watch (
            m_user_id,
            m_video_id,
            watch_position_seconds
        )
    SELECT
        (SELECT m_user_id FROM resolved_user),
        (SELECT m_video_id FROM resolved_video),
        @watch_position_seconds
    WHERE
        NOT EXISTS (SELECT 1 FROM update_same)
    RETURNING
        t_video_watch_id
)
SELECT 1;

-- name: ListDailyWatchStatsByRange :many
SELECT
    DATE(video_watch.watch_start_at) AS watch_date,
    COUNT(DISTINCT video_watch.m_video_id) AS video_count,
    EXTRACT(EPOCH FROM SUM(video_watch.watch_end_at - video_watch.watch_start_at))::bigint AS watch_sum
FROM
    t_video_watch video_watch
WHERE
    video_watch.m_user_id = (
        SELECT
            u.m_user_id
        FROM
            m_user u
        WHERE
            u.public_id = @user_id
        LIMIT
            1
    )
    AND video_watch.watch_start_at BETWEEN @start_date AND @end_date
    AND video_watch.watch_end_at <= CURRENT_TIMESTAMP
GROUP BY
    DATE(video_watch.watch_start_at);
