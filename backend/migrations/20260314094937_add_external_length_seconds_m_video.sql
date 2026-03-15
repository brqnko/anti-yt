-- +goose Up
ALTER TABLE m_video ADD COLUMN external_length_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE m_video ALTER COLUMN external_length_seconds DROP DEFAULT;

CREATE INDEX idx_2_m_video ON m_video(m_channel_id, public_id);

CREATE INDEX idx_3_t_video_watch ON t_video_watch(m_user_id, m_video_id, watch_start_at);

CREATE INDEX idx_3_m_user_subscribing_channel ON m_user_subscribing_channel(m_channel_id);

CREATE INDEX idx_3_m_video ON m_video(m_channel_id, external_created_at DESC, public_id DESC);

-- +goose Down
DROP INDEX idx_3_m_video

DROP INDEX idx_3_m_user_subscribing_channel;

DROP INDEX idx_3_t_video_watch;

DROP INDEX idx_2_m_video;

ALTER TABLE m_video DROP COLUMN external_video_length_seconds;
