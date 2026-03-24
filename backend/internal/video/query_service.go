package video

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

type GetVideoDetailView struct {
	VideoId                         uuid.UUID
	ExternalVideoId                 string
	ExternalVideoTitle              string
	ExternalVideoDescription        string
	ExternalVideoThumbnailUrl       string
	ChannelId                       uuid.UUID
	ChannelCustomId                 string
	ExternalChannelDisplayName      string
	ExternalChannelIconUrl          string
	ExternalChannelSubscribersCount uint64
}

type GetChannelUploadsView struct {
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
	ExternalVideoThumbnailUrl  string
	ExternalVideoTitle         string
	LastWatchSeconds           *int
	VideoId                    uuid.UUID
}

type GetVideoFeedView struct {
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

type VideoQueryService interface {
	Find(ctx context.Context, id uuid.UUID) (GetVideoDetailView, error)
	GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetChannelUploadsView, error)
	GetVideoFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetVideoFeedView, error)
}

type videoQueryServiceImpl struct {
	q sqlc.Querier
}

func NewVideoQueryService(db *pgxpool.Pool) VideoQueryService {
	return &videoQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (v *videoQueryServiceImpl) Find(ctx context.Context, id uuid.UUID) (GetVideoDetailView, error) {
	row, err := v.q.GetVideoDetail(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetVideoDetailView{}, err
		}
		return GetVideoDetailView{}, fmt.Errorf("failed to getVideoDetail(videoQueryService.Find): %w", err)
	}

	return GetVideoDetailView{
		VideoId:                         row.ID,
		ExternalVideoId:                 row.ExternalID,
		ExternalVideoTitle:              row.ExternalTitle,
		ExternalVideoDescription:        row.ExternalDescription,
		ExternalVideoThumbnailUrl:       row.ExternalThumbnailUrl,
		ChannelId:                       row.ChannelID,
		ChannelCustomId:                 row.ChannelCustomID,
		ExternalChannelDisplayName:      row.ExternalDisplayName,
		ExternalChannelIconUrl:          row.ExternalIconUrl,
		ExternalChannelSubscribersCount: uint64(row.ExternalSubscribersCount),
	}, nil
}

func (v *videoQueryServiceImpl) GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetChannelUploadsView, error) {
	rows, err := v.q.ListChannelVideos(ctx, sqlc.ListChannelVideosParams{
		UserID:     userID,
		ChannelID:  channelID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to listChannelVideos(videoQueryService.GetChannelUploads): %w", err)
	}

	views := make([]GetChannelUploadsView, len(rows))
	for i, row := range rows {
		var lastWatchSeconds *int
		if row.LastWatchSeconds != 0 {
			lastWatchSeconds = &row.LastWatchSeconds
		}
		views[i] = GetChannelUploadsView{
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

func (v *videoQueryServiceImpl) GetVideoFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetVideoFeedView, error) {
	rows, err := v.q.ListSubscriptionFeed(ctx, sqlc.ListSubscriptionFeedParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to listSubscriptionFeed(videoQueryService.GetVideoFeed): %w", err)
	}

	views := make([]GetVideoFeedView, len(rows))
	for i, row := range rows {
		var lastWatchSeconds *int
		if row.LastWatchSeconds != 0 {
			lastWatchSeconds = &row.LastWatchSeconds
		}
		views[i] = GetVideoFeedView{
			ChannelId:                  row.ChannelID,
			ExternalChannelDisplayName: row.ExternalDisplayname,
			ExternalChannelIconUrl:     row.ExternalChannelIconUrl,
			ExternalVideoCreatedAt:     row.ExternalCreatedAt,
			ExternalVideoLengthSeconds: row.ExternalLengthSeconds,
			ExternalVideoThumbnailUrl:  row.ExternalVideoThumbnailUrl,
			ExternalVideoTitle:         row.ExternalTitle,
			LastWatchSeconds:           lastWatchSeconds,
			VideoId:                    row.VideoID,
		}
	}
	return views, nil
}

var _ VideoQueryService = (*videoQueryServiceImpl)(nil)
