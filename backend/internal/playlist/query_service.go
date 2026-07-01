package playlist

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
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
	IsWatched                  bool
	LastWatchSeconds           *int
	VideoId                    uuid.UUID
}

type GetChannelPlaylistsView struct {
	PlaylistId           uuid.UUID
	PlaylistTitle        string
	PlaylistVideoCount   int
	PlaylistRegisteredAt time.Time
	TopVideoThumbnailUrl *string
}

type PlaylistQueryService interface {
	FindPlaylists(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetPlaylistsView, error)
	FindChannelPlaylists(ctx context.Context, channelID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetChannelPlaylistsView, error)
	Find(ctx context.Context, userID uuid.UUID, playlistID uuid.UUID) (GetPlaylistDetailView, error)
	FindPlaylistItems(ctx context.Context, userID, playlistID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetPlaylistItemView, error)
	FindRecentPlaylists(ctx context.Context, userID uuid.UUID) ([]GetChannelPlaylistsView, error)
	ExistsByExternalID(ctx context.Context, externalID string) (bool, error)
}

type playlistQueryServiceImpl struct {
	q sqlc.Querier
}

func NewPlaylistQueryService(db *pgxpool.Pool) PlaylistQueryService {
	return &playlistQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (p *playlistQueryServiceImpl) ExistsByExternalID(ctx context.Context, externalID string) (_ bool, err error) {
	defer util.Wrap(&err, "playlist.(*playlistQueryServiceImpl).ExistsByExternalID(externalID=%s)", externalID)

	count, err := p.q.CountPlaylistByExternalID(ctx, externalID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *playlistQueryServiceImpl) FindPlaylists(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetPlaylistsView, err error) {
	defer util.Wrap(&err, "playlist.(*playlistQueryServiceImpl).FindPlaylists(userID=%s)", userID)

	rows, err := p.q.ListUserPlaylists(ctx, sqlc.ListUserPlaylistsParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
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

func (p *playlistQueryServiceImpl) FindChannelPlaylists(ctx context.Context, channelID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetChannelPlaylistsView, err error) {
	defer util.Wrap(&err, "playlist.(*playlistQueryServiceImpl).FindChannelPlaylists(channelID=%s)", channelID)

	rows, err := p.q.ListChannelPlaylists(ctx, sqlc.ListChannelPlaylistsParams{
		ChannelID:  channelID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	views := make([]GetChannelPlaylistsView, len(rows))
	for i, row := range rows {
		var topVideoThumbnailUrl *string
		if row.TopThumbnail != "" {
			topVideoThumbnailUrl = &row.TopThumbnail
		}
		views[i] = GetChannelPlaylistsView{
			PlaylistId:           row.PublicID,
			PlaylistTitle:        row.PlaylistTitle,
			PlaylistVideoCount:   row.VideoCount,
			PlaylistRegisteredAt: row.RegisteredAt,
			TopVideoThumbnailUrl: topVideoThumbnailUrl,
		}
	}
	return views, nil
}

func (p *playlistQueryServiceImpl) Find(ctx context.Context, userID uuid.UUID, playlistID uuid.UUID) (_ GetPlaylistDetailView, err error) {
	defer util.Wrap(&err, "playlist.(*playlistQueryServiceImpl).Find(userID=%s, playlistID=%s)", userID, playlistID)

	row, err := p.q.GetPlaylistWithThumbnail(ctx, sqlc.GetPlaylistWithThumbnailParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetPlaylistDetailView{}, core.ErrNotFound
		}
		return GetPlaylistDetailView{}, err
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

func (p *playlistQueryServiceImpl) FindPlaylistItems(ctx context.Context, userID, playlistID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetPlaylistItemView, err error) {
	defer util.Wrap(&err, "playlist.(*playlistQueryServiceImpl).FindPlaylistItems(userID=%s, playlistID=%s)", userID, playlistID)

	rows, err := p.q.ListPlaylistVideos(ctx, sqlc.ListPlaylistVideosParams{
		UserID:     userID,
		PlaylistID: playlistID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
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
			IsWatched:                  row.IsWatched,
			LastWatchSeconds:           lastWatchSeconds,
			VideoId:                    row.PublicID,
		}
	}
	return views, nil
}

func (p *playlistQueryServiceImpl) FindRecentPlaylists(ctx context.Context, userID uuid.UUID) (_ []GetChannelPlaylistsView, err error) {
	defer util.Wrap(&err, "playlist.(*playlistQueryServiceImpl).FindRecentPlaylists(userID=%s)", userID)

	wl, wlErr := p.q.GetWatchLaterPlaylist(ctx, userID)

	rows, err := p.q.ListRecentPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	var views []GetChannelPlaylistsView

	if wlErr == nil {
		var topVideoThumbnailUrl *string
		if wl.TopThumbnail != "" {
			topVideoThumbnailUrl = &wl.TopThumbnail
		}
		views = append(views, GetChannelPlaylistsView{
			PlaylistId:           wl.PublicID,
			PlaylistTitle:        wl.PlaylistTitle,
			PlaylistVideoCount:   wl.VideoCount,
			PlaylistRegisteredAt: wl.RegisteredAt,
			TopVideoThumbnailUrl: topVideoThumbnailUrl,
		})
	}

	for _, row := range rows {
		if wlErr == nil && row.PublicID == wl.PublicID {
			continue
		}
		var topVideoThumbnailUrl *string
		if row.TopThumbnail != "" {
			topVideoThumbnailUrl = &row.TopThumbnail
		}
		views = append(views, GetChannelPlaylistsView{
			PlaylistId:           row.PublicID,
			PlaylistTitle:        row.PlaylistTitle,
			PlaylistVideoCount:   row.VideoCount,
			PlaylistRegisteredAt: row.RegisteredAt,
			TopVideoThumbnailUrl: topVideoThumbnailUrl,
		})
	}
	return views, nil
}

var _ PlaylistQueryService = (*playlistQueryServiceImpl)(nil)
