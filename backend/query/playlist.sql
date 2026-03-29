-- name: UpsertPlaylist :one
INSERT INTO
    m_playlist (
        m_user_id,
        m_channel_id,
        playlist_title,
        playlist_description,
        visibility_code,
        playlist_code,
        video_count,
        public_id,
        registered_at
    )
VALUES
    (
        (
            SELECT
                m_user.m_user_id
            FROM
                m_user
            WHERE
                m_user.public_id = @user_public_id
            LIMIT
                1
        ),
        COALESCE((
            SELECT
                ch.m_channel_id
            FROM
                m_channel ch
            WHERE
                ch.public_id = sqlc.narg('channel_public_id')::uuid
            LIMIT
                1
        ), 0),
        @playlist_title,
        @playlist_description,
        @visibility_code,
        @playlist_code,
        @video_count,
        @public_id,
        @registered_at
    )
ON CONFLICT (public_id) DO UPDATE SET
    playlist_title = EXCLUDED.playlist_title,
    playlist_description = EXCLUDED.playlist_description,
    visibility_code = EXCLUDED.visibility_code,
    playlist_code = EXCLUDED.playlist_code,
    video_count = EXCLUDED.video_count,
    registered_at = EXCLUDED.registered_at,
    updated_at = CURRENT_TIMESTAMP
RETURNING
    m_playlist.m_playlist_id;

-- name: UpdatePlaylist :one
UPDATE
    m_playlist playlist
SET
    playlist_title = COALESCE(sqlc.narg('new_playlist_title'), playlist.playlist_title),
    playlist_description = COALESCE(sqlc.narg('new_playlist_description'), playlist.playlist_description),
    updated_at = CURRENT_TIMESTAMP
WHERE
    playlist.m_user_id = (
        SELECT
            u.m_user_id
        FROM
            m_user u
        WHERE
            u.public_id = @user_id
        LIMIT
            1
    )
    AND playlist.public_id = @playlist_id
RETURNING
    playlist.public_id;

-- name: GetPlaylistForUpdate :one
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.video_count,
    playlist.registered_at
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
        LIMIT
            1
    )
    AND playlist.public_id = @playlist_id
LIMIT
    1
FOR UPDATE;

-- name: GetPlaylistWithThumbnail :one
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.registered_at,
    playlist.updated_at,
    playlist.video_count,
    COALESCE((
        SELECT
            video.external_thumbnail_url
        FROM
            m_playlist_video playlist_video
            INNER JOIN m_video video ON playlist_video.m_video_id = video.m_video_id
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
        LIMIT
            1
    ), '')::varchar AS top_thumbnail
FROM
    m_playlist playlist
WHERE
    (
        playlist.m_user_id = (
            SELECT
                u.m_user_id
            FROM
                m_user u
            WHERE
                u.public_id = @user_id
            LIMIT
                1
        )
        OR playlist.m_channel_id != 0
    )
    AND playlist.public_id = @playlist_id;

-- name: ListUserPlaylists :many
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.registered_at,
    playlist.updated_at,
    playlist.video_count,
    COALESCE((
        SELECT
            video.external_thumbnail_url
        FROM
            m_playlist_video playlist_video
            INNER JOIN m_video video ON playlist_video.m_video_id = video.m_video_id
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
        LIMIT
            1
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
        LIMIT
            1
    )
    AND (
        sqlc.narg('cursor')::uuid IS NULL
        OR playlist.public_id < sqlc.narg('cursor')::uuid
    )
ORDER BY
    playlist.m_user_id, playlist.public_id DESC
LIMIT
    @query_limit;

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
            LIMIT
                1
        )
        AND playlist.public_id = @playlist_id
    RETURNING
        playlist.public_id
)
SELECT
    deleted.public_id
FROM
    deleted;

-- name: InsertPlaylistVideo :exec
INSERT INTO
    m_playlist_video (
        m_playlist_id,
        m_video_id,
        playlist_position
    )
SELECT
    (
        SELECT
            m_playlist.m_playlist_id
        FROM
            m_playlist
        WHERE
            m_playlist.m_user_id = (
                SELECT
                    m_user.m_user_id
                FROM
                    m_user
                WHERE
                    m_user.public_id = @user_id
                LIMIT
                    1
            )
            AND m_playlist.public_id = @playlist_id
        LIMIT
            1
    ),
    (
        SELECT
            video.m_video_id
        FROM
            m_video video
        WHERE
            video.public_id = @video_id
        LIMIT
            1
    ),
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
                            SELECT
                                m_user.m_user_id
                            FROM
                                m_user
                            WHERE
                                m_user.public_id = @user_id
                            LIMIT
                                1
                        )
                        AND m_playlist.public_id = @playlist_id
                    LIMIT
                        1
                )
        ),
        0
    );

-- name: DeletePlaylistVideo :one
DELETE FROM
    m_playlist_video playlist_video
WHERE
    playlist_video.m_playlist_id = (
        SELECT
            playlist.m_playlist_id
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
                LIMIT
                    1
            )
            AND playlist.public_id = @playlist_id
        LIMIT
            1
    )
    AND playlist_video.m_video_id = (
        SELECT
            video.m_video_id
        FROM
            m_video video
        WHERE
            video.public_id = @video_id
        LIMIT
            1
    )
RETURNING
    playlist_video.m_playlist_video_id;

-- name: ListPlaylistVideos :many
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
                SELECT
                    m_user.m_user_id
                FROM
                    m_user
                WHERE
                    m_user.public_id = @user_id
                LIMIT
                    1
            )
            AND t_video_watch.m_video_id = video.m_video_id
        ORDER BY
            t_video_watch.m_user_id,
            t_video_watch.m_video_id,
            t_video_watch.watch_start_at DESC
        LIMIT
            1
    ), 0)::int AS last_watch_seconds,
    channel.public_id AS channel_id,
    channel.external_icon_url AS external_channel_icon_url,
    channel.external_display_name AS external_channel_displayname
FROM
    m_playlist_video playlist_video
    INNER JOIN m_video video ON video.m_video_id = playlist_video.m_video_id
    INNER JOIN m_channel channel ON channel.m_channel_id = video.m_channel_id
WHERE
    playlist_video.m_playlist_id = (
        SELECT
            playlist.m_playlist_id
        FROM
            m_playlist playlist
        WHERE
            (
                playlist.m_user_id = (
                    SELECT
                        u.m_user_id
                    FROM
                        m_user u
                    WHERE
                        u.public_id = @user_id
                    LIMIT
                        1
                )
                OR playlist.m_channel_id != 0
            )
            AND playlist.public_id = @playlist_id
        LIMIT
            1
    )
    AND (
        sqlc.narg('cursor')::uuid IS NULL
        OR playlist_video.playlist_position > (
            SELECT
                pv.playlist_position
            FROM
                m_playlist_video pv
                INNER JOIN m_video v ON v.m_video_id = pv.m_video_id
            WHERE
                v.public_id = sqlc.narg('cursor')::uuid
                AND pv.m_playlist_id = (
                    SELECT
                        playlist.m_playlist_id
                    FROM
                        m_playlist playlist
                    WHERE
                        (
                            playlist.m_user_id = (
                                SELECT
                                    u.m_user_id
                                FROM
                                    m_user u
                                WHERE
                                    u.public_id = @user_id
                                LIMIT
                                    1
                            )
                            OR playlist.m_channel_id != 0
                        )
                        AND playlist.public_id = @playlist_id
                    LIMIT
                        1
                )
            LIMIT
                1
        )
    )
ORDER BY
    playlist_video.m_playlist_id, playlist_video.playlist_position
LIMIT
    @query_limit;

-- name: ListRecentPlaylists :many
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.registered_at,
    playlist.video_count,
    COALESCE((
        SELECT
            video.external_thumbnail_url
        FROM
            m_playlist_video playlist_video
            INNER JOIN m_video video ON playlist_video.m_video_id = video.m_video_id
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
        LIMIT
            1
    ), '')::varchar AS top_thumbnail
FROM
    m_user u
    INNER JOIN LATERAL UNNEST(u.recent_playlist_ids) WITH ORDINALITY AS t(pid, ord) ON TRUE
    INNER JOIN m_playlist playlist ON playlist.m_playlist_id = t.pid
WHERE
    u.public_id = @user_id
ORDER BY
    t.ord;

-- name: BulkInsertPlaylistVideos :copyfrom
INSERT INTO
    m_playlist_video (
        m_playlist_id,
        m_video_id,
        playlist_position
    )
VALUES
    ($1, $2, $3);

-- name: ListChannelPlaylists :many
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.registered_at,
    playlist.updated_at,
    playlist.video_count,
    COALESCE((
        SELECT
            video.external_thumbnail_url
        FROM
            m_playlist_video playlist_video
            INNER JOIN m_video video ON playlist_video.m_video_id = video.m_video_id
        WHERE
            playlist_video.m_playlist_id = playlist.m_playlist_id
        LIMIT
            1
    ), '')::varchar AS top_thumbnail
FROM
    m_playlist playlist
WHERE
    playlist.m_channel_id = (
        SELECT
            ch.m_channel_id
        FROM
            m_channel ch
        WHERE
            ch.public_id = @channel_id
        LIMIT
            1
    )
    AND (
        sqlc.narg('cursor')::uuid IS NULL
        OR playlist.public_id < sqlc.narg('cursor')::uuid
    )
ORDER BY
    playlist.m_channel_id, playlist.public_id DESC
LIMIT
    @query_limit;
