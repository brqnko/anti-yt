-- +goose Up
ALTER TABLE m_refresh_token ADD COLUMN activated_at timestamptz DEFAULT current_timestamp NOT NULL;
ALTER TABLE m_refresh_token ALTER COLUMN activated_at DROP DEFAULT;
ALTER TABLE m_refresh_token ADD COLUMN last_logged_in_at timestamptz DEFAULT current_timestamp NOT NULL;
ALTER TABLE m_refresh_token ALTER COLUMN last_logged_in_at DROP DEFAULT;

ALTER TABLE m_user_subscribing_channel ALTER COLUMN subscribed_at DROP DEFAULT;

CREATE INDEX idx_3_m_refresh_token ON m_refresh_token (m_user_authorization_id, public_id);

DROP INDEX idx_2_m_user_subscribing_channel;
DROP INDEX uk_2_m_user_subscribing_channel;
ALTER TABLE m_user_subscribing_channel DROP COLUMN public_id;

ALTER TABLE m_playlist ADD COLUMN registered_at timestamptz DEFAULT current_timestamp NOT NULL;
ALTER TABLE m_playlist ALTER COLUMN registered_at DROP DEFAULT;

DROP INDEX uk_1_m_playlist_video;
CREATE UNIQUE INDEX uk_1_m_playlist_video ON m_playlist_video (m_playlist_id, m_video_id, playlist_position);

-- +goose Down
DROP INDEX uk_1_m_playlist_video;
CREATE UNIQUE INDEX uk_1_m_playlist_video ON m_playlist_video (m_playlist_id, m_video_id);

ALTER TABLE m_playlist DROP COLUMN registered_at;
ALTER TABLE m_user_subscribing_channel ADD COLUMN public_id uuid NOT NULL DEFAULT uuidv7();
CREATE UNIQUE INDEX uk_2_m_user_subscribing_channel ON m_user_subscribing_channel (public_id);
CREATE INDEX idx_2_m_user_subscribing_channel ON m_user_subscribing_channel (m_user_id, public_id);

DROP INDEX idx_3_m_refresh_token;

ALTER TABLE m_user_subscribing_channel ALTER COLUMN subscribed_at ADD DEFAULT current_timestamp;

ALTER TABLE m_refresh_token DROP COLUMN last_logged_in_at;
ALTER TABLE m_refresh_token DROP COLUMN activated_at;
