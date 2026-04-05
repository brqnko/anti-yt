-- +goose Up
-- +goose StatementBegin
INSERT INTO m_playlist (m_user_id, visibility_code, playlist_title, playlist_description, playlist_code, video_count, public_id, registered_at, m_channel_id)
SELECT m_user_id, 0, 'Watch Later', '', 2, 0, uuidv7(), NOW(), 0
FROM m_user
WHERE NOT EXISTS (
    SELECT 1 FROM m_playlist
    WHERE m_playlist.m_user_id = m_user.m_user_id
    AND m_playlist.playlist_code = 2
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM m_playlist WHERE playlist_code = 2;
-- +goose StatementEnd
