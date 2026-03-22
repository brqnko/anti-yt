-- +goose Up
CREATE INDEX idx_2_m_playlist ON m_playlist (m_user_id, public_id DESC);

ALTER TABLE
    m_playlist_video
ADD CONSTRAINT
    fk_1_m_playlist_video
    FOREIGN KEY (m_playlist_id) REFERENCES m_playlist (m_playlist_id)
    ON DELETE CASCADE;

CREATE INDEX idx_1_m_playlist_video ON m_playlist_video (m_playlist_id, playlist_position);

-- +goose Down
DROP INDEX idx_1_m_playlist_video;

ALTER TABLE m_playlist_video DROP CONSTRAINT fk_1_m_playlist_video;

DROP INDEX idx_2_m_playlist;
