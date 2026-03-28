-- +goose Up
-- +goose StatementBegin
ALTER TABLE m_playlist ADD COLUMN m_channel_id bigint NOT NULL DEFAULT 0;
ALTER TABLE m_playlist ALTER COLUMN m_channel_id DROP DEFAULT;
CREATE INDEX idx_3_m_playlist ON m_playlist (m_channel_id, public_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_3_m_playlist;
ALTER TABLE m_playlist DROP COLUMN m_channel_id;
-- +goose StatementEnd
