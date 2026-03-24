-- name: ListValuableChannels :many
SELECT
    m_channel.public_id AS channel_public_id,
    m_channel.external_custom_id,
    m_channel.external_display_name,
    m_channel.external_icon_url,
    m_valuable_channel.category_code,
    m_valuable_channel.valuable_description
FROM
    m_valuable_channel
    INNER JOIN m_channel ON m_channel.m_channel_id = m_valuable_channel.m_channel_id
ORDER BY
    m_valuable_channel.m_valuable_channel_id DESC;

-- name: UpsertValuableChannel :one
INSERT INTO
    m_valuable_channel (m_channel_id, category_code, valuable_description)
SELECT
    m_channel.m_channel_id, @category_code, @valuable_description
FROM
    m_channel
WHERE
    m_channel.public_id = @channel_public_id
LIMIT 1
ON CONFLICT (m_channel_id) DO UPDATE SET
    category_code = EXCLUDED.category_code,
    valuable_description = EXCLUDED.valuable_description,
    updated_at = current_timestamp
RETURNING
    m_valuable_channel_id;

-- name: DeleteValuableChannel :exec
DELETE FROM
    m_valuable_channel
USING
    m_channel
WHERE
    m_channel.m_channel_id = m_valuable_channel.m_channel_id
    AND m_channel.public_id = @channel_public_id;

-- name: GetValuableChannelForUpdate :one
SELECT
    m_channel.public_id AS channel_public_id,
    m_valuable_channel.category_code,
    m_valuable_channel.valuable_description
FROM
    m_valuable_channel
    INNER JOIN m_channel ON m_channel.m_channel_id = m_valuable_channel.m_channel_id
WHERE
    m_channel.public_id = @channel_public_id
LIMIT 1
FOR UPDATE OF m_valuable_channel;
