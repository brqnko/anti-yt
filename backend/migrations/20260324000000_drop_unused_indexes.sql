-- +goose Up
-- +goose StatementBegin

-- 冗長: uk_1(m_user_id, m_channel_id) でカバー済み
DROP INDEX idx_1_m_user_subscribing_channel;

-- 冗長: idx_2(m_channel_id, public_id), idx_3(m_channel_id, external_created_at DESC, public_id DESC) でカバー済み
DROP INDEX idx_1_m_video;

-- 冗長: idx_2(m_user_id, watch_start_at), idx_3(m_user_id, m_video_id, watch_start_at) でカバー済み
DROP INDEX idx_1_t_video_watch;

-- 未使用: fetched_atでフィルタ/ソートするクエリなし
DROP INDEX idx_2_m_channel;

-- 未使用: idx_3(m_user_authorization_id, public_id) が実際のクエリをカバー
DROP INDEX idx_2_m_refresh_token;

-- 未使用: h_searchテーブルがquery/で未参照
DROP INDEX idx_1_h_search;

-- 未使用: idx_2(m_user_id, public_id DESC) が全playlistクエリをカバー
DROP INDEX idx_1_m_playlist;

-- 未使用: idx_3(m_channel_id, external_created_at DESC, public_id DESC) が全クエリをカバー
DROP INDEX idx_2_m_video;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX idx_1_m_user_subscribing_channel ON m_user_subscribing_channel (m_user_id);
CREATE INDEX idx_1_m_video ON m_video (m_channel_id);
CREATE INDEX idx_1_t_video_watch ON t_video_watch (m_user_id);
CREATE INDEX idx_2_m_channel ON m_channel (fetched_at);
CREATE INDEX idx_2_m_refresh_token ON m_refresh_token (m_user_authorization_id, updated_at);
CREATE INDEX idx_1_h_search ON h_search (m_user_id);
CREATE INDEX idx_1_m_playlist ON m_playlist (m_user_id, visibility_code, playlist_code, created_at);
CREATE INDEX idx_2_m_video ON m_video (m_channel_id, public_id);
-- +goose StatementEnd
