-- +goose Up
-- +goose StatementBegin

-- m_video: rename length_seconds to external_length_seconds and change type to int
ALTER TABLE m_video RENAME COLUMN length_seconds TO external_length_seconds;
ALTER TABLE m_video ALTER COLUMN external_length_seconds TYPE int;

-- m_video: add external_created_at
ALTER TABLE m_video ADD COLUMN external_created_at timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00UTC';

-- w_monthly_video_watch: add target_month and fail_reason
ALTER TABLE w_monthly_video_watch ADD COLUMN target_month date NOT NULL DEFAULT '1970-01-01';
ALTER TABLE w_monthly_video_watch ADD COLUMN fail_reason VARCHAR(128) NOT NULL DEFAULT '';

-- m_playlist: recreate idx_1_m_playlist with created_at
DROP INDEX idx_1_m_playlist;
CREATE INDEX idx_1_m_playlist ON m_playlist (m_user_id, visibility_code, playlist_code, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX idx_1_m_playlist;
CREATE INDEX idx_1_m_playlist ON m_playlist (m_user_id, visibility_code, playlist_code);

ALTER TABLE w_monthly_video_watch DROP COLUMN fail_reason;
ALTER TABLE w_monthly_video_watch DROP COLUMN target_month;

ALTER TABLE m_video DROP COLUMN external_created_at;

ALTER TABLE m_video ALTER COLUMN external_length_seconds TYPE bigint;
ALTER TABLE m_video RENAME COLUMN external_length_seconds TO length_seconds;

-- +goose StatementEnd
