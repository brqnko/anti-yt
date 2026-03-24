package playlist

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
)

type PlaylistRepository interface {
	Save(ctx context.Context, userID uuid.UUID, playlist *Playlist) (int64, error)
	Remove(ctx context.Context, userID, playlistID uuid.UUID) error
	FindForUpdate(ctx context.Context, userID, playlistID uuid.UUID) (*Playlist, error)
	InsertVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) error
	BulkInsertVideos(ctx context.Context, playlistInternalID int64, videoInternalIDs []int64) error
	RemoveVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) error
}

type playlistRepositoryImpl struct {
	q sqlc.Querier
}

func NewPlaylistRepository(q sqlc.Querier) PlaylistRepository {
	return &playlistRepositoryImpl{
		q: q,
	}
}

func (r *playlistRepositoryImpl) Save(ctx context.Context, userID uuid.UUID, playlist *Playlist) (int64, error) {
	id, err := r.q.UpsertPlaylist(ctx, sqlc.UpsertPlaylistParams{
		UserPublicID:        userID,
		PlaylistTitle:       string(playlist.Title),
		PlaylistDescription: string(playlist.Description),
		VisibilityCode:      int(playlist.VisibilityCode),
		PlaylistCode:        int(playlist.PlaylistCode),
		VideoCount:          playlist.VideoCount,
		PublicID:            playlist.ID,
		RegisteredAt:        playlist.RegisteredAt,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to savePlaylist(playlistRepository.Save): %w", err)
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

func (r *playlistRepositoryImpl) BulkInsertVideos(ctx context.Context, playlistInternalID int64, videoInternalIDs []int64) error {
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
		return fmt.Errorf("failed to bulkInsertIntoPlaylist(playlistRepository.BulkInsertVideos): %w", err)
	}
	if rowsAffected != int64(len(params)) {
		slog.Warn("rowsAffected mismatch(playlistRepository.BulkInsertVideos)", "expected", len(params), "actual", rowsAffected)
	}
	return nil
}

func (r *playlistRepositoryImpl) RemoveVideo(ctx context.Context, userID, playlistID, videoID uuid.UUID) error {
	_, err := r.q.DeletePlaylistVideo(ctx, sqlc.DeletePlaylistVideoParams{
		UserID:     userID,
		PlaylistID: playlistID,
		VideoID:    videoID,
	})
	return err
}

var _ PlaylistRepository = (*playlistRepositoryImpl)(nil)
