package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetSubscriptionsView struct {
	ChannelId                  uuid.UUID
	ExternalChannelId          string
	ChannelCustomId            string
	ExternalChannelDisplayName string
	ExternalChannelIconUrl     string
	ChannelSubscribersCount    int64
}

type SubscriptionQueryService interface {
	GetSubscriptions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetSubscriptionsView, error)
}

type subscriptionQueryServiceImpl struct {
	q sqlc.Querier
}

func NewSubscriptionQueryService(db *pgxpool.Pool) SubscriptionQueryService {
	return &subscriptionQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (c *subscriptionQueryServiceImpl) GetSubscriptions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetSubscriptionsView, error) {
	rows, err := c.q.ListSubscribedChannels(ctx, sqlc.ListSubscribedChannelsParams{
		UserPublicID:   userID,
		CursorPublicID: cursor,
		QueryLimit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to getSubscribingChannel")
	}

	views := make([]GetSubscriptionsView, len(rows))
	for i, row := range rows {
		views[i] = GetSubscriptionsView{
			ChannelId:                  row.ChannelPublicID,
			ExternalChannelId:          row.ExternalID,
			ChannelCustomId:            row.ExternalCustomID,
			ExternalChannelDisplayName: row.ExternalDisplayName,
			ExternalChannelIconUrl:     row.ExternalIconUrl,
			ChannelSubscribersCount:    row.ExternalSubscribersCount,
		}
	}
	return views, nil
}

var _ SubscriptionQueryService = (*subscriptionQueryServiceImpl)(nil)

type GetValuableChannelView struct {
	ChannelId                  uuid.UUID
	ExternalChannelCustomUrl   string
	ExternalChannelDisplayName string
	ExternalChannelIconUrl     string
	CategoryCode               int
	ValuableDescription        string
}

type ValuableChannelQueryService interface {
	GetValuableChannels(ctx context.Context) ([]GetValuableChannelView, error)
}

type valuableChannelQueryServiceImpl struct {
	q sqlc.Querier
}

func NewValuableChannelQueryService(db *pgxpool.Pool) ValuableChannelQueryService {
	return &valuableChannelQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (v *valuableChannelQueryServiceImpl) GetValuableChannels(ctx context.Context) (_ []GetValuableChannelView, err error) {
	defer util.Wrap(&err, "valuableChannelQueryService.GetValuableChannels")

	rows, err := v.q.ListValuableChannels(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]GetValuableChannelView, len(rows))
	for i, row := range rows {
		views[i] = GetValuableChannelView{
			ChannelId:                  row.ChannelPublicID,
			ExternalChannelCustomUrl:   row.ExternalCustomID,
			ExternalChannelDisplayName: row.ExternalDisplayName,
			ExternalChannelIconUrl:     row.ExternalIconUrl,
			CategoryCode:               row.CategoryCode,
			ValuableDescription:        row.ValuableDescription,
		}
	}
	return views, nil
}

var _ ValuableChannelQueryService = (*valuableChannelQueryServiceImpl)(nil)

type GetChannelUploadsView struct {
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
	ExternalVideoThumbnailUrl  string
	ExternalVideoTitle         string
	LastWatchSeconds           *int
	VideoId                    uuid.UUID
}

type UploadsQueryService interface {
	GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetChannelUploadsView, error)
}

type uploadsQueryServiceImpl struct {
	q sqlc.Querier
}

func NewUploadsQueryService(db *pgxpool.Pool) UploadsQueryService {
	return &uploadsQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (u *uploadsQueryServiceImpl) GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetChannelUploadsView, err error) {
	defer util.Wrap(&err, "uploadsQueryService.GetChannelUploads(userID=%s, channelID=%s)", userID, channelID)
	rows, err := u.q.ListChannelVideos(ctx, sqlc.ListChannelVideosParams{
		UserID:     userID,
		ChannelID:  channelID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
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

var _ UploadsQueryService = (*uploadsQueryServiceImpl)(nil)

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
	defer util.Wrap(&err, "feedQueryService.GetVideoFeed(userID=%s)", userID)
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
