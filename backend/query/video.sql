-- name: UpsertVideo :one
INSERT INTO
    m_video (
        m_channel_id,
        external_id,
        external_title,
        external_description,
        fetched_at,
        external_created_at,
        external_thumbnail_url,
        external_length_seconds,
        public_id
    )
SELECT
    (
        SELECT
            channel.m_channel_id
        FROM
            m_channel channel
        WHERE
            channel.public_id = @channel_id
        LIMIT
            1
    ),
    @external_id,
    @external_title,
    @external_description,
    @fetched_at,
    @external_created_at,
    @external_thumbnail_url,
    @external_length_seconds,
    @id
ON CONFLICT (external_id) DO UPDATE SET
    external_title = EXCLUDED.external_title,
    external_description = EXCLUDED.external_description,
    external_thumbnail_url = EXCLUDED.external_thumbnail_url,
    external_length_seconds = EXCLUDED.external_length_seconds,
    external_created_at = EXCLUDED.external_created_at,
    fetched_at = EXCLUDED.fetched_at,
    updated_at = CURRENT_TIMESTAMP
RETURNING
    m_video_id, public_id;

-- name: ListChannelVideos :many
SELECT
    m_video.public_id,
    m_video.external_thumbnail_url,
    m_video.external_title,
    m_video.external_created_at,
    m_video.external_length_seconds,
    COALESCE((
        SELECT
            t_video_watch.watch_position_seconds
        FROM
            t_video_watch
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
            AND t_video_watch.m_video_id = m_video.m_video_id
        ORDER BY
            t_video_watch.m_user_id,
            t_video_watch.m_video_id,
            t_video_watch.watch_start_at DESC
        LIMIT
            1
    ), 0)::int AS last_watch_seconds
FROM
    m_video
    INNER JOIN m_channel ON m_channel.m_channel_id = m_video.m_channel_id
WHERE
    m_channel.public_id = @channel_id
    AND (
        sqlc.narg('cursor')::uuid IS NULL
        OR (
            m_video.external_created_at < (
                SELECT
                    mv.external_created_at
                FROM
                    m_video mv
                WHERE
                    mv.public_id = sqlc.narg('cursor')::uuid
                LIMIT
                    1
            )
        )
        OR (
            m_video.external_created_at = (
                SELECT
                    mv.external_created_at
                FROM
                    m_video mv
                WHERE
                    mv.public_id = sqlc.narg('cursor')::uuid
                LIMIT
                    1
            )
            AND m_video.public_id < sqlc.narg('cursor')::uuid
        )
    )
ORDER BY
    m_video.external_created_at DESC,
    m_video.public_id DESC
LIMIT
    @query_limit;

-- name: ListChannelVideosOlder :many
SELECT
    m_video.public_id,
    m_video.external_thumbnail_url,
    m_video.external_title,
    m_video.external_created_at,
    m_video.external_length_seconds,
    COALESCE((
        SELECT
            t_video_watch.watch_position_seconds
        FROM
            t_video_watch
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
            AND t_video_watch.m_video_id = m_video.m_video_id
        ORDER BY
            t_video_watch.m_user_id,
            t_video_watch.m_video_id,
            t_video_watch.watch_start_at DESC
        LIMIT
            1
    ), 0)::int AS last_watch_seconds
FROM
    m_video
    INNER JOIN m_channel ON m_channel.m_channel_id = m_video.m_channel_id
WHERE
    m_channel.public_id = @channel_id
    AND (
        sqlc.narg('cursor')::uuid IS NULL
        OR (
            m_video.external_created_at > (
                SELECT
                    mv.external_created_at
                FROM
                    m_video mv
                WHERE
                    mv.public_id = sqlc.narg('cursor')::uuid
                LIMIT
                    1
            )
        )
        OR (
            m_video.external_created_at = (
                SELECT
                    mv.external_created_at
                FROM
                    m_video mv
                WHERE
                    mv.public_id = sqlc.narg('cursor')::uuid
                LIMIT
                    1
            )
            AND m_video.public_id > sqlc.narg('cursor')::uuid
        )
    )
ORDER BY
    m_video.external_created_at ASC,
    m_video.public_id ASC
LIMIT
    @query_limit;

-- ユーザーが登録しているチャンネルがだしている動画を最新順(public_id)で取得する。
-- name: ListSubscriptionFeed :many
SELECT
    m_video.public_id AS video_id,
    m_video.external_thumbnail_url AS external_video_thumbnail_url,
    m_video.external_title AS external_title,
    m_video.external_created_at AS external_created_at,
    m_video.external_length_seconds AS external_length_seconds,
    COALESCE((
        SELECT
            t_video_watch.watch_position_seconds
        FROM
            t_video_watch
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
            AND t_video_watch.m_video_id = m_video.m_video_id
        ORDER BY
            t_video_watch.m_user_id,
            t_video_watch.m_video_id,
            t_video_watch.watch_start_at DESC
        LIMIT
            1
    ), 0)::int AS last_watch_seconds,
    m_channel.public_id AS channel_id,
    m_channel.external_icon_url AS external_channel_icon_url,
    m_channel.external_display_name AS external_displayname
FROM
    m_video
    INNER JOIN m_user_subscribing_channel ON m_video.m_channel_id = m_user_subscribing_channel.m_channel_id
    INNER JOIN m_channel ON m_channel.m_channel_id = m_video.m_channel_id
WHERE
    m_user_subscribing_channel.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_id
        LIMIT
            1
    )
    AND NOT EXISTS (
        SELECT 1 FROM t_video_watched
        WHERE t_video_watched.m_user_id = m_user_subscribing_channel.m_user_id
            AND t_video_watched.m_video_id = m_video.m_video_id
    )
    AND (
        -- 日付がカーソルより昔 or (日付がカーソルと同じ and public_idがカーソルより昔)
        sqlc.narg('cursor')::uuid IS NULL
        OR (
            m_video.external_created_at < (
                SELECT
                    m_video.external_created_at
                FROM
                    m_video
                WHERE
                    m_video.public_id = sqlc.narg('cursor')::uuid
                LIMIT
                    1
            )
        )
        OR (
            m_video.external_created_at = (
                SELECT
                    m_video.external_created_at
                FROM
                    m_video
                WHERE
                    m_video.public_id = sqlc.narg('cursor')::uuid
                LIMIT
                    1
            )
            AND m_video.public_id < sqlc.narg('cursor')::uuid
        )
    )
ORDER BY
    m_video.external_created_at DESC,
    m_video.public_id DESC
LIMIT
    @query_limit;

-- name: GetVideoDetail :one
SELECT
    video.public_id AS id,
    video.external_id,
    video.external_title,
    video.external_description,
    video.external_thumbnail_url,
    video.external_created_at,
    channel.public_id AS channel_id,
    channel.external_id AS channel_external_id,
    channel.external_display_name,
    channel.external_custom_id AS channel_custom_id,
    channel.external_icon_url,
    channel.external_subscribers_count
FROM
    m_video video
    INNER JOIN m_channel channel ON video.m_channel_id = channel.m_channel_id
WHERE
    video.public_id = @video_id
LIMIT
    1;
