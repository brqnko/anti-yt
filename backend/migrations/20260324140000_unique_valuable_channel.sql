-- +goose Up
DROP INDEX uk_1_m_valuable_channel;
CREATE UNIQUE INDEX uk_1_m_valuable_channel ON m_valuable_channel (m_channel_id);

-- +goose Down
DROP INDEX uk_1_m_valuable_channel;
CREATE INDEX uk_1_m_valuable_channel ON m_valuable_channel (m_channel_id);
