-- name: CreatePlaylist :one
INSERT INTO
    m_playlist (
        m_user_id,
        playlist_title,
        playlist_description,
        visibility_code,
        playlist_code
    )
VALUES
    (
        (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_public_id LIMIT 1),
        @playlist_title,
        @playlist_description,
        @visibility_code,
        @playlist_code
    )
RETURNING
    m_playlist.public_id,
    m_playlist.playlist_title,
    m_playlist.playlist_description,
    m_playlist.visibility_code,
    m_playlist.playlist_code,
    m_playlist.video_count,
    m_playlist.created_at,
    m_playlist.updated_at;
