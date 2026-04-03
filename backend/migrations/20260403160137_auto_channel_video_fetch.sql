-- +goose Up
ALTER TABLE m_channel ADD COLUMN bulk_fetched_at timestamptz NOT NULL DEFAULT current_timestamp - INTERVAL '1 year';
ALTER TABLE m_channel ALTER COLUMN bulk_fetched_at DROP DEFAULT;
CREATE INDEX idx_2_m_channel ON m_channel(bulk_fetched_at);

-- +goose Down
ALTER TABLE m_channel DROP COLUMN bulk_fetched_at;
