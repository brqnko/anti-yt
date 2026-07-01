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
        registered_at,
        external_id
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
        @registered_at,
        @external_id
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

-- name: GetPlaylistForUpdate :one
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.video_count,
    playlist.registered_at,
    playlist.external_id
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
        AND playlist.playlist_code = 0
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
    EXISTS (
        SELECT 1 FROM t_video_watched
        WHERE t_video_watched.m_video_id = video.m_video_id
            AND t_video_watched.m_user_id = (
                SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_id LIMIT 1
            )
    )::bool AS is_watched,
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
                OR playlist.visibility_code = 1
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

-- name: CopyPlaylistVideos :one
WITH inserted AS (
    INSERT INTO
        m_playlist_video (m_playlist_id, m_video_id, playlist_position)
    SELECT
        @dest_playlist_id, pv.m_video_id, pv.playlist_position
    FROM
        m_playlist_video pv
    WHERE
        pv.m_playlist_id = (
            SELECT
                p.m_playlist_id
            FROM
                m_playlist p
            WHERE
                (
                    p.m_user_id = (
                        SELECT
                            u.m_user_id
                        FROM
                            m_user u
                        WHERE
                            u.public_id = @user_id
                        LIMIT
                            1
                    )
                    OR p.visibility_code = 1
                )
                AND p.public_id = @source_playlist_id
            LIMIT
                1
        )
    RETURNING 1
)
SELECT COUNT(*)::int AS copied_count FROM inserted;

-- name: GetWatchLaterForUpdate :one
SELECT
    playlist.public_id,
    playlist.playlist_title,
    playlist.playlist_description,
    playlist.visibility_code,
    playlist.playlist_code,
    playlist.video_count,
    playlist.registered_at,
    playlist.external_id
FROM m_playlist playlist
WHERE
    playlist.m_user_id = (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_id LIMIT 1)
    AND playlist.playlist_code = 2
LIMIT 1
FOR UPDATE;

-- name: InsertWatchLater :exec
INSERT INTO
    m_playlist_video (
        m_playlist_id,
        m_video_id,
        playlist_position
    )
SELECT
    m_playlist.m_playlist_id,
    (SELECT m_video.m_video_id FROM m_video WHERE m_video.public_id = @video_id LIMIT 1),
    COALESCE(
            (
                SELECT
                    MAX(m_playlist_video.playlist_position) + 1048576 -- NOTE: 2^20
                FROM
                    m_playlist_video
                WHERE
                    m_playlist_video.m_playlist_id = m_playlist.m_playlist_id
            ),
            0
    )
FROM m_playlist
WHERE
    m_playlist.public_id = @playlist_id
    AND NOT EXISTS(
        SELECT 1 FROM m_playlist_video
        WHERE
            m_playlist_video.m_playlist_id = m_playlist.m_playlist_id
            AND m_playlist_video.m_video_id = (SELECT m_video.m_video_id FROM m_video WHERE m_video.public_id = @video_id LIMIT 1));

-- name: GetWatchLaterPlaylist :one
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
    AND playlist.playlist_code = 2
LIMIT
    1;

-- name: BulkInsertPlaylistVideos :copyfrom
INSERT INTO
    m_playlist_video (
        m_playlist_id,
        m_video_id,
        playlist_position
    )
VALUES
    (@m_playlist_id, @m_video_id, @playlist_position);

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

-- m_user.recent_playlist_idsを更新する。先頭に追加し、重複を除去し、最大5件に制限する。
-- name: PushRecentPlaylistId :exec
UPDATE m_user
SET recent_playlist_ids = (
    SELECT COALESCE(array_agg(val), '{}')
    FROM (
             SELECT val
             FROM UNNEST(
                          ARRAY[(SELECT p.m_playlist_id FROM m_playlist p WHERE p.public_id = @playlist_public_id LIMIT 1)]
            || m_user.recent_playlist_ids
        ) WITH ORDINALITY AS t(val, ord)
             WHERE val IS NOT NULL
             GROUP BY val
             ORDER BY MIN(ord)
                 LIMIT 5
         ) sub
),
    updated_at = CURRENT_TIMESTAMP
WHERE m_user.public_id = @user_public_id;

-- name: UpsertSystemPlaylist :one
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
        registered_at,
        external_id
    )
VALUES
    (
        0,
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
        @registered_at,
        @external_id
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

-- name: CountUserPlaylists :one
SELECT COUNT(*)::int
FROM m_playlist
WHERE m_user_id = (
    SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @user_public_id LIMIT 1
);

-- name: CountPlaylistByExternalID :one
SELECT COUNT(*)::int
FROM m_playlist
WHERE external_id = @external_id;
