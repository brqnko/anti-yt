-- +goose Up
ALTER TABLE m_channel ADD COLUMN last_seen_at timestamptz NOT NULL DEFAULT current_timestamp;
ALTER TABLE m_channel ALTER COLUMN last_seen_at DROP DEFAULT;
CREATE INDEX idx_3_m_channel ON m_channel(last_seen_at);

-- +goose Down
DROP INDEX idx_3_m_channel;
ALTER TABLE m_channel DROP COLUMN last_seen_at;
