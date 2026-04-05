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
	GetVideoFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetVideoFeedView, error)
}

type feedQueryServiceImpl struct {
	q sqlc.Querier
}

func NewFeedQueryService(db *pgxpool.Pool) FeedQueryService {
	return &feedQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (f *feedQueryServiceImpl) GetVideoFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetVideoFeedView, err error) {
	defer util.Wrap(&err, "feed.(*feedQueryServiceImpl).GetVideoFeed(userID=%s)", userID)

	rows, err := f.q.ListSubscriptionFeed(ctx, sqlc.ListSubscriptionFeedParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: limit,
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

var _ FeedQueryService = (*feedQueryServiceImpl)(nil)
