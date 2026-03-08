-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_1_m_user ON m_user (m_user_authorization_id);
CREATE INDEX idx_2_m_refresh_token ON m_refresh_token (m_user_authorization_id, updated_at);

ALTER TABLE m_refresh_token DROP COLUMN access_token_hash;

DELETE FROM m_refresh_token;
ALTER TABLE m_refresh_token ADD COLUMN access_token_jti UUID NOT NULL;

ALTER TABLE t_video_watch DROP CONSTRAINT excl_1_t_video_watch;

CREATE INDEX idx_2_t_video_watch ON t_video_watch (m_user_id, watch_start_at);

ALTER TABLE m_playlist ADD COLUMN playlist_description VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE m_playlist ALTER COLUMN playlist_description DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_playlist DROP COLUMN playlist_description;

DROP INDEX idx_2_t_video_watch;

ALTER TABLE t_video_watch ADD CONSTRAINT excl_1_t_video_watch EXCLUDE USING gist (
    m_user_id WITH =,
    m_video_id WITH =,
    tstzrange(watch_start_at, watch_end_at) WITH &&
);

ALTER TABLE m_refresh_token DROP COLUMN access_token_jti;

ALTER TABLE m_refresh_token
    ADD COLUMN access_token_hash VARCHAR(256) NOT NULL DEFAULT '';
ALTER TABLE m_refresh_token
    ALTER COLUMN access_token_hash DROP DEFAULT;

DROP INDEX idx_2_m_refresh_token;
DROP INDEX idx_1_m_user;
-- +goose StatementEnd
