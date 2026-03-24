package playlist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetPlaylistsView struct {
	PlaylistRegisteredAt time.Time
	PlaylistDescription  string
	PlaylistId           uuid.UUID
	PlaylistTitle        string
	PlaylistType         string
	PlaylistUpdatedAt    time.Time
	PlaylistVideoCount   int
	PlaylistVisibility   string
	TopVideoThumbnailUrl *string
}

type GetPlaylistDetailView struct {
	PlaylistDescription  string
	PlaylistId           uuid.UUID
	PlaylistRegisteredAt time.Time
	PlaylistTitle        string
	PlaylistType         string
	PlaylistUpdatedAt    time.Time
	PlaylistVideoCount   int
	PlaylistVisibility   string
	TopVideoThumbnailUrl *string
}

type GetPlaylistItemView struct {
	ChannelId                  uuid.UUID
	ExternalChannelDisplayName string
	ExternalChannelIconUrl     string
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
	ExternalVideoThumbnailUrl  string
	ExternalVideoTitle         string
	LastWatchSeconds           *int
	VideoId                    uuid.UUID
}

type PlaylistQueryService interface {
	FindPlaylists(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetPlaylistsView, error)
	Find(ctx context.Context, userID uuid.UUID, playlistID uuid.UUID) (GetPlaylistDetailView, error)
	FindPlaylistItems(ctx context.Context, userID, playlistID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetPlaylistItemView, error)
}

type playlistQueryServiceImpl struct {
	q sqlc.Querier
}

func NewPlaylistQueryService(db *pgxpool.Pool) PlaylistQueryService {
	return &playlistQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (p *playlistQueryServiceImpl) FindPlaylists(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetPlaylistsView, error) {
	rows, err := p.q.ListUserPlaylists(ctx, sqlc.ListUserPlaylistsParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to getUserPlaylists(playlistQueryService.FindPlaylists): %w", err)
	}

	views := make([]GetPlaylistsView, len(rows))
	for i, row := range rows {
		var topVideoThumbnailUrl *string
		if row.TopThumbnail != "" {
			topVideoThumbnailUrl = &row.TopThumbnail
		}
		views[i] = GetPlaylistsView{
			PlaylistRegisteredAt: row.RegisteredAt,
			PlaylistDescription:  row.PlaylistDescription,
			PlaylistId:           row.PublicID,
			PlaylistTitle:        row.PlaylistTitle,
			PlaylistType:         PlaylistCode(row.PlaylistCode).String(),
			PlaylistUpdatedAt:    row.UpdatedAt,
			PlaylistVideoCount:   row.VideoCount,
			PlaylistVisibility:   VisibilityCode(row.VisibilityCode).String(),
			TopVideoThumbnailUrl: topVideoThumbnailUrl,
		}
	}
	return views, nil
}

func (p *playlistQueryServiceImpl) Find(ctx context.Context, userID uuid.UUID, playlistID uuid.UUID) (GetPlaylistDetailView, error) {
	row, err := p.q.GetPlaylistWithThumbnail(ctx, sqlc.GetPlaylistWithThumbnailParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetPlaylistDetailView{}, err
		}
		return GetPlaylistDetailView{}, fmt.Errorf("failed to getPlaylist(playlistQueryService.Find): %w", err)
	}

	var topVideoThumbnailUrl *string
	if row.TopThumbnail != "" {
		topVideoThumbnailUrl = &row.TopThumbnail
	}

	return GetPlaylistDetailView{
		PlaylistDescription:  row.PlaylistDescription,
		PlaylistId:           row.PublicID,
		PlaylistRegisteredAt: row.RegisteredAt,
		PlaylistTitle:        row.PlaylistTitle,
		PlaylistType:         PlaylistCode(row.PlaylistCode).String(),
		PlaylistUpdatedAt:    row.UpdatedAt,
		PlaylistVideoCount:   row.VideoCount,
		PlaylistVisibility:   VisibilityCode(row.VisibilityCode).String(),
		TopVideoThumbnailUrl: topVideoThumbnailUrl,
	}, nil
}

func (p *playlistQueryServiceImpl) FindPlaylistItems(ctx context.Context, userID, playlistID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetPlaylistItemView, error) {
	rows, err := p.q.ListPlaylistVideos(ctx, sqlc.ListPlaylistVideosParams{
		UserID:     userID,
		PlaylistID: playlistID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to getPlaylistVideos(playlistQueryService.FindPlaylistItems): %w", err)
	}

	views := make([]GetPlaylistItemView, len(rows))
	for i, row := range rows {
		var lastWatchSeconds *int
		if row.LastWatchSeconds != 0 {
			lastWatchSeconds = &row.LastWatchSeconds
		}
		views[i] = GetPlaylistItemView{
			ChannelId:                  row.ChannelID,
			ExternalChannelDisplayName: row.ExternalChannelDisplayname,
			ExternalChannelIconUrl:     row.ExternalChannelIconUrl,
			ExternalVideoCreatedAt:     row.ExternalCreatedAt,
			ExternalVideoLengthSeconds: row.ExternalLengthSeconds,
			ExternalVideoThumbnailUrl:  row.ExternalThumbnailUrl,
			ExternalVideoTitle:         row.ExternalTitle,
			LastWatchSeconds:           lastWatchSeconds,
			VideoId:                    row.PublicID,
		}
	}
	return views, nil
}

var _ PlaylistQueryService = (*playlistQueryServiceImpl)(nil)
