-- +goose Up
-- +goose StatementBegin
DROP INDEX idx_1_m_channel;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX idx_1_m_channel ON m_channel (rss_fetched_at);
-- +goose StatementEnd
