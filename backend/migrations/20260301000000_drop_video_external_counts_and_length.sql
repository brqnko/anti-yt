-- +goose Up
-- +goose StatementBegin

ALTER TABLE m_video DROP COLUMN external_like_count;
ALTER TABLE m_video DROP COLUMN external_watch_count;
ALTER TABLE m_video DROP COLUMN external_length_seconds;
ALTER TABLE m_playlist DROP COLUMN playlist_total_video_length_seconds;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE m_playlist ADD COLUMN playlist_total_video_length_seconds integer NOT NULL DEFAULT 0;
ALTER TABLE m_video ADD COLUMN external_length_seconds integer NOT NULL DEFAULT 0;
ALTER TABLE m_video ADD COLUMN external_watch_count bigint NOT NULL DEFAULT 0;
ALTER TABLE m_video ADD COLUMN external_like_count bigint NOT NULL DEFAULT 0;

-- +goose StatementEnd
