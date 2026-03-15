-- +goose Up
ALTER TABLE m_channel ADD COLUMN rss_fetched_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL;

ALTER TABLE m_video ADD COLUMN external_thumbnail_url VARCHAR(128) DEFAULT '' NOT NULL;

-- +goose Down
ALTER TABLE m_video DROP COLUMN external_thumbnail_url;

ALTER TABLE m_channel DROP COLUMN rss_fetched_at;
