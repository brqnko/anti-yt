-- +goose Up
ALTER TABLE m_channel ADD COLUMN external_uploads_playlist_id VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE m_channel ALTER COLUMN external_uploads_playlist_id DROP DEFAULT;

CREATE INDEX idx_1_m_channel ON m_channel(rss_fetched_at);

-- +goose Down
DROP INDEX idx_1_m_channel;

ALTER TABLE m_channel DROP COLUMN external_uploads_playlist_id;
