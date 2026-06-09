-- +goose Up
ALTER TABLE m_playlist ADD COLUMN external_id varchar(64) NOT NULL DEFAULT '';
UPDATE m_playlist SET external_id = public_id::text;
ALTER TABLE m_playlist ALTER COLUMN external_id DROP DEFAULT;
CREATE UNIQUE INDEX uk_2_m_playlist ON m_playlist (external_id);
DELETE FROM m_playlist WHERE playlist_code = 1;
UPDATE m_user SET recent_playlist_ids = '{}';

-- +goose Down
DROP INDEX uk_2_m_playlist;
ALTER TABLE m_playlist DROP COLUMN external_id;
