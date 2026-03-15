-- +goose Up
CREATE INDEX idx_2_m_channel ON m_channel(fetched_at);

-- +goose Down
DROP INDEX idx_2_m_channel;
