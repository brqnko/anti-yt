-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_4_t_video_watch ON t_video_watch (m_user_id, public_id DESC);
CREATE INDEX idx_1_h_user ON h_user (m_user_authorization_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_1_h_user;
DROP INDEX idx_4_t_video_watch;
-- +goose StatementEnd
