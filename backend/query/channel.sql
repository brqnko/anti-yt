-- name: FindChannelByExternalID :one
SELECT
    m_channel.m_channel_id,
    m_channel.public_id,
    m_channel.external_id,
    m_channel.external_custom_id,
    m_channel.external_display_name,
    m_channel.external_description,
    m_channel.external_icon_url,
    m_channel.external_subscribers_count,
    m_channel.external_created_at,
    m_channel.external_uploads_playlist_id,
    m_channel.fetched_at,
    m_channel.rss_fetched_at,
    m_channel.bulk_fetched_at
FROM
    m_channel
WHERE
    m_channel.external_id = @external_id
    OR m_channel.external_custom_id = @external_custom_id
LIMIT
    1;

-- NOTE: SaveChannelの前にこれをする. CTEでやろうと思ったけど、トランザクション...
-- name: ClearStaleChannelCustomID :exec
UPDATE
    m_channel
SET
    external_custom_id = '@' || external_id
WHERE
    external_custom_id = @external_custom_id
    AND external_id != @external_id;

-- name: UpsertChannel :one
INSERT INTO
    m_channel (
        external_id,
        external_display_name,
        external_custom_id,
        external_icon_url,
        external_description,
        external_subscribers_count,
        external_created_at,
        external_uploads_playlist_id,
        public_id,
        rss_fetched_at,
        fetched_at,
        bulk_fetched_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (external_id) DO UPDATE SET
    external_display_name = EXCLUDED.external_display_name,
    external_custom_id = EXCLUDED.external_custom_id,
    external_icon_url = EXCLUDED.external_icon_url,
    external_description = EXCLUDED.external_description,
    external_subscribers_count = EXCLUDED.external_subscribers_count,
    external_created_at = EXCLUDED.external_created_at,
    external_uploads_playlist_id = EXCLUDED.external_uploads_playlist_id,
    updated_at = CURRENT_TIMESTAMP,
    rss_fetched_at = EXCLUDED.rss_fetched_at,
    fetched_at = EXCLUDED.fetched_at,
    bulk_fetched_at = EXCLUDED.bulk_fetched_at
RETURNING
    m_channel.m_channel_id,
    m_channel.public_id;

-- name: InsertSubscription :one
INSERT INTO
    m_user_subscribing_channel (m_user_id, m_channel_id, subscribed_at)
SELECT
    (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_public_id
        LIMIT
            1
    ) AS m_user_id,
    (
        SELECT
            m_channel.m_channel_id
        FROM
            m_channel
        WHERE
            m_channel.public_id = @channel_id
        LIMIT
            1
    ),
    @subscribed_at
RETURNING
    m_user_subscribing_channel.m_user_subscribing_channel_id;

-- name: ListSubscribedChannels :many
SELECT
    m_channel.public_id AS channel_public_id,
    m_channel.external_id,
    m_channel.external_custom_id,
    m_channel.external_display_name,
    m_channel.external_icon_url,
    m_channel.external_subscribers_count
FROM
    m_user_subscribing_channel
    INNER JOIN m_channel ON m_channel.m_channel_id = m_user_subscribing_channel.m_channel_id
WHERE
    m_user_subscribing_channel.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_public_id
        LIMIT
            1
    )
    AND (
        sqlc.narg('cursor_public_id') :: uuid IS NULL
        OR m_channel.public_id < sqlc.narg('cursor_public_id') :: uuid
    )
ORDER BY
    m_channel.public_id DESC
LIMIT
    @query_limit;

-- name: ListSubscribersByChannelPublicID :many
-- 指定チャンネルを購読しているユーザーのうち、指定動画をまだ視聴していないユーザーだけを返す。
-- fan-out時に視聴済み動画がfeedに再挿入されるのを防ぐ目的。
SELECT
    m_user.public_id
FROM
    m_user_subscribing_channel
    INNER JOIN m_user ON m_user.m_user_id = m_user_subscribing_channel.m_user_id
WHERE
    m_user_subscribing_channel.m_channel_id = (
        SELECT
            m_channel.m_channel_id
        FROM
            m_channel
        WHERE
            m_channel.public_id = @channel_public_id
        LIMIT
            1
    )
    AND NOT EXISTS (
        SELECT 1 FROM t_video_watched
        WHERE t_video_watched.m_user_id = m_user_subscribing_channel.m_user_id
            AND t_video_watched.m_video_id = (
                SELECT
                    m_video.m_video_id
                FROM
                    m_video
                WHERE
                    m_video.public_id = @video_public_id
                LIMIT
                    1
            )
    );

-- name: DeleteSubscription :one
DELETE FROM
    m_user_subscribing_channel
WHERE
    m_user_subscribing_channel.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_public_id
        LIMIT
            1
    )
    AND m_user_subscribing_channel.m_channel_id = (
        SELECT
            c.m_channel_id
        FROM
            m_channel c
        WHERE
            c.public_id = @channel_id
        LIMIT
            1
    )
RETURNING m_user_subscribing_channel.m_user_subscribing_channel_id;

-- name: GetChannelByPublicID :one
SELECT
    m_channel.public_id,
    m_channel.external_custom_id,
    m_channel.external_display_name,
    m_channel.external_description,
    m_channel.external_icon_url,
    m_channel.external_subscribers_count
FROM
    m_channel
WHERE
    m_channel.public_id = sqlc.narg('channel_id')::uuid
    OR m_channel.external_id = sqlc.narg('external_channel_id')
    OR m_channel.external_custom_id = sqlc.narg('external_channel_id')
LIMIT
    1;

-- name: GetChannelForUpdate :one
SELECT
    m_channel.public_id,
    m_channel.external_id,
    m_channel.external_display_name,
    m_channel.external_description,
    m_channel.external_custom_id,
    m_channel.external_icon_url,
    m_channel.external_subscribers_count,
    m_channel.external_created_at,
    m_channel.rss_fetched_at,
    m_channel.fetched_at,
    m_channel.external_uploads_playlist_id,
    m_channel.bulk_fetched_at
FROM
    m_channel
WHERE
    m_channel.public_id = @channel_id
LIMIT
    1
FOR UPDATE;

-- name: ListStaleRSSChannelsForUpdate :many
SELECT
    c.public_id,
    c.external_id,
    c.external_display_name,
    c.external_description,
    c.external_custom_id,
    c.external_icon_url,
    c.external_subscribers_count,
    c.external_created_at,
    c.external_uploads_playlist_id,
    c.rss_fetched_at,
    c.fetched_at,
    c.bulk_fetched_at
FROM
    m_channel c
    INNER JOIN m_user_subscribing_channel sub ON c.m_channel_id = sub.m_channel_id
WHERE
    sub.m_user_id = (
        SELECT
            u.m_user_id
        FROM
            m_user u
        WHERE
            u.public_id = @user_id
        LIMIT
            1
    )
    AND c.rss_fetched_at < @rss_fetch
ORDER BY
    c.rss_fetched_at
LIMIT
    @query_limit
FOR UPDATE;

-- name: ListChannelsBulkFetchedBefore :many
SELECT
    c.public_id,
    c.external_id,
    c.external_display_name,
    c.external_description,
    c.external_custom_id,
    c.external_icon_url,
    c.external_subscribers_count,
    c.external_created_at,
    c.external_uploads_playlist_id,
    c.rss_fetched_at,
    c.fetched_at,
    c.bulk_fetched_at
FROM
    m_channel c
WHERE
    c.bulk_fetched_at < @bulk_fetched_before;
