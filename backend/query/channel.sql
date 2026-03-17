-- name: GetChannelByIdOrHandle :one
SELECT
    m_channel.m_channel_id,
    m_channel.public_id,
    m_channel.external_id,
    m_channel.external_custom_id,
    m_channel.external_display_name,
    m_channel.external_description,
    m_channel.external_icon_url,
    m_channel.external_subscribers_count,
    m_channel.external_created_at
FROM
    m_channel
WHERE
    m_channel.external_id = @external_id
    OR m_channel.external_custom_id = @external_custom_id
LIMIT
    1;

-- name: SaveChannel :one
INSERT INTO
    m_channel (
        external_id,
        external_display_name,
        external_custom_id,
        external_icon_url,
        external_description,
        external_subscribers_count,
        external_created_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    m_channel.m_channel_id,
    m_channel.public_id,
    m_channel.created_at;

-- name: SaveChannelSubscription :one
WITH subscriber AS (
    SELECT
        m_user.m_user_id
    FROM
        m_user
    WHERE
        m_user.public_id = @user_public_id
    LIMIT
        1
)
INSERT INTO
    m_user_subscribing_channel (m_user_id, m_channel_id)
SELECT
    subscriber.m_user_id AS m_user_id,
    @channel_id AS m_channel_id
FROM
    subscriber
RETURNING
    m_user_subscribing_channel.public_id,
    m_user_subscribing_channel.created_at;

-- name: GetChannelSubscriptions :many
SELECT
    m_user_subscribing_channel.public_id,
    m_user_subscribing_channel.created_at,
    m_channel.public_id AS channel_public_id,
    m_channel.external_id,
    m_channel.external_display_name,
    m_channel.external_description,
    m_channel.external_custom_id,
    m_channel.external_icon_url,
    m_channel.external_subscribers_count,
    m_channel.external_created_at
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
        OR m_user_subscribing_channel.public_id < sqlc.narg('cursor_public_id') :: uuid
    )
ORDER BY
    m_user_subscribing_channel.m_user_id,
    m_user_subscribing_channel.public_id DESC
LIMIT
    @query_limit;

-- name: DeleteChannelSubscription :execrows
DELETE FROM
    m_user_subscribing_channel
WHERE
    m_user_subscribing_channel.public_id = @subscription_public_id
    AND m_user_subscribing_channel.m_user_id = (
        SELECT
            m_user.m_user_id
        FROM
            m_user
        WHERE
            m_user.public_id = @user_public_id
        LIMIT
            1
    );

-- name: GetChannelRSSFetchedAtForUpdate :one
SELECT
    m_channel.m_channel_id,
    m_channel.rss_fetched_at,
    m_channel.external_id
FROM
    m_channel
WHERE
    m_channel.public_id = @channel_id
LIMIT
    1
FOR UPDATE;

-- name: GetChannelsToFetchRSSForUpdate :many
SELECT
    c.external_id,
    c.m_channel_id
FROM
    m_channel c
INNER JOIN
    m_user_subscribing_channel sub
ON
    c.m_channel_id = sub.m_channel_id
WHERE
    sub.m_user_id = (SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id) AND
    c.rss_fetched_at < @rss_fetch
FOR UPDATE;

-- name: MarkChannelRSSAsFetched :one
UPDATE
    m_channel
SET
    rss_fetched_at = CURRENT_TIMESTAMP
WHERE
    m_channel_id = ANY(@m_channel_ids::bigint[])
RETURNING
    public_id;
