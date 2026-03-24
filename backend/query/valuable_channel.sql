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
