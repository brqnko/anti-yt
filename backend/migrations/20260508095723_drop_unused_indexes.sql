-- +goose Up
-- +goose StatementBegin
DROP INDEX idx_1_m_refresh_token;
DROP INDEX uk_1_s_monthly_video_watch;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX idx_1_m_refresh_token ON m_refresh_token (expires_at);
CREATE UNIQUE INDEX uk_1_s_monthly_video_watch ON s_monthly_video_watch (public_id);
-- +goose StatementEnd
