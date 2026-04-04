package playlist

import (
	"context"
	"errors"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PlaylistRepository interface {
	Save(ctx context.Context, playlist *Playlist) (int64, error)
	Remove(ctx context.Context, userID, playlistID uuid.UUID) error
	FindForUpdate(ctx context.Context, userID, playlistID uuid.UUID) (*Playlist, error)
	InsertVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) error
	BulkInsertVideos(ctx context.Context, playlistInternalID int64, videoInternalIDs []int64) error
	RemoveVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) error
	CopyVideos(ctx context.Context, userID uuid.UUID, sourcePlaylistID uuid.UUID, destPlaylistInternalID int64) (int, error)
	PushRecentPlaylistID(ctx context.Context, userID, playlistID uuid.UUID) error
	FindWatchLaterForUpdate(ctx context.Context, userID uuid.UUID) (*Playlist, error)
	InsertWatchLater(ctx context.Context, userID, playlistID, videoID uuid.UUID) error
}

type playlistRepositoryImpl struct {
	q sqlc.Querier
}

func NewPlaylistRepository(q sqlc.Querier) PlaylistRepository {
	return &playlistRepositoryImpl{
		q: q,
	}
}

func (r *playlistRepositoryImpl) Save(ctx context.Context, playlist *Playlist) (_ int64, err error) {
	defer util.Wrap(&err, "playlist.(*playlistRepositoryImpl).Save(playlistID=%s)", playlist.ID)

	id, err := r.q.UpsertPlaylist(ctx, sqlc.UpsertPlaylistParams{
		UserPublicID:        playlist.UserID,
		ChannelPublicID:     playlist.ChannelID,
		PlaylistTitle:       string(playlist.Title),
		PlaylistDescription: string(playlist.Description),
		VisibilityCode:      int(playlist.VisibilityCode),
		PlaylistCode:        int(playlist.PlaylistCode),
		VideoCount:          playlist.VideoCount,
		PublicID:            playlist.ID,
		RegisteredAt:        playlist.RegisteredAt,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *playlistRepositoryImpl) Remove(ctx context.Context, userID uuid.UUID, playlistID uuid.UUID) error {
	_, err := r.q.DeletePlaylist(ctx, sqlc.DeletePlaylistParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	return err
}

func (r *playlistRepositoryImpl) FindForUpdate(ctx context.Context, userID, playlistID uuid.UUID) (*Playlist, error) {
	row, err := r.q.GetPlaylistForUpdate(ctx, sqlc.GetPlaylistForUpdateParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	if err != nil {
		return nil, err
	}

	p, err := NewPlaylist(
		userID,
		row.PlaylistTitle,
		row.PlaylistDescription,
		VisibilityCode(row.VisibilityCode).String(),
		PlaylistCode(row.PlaylistCode).String(),
		WithPlaylistID(row.PublicID),
		WithPlaylistVideoCount(row.VideoCount),
		WithPlaylistRegisteredAt(row.RegisteredAt),
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *playlistRepositoryImpl) InsertVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) error {
	return r.q.InsertPlaylistVideo(ctx, sqlc.InsertPlaylistVideoParams{
		UserID:     userID,
		PlaylistID: playlistID,
		VideoID:    videoID,
	})
}

func (r *playlistRepositoryImpl) BulkInsertVideos(ctx context.Context, playlistInternalID int64, videoInternalIDs []int64) (err error) {
	defer util.Wrap(&err, "playlist.(*playlistRepositoryImpl).BulkInsertVideos(playlistInternalID=%d)", playlistInternalID)

	params := make([]sqlc.BulkInsertPlaylistVideosParams, len(videoInternalIDs))
	for i, videoID := range videoInternalIDs {
		params[i] = sqlc.BulkInsertPlaylistVideosParams{
			MPlaylistID:      playlistInternalID,
			MVideoID:         videoID,
			PlaylistPosition: int64(i) * 1048576, // 2^20
		}
	}
	rowsAffected, err := r.q.BulkInsertPlaylistVideos(ctx, params)
	if err != nil {
		return err
	}
	if rowsAffected != int64(len(params)) {
		util.LoggerFromContext(ctx).WarnContext(ctx, "rowsAffected mismatch(playlistRepository.BulkInsertVideos)", slog.Int("expected", len(params)), slog.Int64("actual", rowsAffected))
	}
	return nil
}

func (r *playlistRepositoryImpl) RemoveVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "plalyist.(*playlistRepositoryImpl).RemoveVideo")

	_, err = r.q.DeletePlaylistVideo(ctx, sqlc.DeletePlaylistVideoParams{
		UserID:     userID,
		PlaylistID: playlistID,
		VideoID:    videoID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.ErrNotFound
		}
		return err
	}

	return nil
}

func (r *playlistRepositoryImpl) CopyVideos(ctx context.Context, userID uuid.UUID, sourcePlaylistID uuid.UUID, destPlaylistInternalID int64) (_ int, err error) {
	defer util.Wrap(&err, "playlist.(*playlistRepositoryImpl).CopyVideos(sourcePlaylistID=%s, destPlaylistInternalID=%d)", sourcePlaylistID, destPlaylistInternalID)

	copiedCount, err := r.q.CopyPlaylistVideos(ctx, sqlc.CopyPlaylistVideosParams{
		DestPlaylistID:   destPlaylistInternalID,
		UserID:           userID,
		SourcePlaylistID: sourcePlaylistID,
	})
	if err != nil {
		return 0, err
	}
	return copiedCount, nil
}

func (r *playlistRepositoryImpl) PushRecentPlaylistID(ctx context.Context, userID, playlistID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*playlistRepositoryImpl).PushRecentPlaylistID(userID=%s, playlistID=%s)", userID, playlistID)

	return r.q.PushRecentPlaylistId(ctx, sqlc.PushRecentPlaylistIdParams{
		PlaylistPublicID: playlistID,
		UserPublicID:     userID,
	})
}

func (r *playlistRepositoryImpl) FindWatchLaterForUpdate(ctx context.Context, userID uuid.UUID) (_ *Playlist, err error) {
	defer util.Wrap(&err, "playlist.(*playlistRepositoryImpl).FindWatchLaterForUpdate")

	row, err := r.q.GetWatchLaterForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}

	p, err := NewPlaylist(
		userID,
		row.PlaylistTitle,
		row.PlaylistDescription,
		VisibilityCode(row.VisibilityCode).String(),
		PlaylistCode(row.PlaylistCode).String(),
		WithPlaylistID(row.PublicID),
		WithPlaylistVideoCount(row.VideoCount),
		WithPlaylistRegisteredAt(row.RegisteredAt),
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// InsertWatchLater watch laterとして挿入する。
// 普通のinsertと違うのは、動画の重複チェックがあること
func (r *playlistRepositoryImpl) InsertWatchLater(ctx context.Context, userID, playlistID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*playlistRepositoryImpl).InsertWatchLater")

	if err := r.q.InsertWatchLater(ctx, sqlc.InsertWatchLaterParams{
		VideoID:    videoID,
		PlaylistID: playlistID,
	}); err != nil {
		return err
	}

	return nil
}

var _ PlaylistRepository = (*playlistRepositoryImpl)(nil)
