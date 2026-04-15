package feed

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

type FeedQueryService interface {
	ListAllActiveUserIDs(ctx context.Context) ([]uuid.UUID, error)
	ListSubscriptionVideoIDs(ctx context.Context, userID uuid.UUID, limit int32) ([]uuid.UUID, error)
	HydrateVideoFeed(ctx context.Context, userID uuid.UUID, videoIDs []uuid.UUID) ([]GetVideoFeedView, error)
	ListLatestVideos(ctx context.Context, cursor *uuid.UUID, limit int32) ([]GetVideoFeedView, error)
}

type feedQueryServiceImpl struct {
	q sqlc.Querier
}

func NewFeedQueryService(db *pgxpool.Pool) FeedQueryService {
	return &feedQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (f *feedQueryServiceImpl) ListAllActiveUserIDs(ctx context.Context) (_ []uuid.UUID, err error) {
	defer util.Wrap(&err, "feed.(*feedQueryServiceImpl).ListAllActiveUserIDs")

	rows, err := f.q.ListAllActiveUserIDs(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (f *feedQueryServiceImpl) ListSubscriptionVideoIDs(ctx context.Context, userID uuid.UUID, limit int32) (_ []uuid.UUID, err error) {
	defer util.Wrap(&err, "feed.(*feedQueryServiceImpl).ListSubscriptionVideoIDs(userID=%s)", userID)

	rows, err := f.q.ListSubscriptionFeed(ctx, sqlc.ListSubscriptionFeedParams{
		UserID:     userID,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (f *feedQueryServiceImpl) HydrateVideoFeed(ctx context.Context, userID uuid.UUID, videoIDs []uuid.UUID) (_ []GetVideoFeedView, err error) {
	defer util.Wrap(&err, "feed.(*feedQueryServiceImpl).HydrateVideoFeed(userID=%s)", userID)

	if len(videoIDs) == 0 {
		return []GetVideoFeedView{}, nil
	}

	rows, err := f.q.ListVideoFeedByIDs(ctx, sqlc.ListVideoFeedByIDsParams{
		UserID:   userID,
		VideoIds: videoIDs,
	})
	if err != nil {
		return nil, err
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

func (f *feedQueryServiceImpl) ListLatestVideos(ctx context.Context, cursor *uuid.UUID, limit int32) (_ []GetVideoFeedView, err error) {
	defer util.Wrap(&err, "feed.(*feedQueryServiceImpl).ListLatestVideos")

	rows, err := f.q.ListLatestVideos(ctx, sqlc.ListLatestVideosParams{
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	views := make([]GetVideoFeedView, len(rows))
	for i, row := range rows {
		views[i] = GetVideoFeedView{
			VideoId:                    row.VideoID,
			ExternalVideoThumbnailUrl:  row.ExternalVideoThumbnailUrl,
			ExternalVideoTitle:         row.ExternalTitle,
			ExternalVideoCreatedAt:     row.ExternalCreatedAt,
			ExternalVideoLengthSeconds: row.ExternalLengthSeconds,
			ChannelId:                  row.ChannelID,
			ExternalChannelIconUrl:     row.ExternalChannelIconUrl,
			ExternalChannelDisplayName: row.ExternalDisplayname,
		}
	}
	return views, nil
}

var _ FeedQueryService = (*feedQueryServiceImpl)(nil)
