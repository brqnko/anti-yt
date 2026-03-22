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
    m_playlist.m_playlist_id,
    m_playlist.public_id,
    m_playlist.created_at,
    m_playlist.updated_at;

-- name: UpdatePlaylist :one
UPDATE
    m_playlist playlist
SET
    playlist_title = COALESCE(sqlc.narg('new_playlist_title'), playlist.playlist_title),
    playlist_description = COALESCE(sqlc.narg('new_playlist_description'), playlist.playlist_description),
    updated_at = CURRENT_TIMESTAMP
WHERE
    playlist.m_user_id = (
        SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id LIMIT 1
    ) AND
    playlist.public_id = @playlist_id
RETURNING playlist.public_id;

-- name: GetPlaylist :one
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.created_at,
    (
        SELECT
            COUNT(*)
        FROM
            m_playlist_video playlist_video
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
    )::int AS video_count,
    COALESCE((
        SELECT
            video.external_thumbnail_url
        FROM
            m_playlist_video playlist_video
        INNER JOIN
            m_video video
        ON
            playlist_video.m_video_id = video.m_video_id
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
        LIMIT 1
    ), '')::varchar AS top_thumbnail
FROM
    m_playlist playlist
WHERE
    playlist.m_user_id = (
        SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id LIMIT 1
    ) AND
    playlist.public_id = @playlist_id;

-- name: GetUserPlaylists :many
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.created_at,
    (
        SELECT
            COUNT(*)
        FROM
            m_playlist_video playlist_video
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
    )::int AS video_count,
    COALESCE((
        SELECT
            video.external_thumbnail_url
        FROM
            m_playlist_video playlist_video
        INNER JOIN
            m_video video
        ON
            playlist_video.m_video_id = video.m_video_id
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
        LIMIT 1
    ), '')::varchar AS top_thumbnail
FROM
    m_playlist playlist
WHERE
    playlist.m_user_id = (
        SELECT
            u.m_user_id
        FROM
            m_user u
        WHERE
            u.public_id = @user_id
        LIMIT 1
    ) AND
    (
        sqlc.narg('cursor')::uuid IS NULL OR
        playlist.public_id < sqlc.narg('cursor')::uuid
    )
ORDER BY
    playlist.m_user_id, playlist.public_id DESC
LIMIT @query_limit;

-- name: DeletePlaylist :one
WITH deleted AS (
    DELETE FROM
        m_playlist playlist -- NOTE: cascadeつけてるのでm_playlist_videoは勝手に消えてくれる
    WHERE
        playlist.m_user_id = (
            SELECT
                u.m_user_id
            FROM
                m_user u
            WHERE
                u.public_id = @user_id
            LIMIT 1
        ) AND
        playlist.public_id = @playlist_id
    RETURNING playlist.public_id
)
SELECT
    deleted.public_id
FROM
    deleted;

-- name: InsertIntoPlaylist :one
WITH inserted AS (
    INSERT INTO
        m_playlist_video (
            m_playlist_id,
            m_video_id,
            playlist_position
        )
    SELECT
        -- m_playlist_id
        (
            SELECT
                m_playlist.m_playlist_id
            FROM
                m_playlist
            WHERE
                m_playlist.m_user_id = (
                    SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_id LIMIT 1
                ) AND
                m_playlist.public_id = @playlist_id
            LIMIT 1
        ),
        -- m_video_id
        (
            SELECT
                video.m_video_id
            FROM
                m_video video
            WHERE
                video.public_id = @video_id
            LIMIT 1
        ),
        -- playlist_position
        COALESCE(
            (
                SELECT
                    MAX(m_playlist_video.playlist_position) + 1048576 -- NOTE: 2^20
                FROM
                    m_playlist_video
                WHERE
                    m_playlist_video.m_playlist_id = (
                        SELECT
                            m_playlist.m_playlist_id
                        FROM
                            m_playlist
                        WHERE
                            m_playlist.m_user_id = (
                                SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_id LIMIT 1
                            ) AND
                            m_playlist.public_id = @playlist_id
                        LIMIT 1
                    )
            ),
            0
        )
    RETURNING m_playlist_id
)
UPDATE
    m_playlist playlist
SET
    updated_at = CURRENT_TIMESTAMP
WHERE
    EXISTS (SELECT 1 FROM inserted) AND
    playlist.m_user_id = (
        SELECT
            u.m_user_id
        FROM
            m_user u
        WHERE
            u.public_id = @user_id
        LIMIT 1
    ) AND
    playlist.public_id = @playlist_id
RETURNING playlist.public_id, playlist.updated_at;

-- name: RemoveVideoFromPlaylist :one
WITH removed AS (
    DELETE FROM
        m_playlist_video playlist_video
    WHERE
        playlist_video.m_playlist_id = (
            SELECT playlist.m_playlist_id FROM m_playlist playlist WHERE playlist.m_user_id = (SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id LIMIT 1) AND playlist.public_id = @playlist_id LIMIT 1
        ) AND
        playlist_video.m_video_id = (
            SELECT video.m_video_id FROM m_video video WHERE video.public_id = @video_id LIMIT 1
        )
    RETURNING playlist_video.m_playlist_video_id
)
UPDATE
    m_playlist playlist
SET
    updated_at = CURRENT_TIMESTAMP
WHERE
    EXISTS (SELECT 1 FROM removed) AND
    playlist.public_id = @playlist_id
RETURNING playlist.public_id;

-- name: GetPlaylistVideos :many
SELECT
    video.public_id,
    video.external_thumbnail_url,
    video.external_title,
    video.external_created_at,
    video.external_length_seconds,
    COALESCE((
        SELECT
            t_video_watch.watch_position_seconds
        FROM
            t_video_watch
        WHERE
            t_video_watch.m_user_id = (
                SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_id LIMIT 1
            )
            AND t_video_watch.m_video_id = video.m_video_id
        ORDER BY
            t_video_watch.m_user_id,
            t_video_watch.m_video_id,
            t_video_watch.watch_start_at DESC
        LIMIT 1
    ), 0)::int AS last_watch_seconds,
    channel.public_id AS channel_id,
    channel.external_icon_url AS external_channel_icon_url,
    channel.external_display_name AS external_channel_displayname
FROM
    m_playlist_video playlist_video
INNER JOIN
    m_video video
ON
    video.m_video_id = playlist_video.m_video_id
INNER JOIN
    m_channel channel
ON
    channel.m_channel_id = video.m_channel_id
WHERE
    playlist_video.m_playlist_id = (
        SELECT playlist.m_playlist_id FROM m_playlist playlist
        WHERE playlist.m_user_id = (SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id LIMIT 1) AND
        playlist.public_id = @playlist_id
        LIMIT 1
    ) AND (
        sqlc.narg('cursor')::uuid IS NULL OR
        playlist_video.playlist_position > (
            SELECT pv.playlist_position FROM m_playlist_video pv
            INNER JOIN m_video v ON v.m_video_id = pv.m_video_id
            WHERE v.public_id = sqlc.narg('cursor')::uuid AND
            pv.m_playlist_id = (
                SELECT playlist.m_playlist_id FROM m_playlist playlist
                WHERE playlist.m_user_id = (SELECT u.m_user_id FROM m_user u WHERE u.public_id = @user_id LIMIT 1) AND
                playlist.public_id = @playlist_id
                LIMIT 1
            )
            LIMIT 1
        )
    )
ORDER BY
    playlist_video.m_playlist_id, playlist_video.playlist_position
LIMIT @query_limit;

-- name: BulkInsertIntoPlaylist :copyfrom
INSERT INTO
    m_playlist_video (
        m_playlist_id,
        m_video_id,
        playlist_position
    )
VALUES
    ($1, $2, $3);