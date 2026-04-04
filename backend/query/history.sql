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
    AND sqlc.arg('today_start') <= t_video_watch.watch_start_at
    AND t_video_watch.watch_start_at < sqlc.arg('today_start') + INTERVAL '1 day';

-- name: ListWatchHistory :many
SELECT
    t_video_watch.public_id AS watch_id,
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

-- name: GetLastHeartbeatForUpdate :one
SELECT
    (
        SELECT
            m_video.public_id
        FROM
            m_video
        WHERE
            m_video.m_video_id = t_video_watch.m_video_id
        LIMIT 1
    ) AS video_id,
    (
        SELECT
            m_video.external_length_seconds
        FROM
            m_video
        WHERE
            m_video.m_video_id = t_video_watch.m_video_id
        LIMIT 1
    ) AS video_length,
    t_video_watch.updated_at,
    t_video_watch.public_id,
    t_video_watch.watch_start_at,
    t_video_watch.watch_end_at,
    t_video_watch.watch_position_seconds,
    t_video_watch.t_video_watch_id
FROM
    t_video_watch
WHERE
    m_user_id = (
        SELECT
            m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_public_id
        LIMIT
            1
    )
    AND watch_end_at > CURRENT_TIMESTAMP
LIMIT 1
FOR UPDATE;

-- name: InsertHeartbeat :exec
INSERT INTO
    t_video_watch (
        m_user_id,
        m_video_id,
        public_id,
        watch_start_at,
        watch_end_at,
        watch_position_seconds
    )
VALUES (
    (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_public_id LIMIT 1),
    (SELECT m_video.m_video_id FROM m_video WHERE m_video.public_id = @video_public_id LIMIT 1),
    @public_id,
    @watch_start_at,
    @watch_end_at,
    @watch_position_seconds
);

-- name: UpdateHeartbeat :exec
UPDATE
    t_video_watch
SET
    watch_position_seconds = @watch_position_seconds,
    watch_start_at = @watch_start_at,
    watch_end_at = @watch_end_at,
    updated_at = CURRENT_TIMESTAMP
WHERE
    t_video_watch.t_video_watch_id = @t_video_watch_id;

-- name: ListDailyWatchStatsByRange :many
SELECT
    DATE(video_watch.watch_start_at + make_interval(secs => sqlc.arg('tz_offset')::int)) AS watch_date,
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
    DATE(video_watch.watch_start_at + make_interval(secs => sqlc.arg('tz_offset')::int));

-- name: GetLatestMonthlyVideoWatchSummary :one
SELECT
    s.ai_summary_title,
    s.ai_summary_description,
    s.created_at
FROM
    s_monthly_video_watch s
WHERE
    s.m_user_id = (SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id LIMIT 1)
ORDER BY
    s.target_month DESC
LIMIT 1;

-- name: MarkVideoWatched :exec
INSERT INTO
    t_video_watched (m_user_id, m_video_id)
SELECT
    (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_public_id LIMIT 1),
    (SELECT m_video.m_video_id FROM m_video WHERE m_video.public_id = @video_public_id LIMIT 1)
ON CONFLICT (m_user_id, m_video_id) DO NOTHING;

-- name: UnmarkVideoWatched :exec
DELETE FROM
    t_video_watched
WHERE
    m_user_id = (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_public_id LIMIT 1)
    AND m_video_id = (SELECT m_video.m_video_id FROM m_video WHERE m_video.public_id = @video_public_id LIMIT 1);
