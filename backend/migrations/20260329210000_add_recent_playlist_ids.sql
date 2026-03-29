-- +goose Up
ALTER TABLE m_user ADD COLUMN recent_playlist_ids bigint[] NOT NULL DEFAULT '{}';
ALTER TABLE m_user ALTER COLUMN recent_playlist_ids DROP DEFAULT;

-- +goose Down
ALTER TABLE m_user DROP COLUMN recent_playlist_ids;
