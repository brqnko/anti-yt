package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type ChannelQueryService interface {
	GetSubscriptions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetSubscriptionsView, error)
	GetChannelDetail(ctx context.Context, channelID *uuid.UUID, externalChannelID *string) (GetChannelDetailView, error)
	GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32, order string) ([]GetChannelUploadsView, error)
	ListChannelVideoIDs(ctx context.Context, userID, channelID uuid.UUID, limit int32) ([]uuid.UUID, error)
}

type channelQueryServiceImpl struct {
	q sqlc.Querier
}

func NewChannelQueryService(db *pgxpool.Pool) ChannelQueryService {
	return &channelQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (c *channelQueryServiceImpl) GetSubscriptions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetSubscriptionsView, error) {
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
	defer util.Wrap(&err, "channel.(*valuableChannelQueryServiceImpl).GetValuableChannels")

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

type GetChannelDetailView struct {
	ChannelID          uuid.UUID
	CustomID           string
	DisplayName        string
	Description        string
	IconURL            string
	SubscribersCount   int64
}

func (c *channelQueryServiceImpl) GetChannelDetail(ctx context.Context, channelID *uuid.UUID, externalChannelID *string) (_ GetChannelDetailView, err error) {
	defer util.Wrap(&err, "channel.(*channelQueryServiceImpl).GetChannelDetail")

	row, err := c.q.GetChannelByPublicID(ctx, sqlc.GetChannelByPublicIDParams{
		ChannelID:         channelID,
		ExternalChannelID: externalChannelID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetChannelDetailView{}, core.ErrNotFound
		}
		return GetChannelDetailView{}, err
	}

	return GetChannelDetailView{
		ChannelID:        row.PublicID,
		CustomID:         row.ExternalCustomID,
		DisplayName:      row.ExternalDisplayName,
		Description:      row.ExternalDescription,
		IconURL:          row.ExternalIconUrl,
		SubscribersCount: row.ExternalSubscribersCount,
	}, nil
}

type GetChannelUploadsView struct {
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
	ExternalVideoThumbnailUrl  string
	ExternalVideoTitle         string
	IsWatched                  bool
	LastWatchSeconds           *int
	VideoId                    uuid.UUID
}

func (c *channelQueryServiceImpl) GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32, order string) (_ []GetChannelUploadsView, err error) {
	defer util.Wrap(&err, "channel.(*channelQueryServiceImpl).GetChannelUploads(userID=%s, channelID=%s)", userID, channelID)

	if order == "older" {
		rows, err := c.q.ListChannelVideosOlder(ctx, sqlc.ListChannelVideosOlderParams{
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
				IsWatched:                  row.IsWatched,
				LastWatchSeconds:           lastWatchSeconds,
				VideoId:                    row.PublicID,
			}
		}
		return views, nil
	}

	rows, err := c.q.ListChannelVideos(ctx, sqlc.ListChannelVideosParams{
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
			IsWatched:                  row.IsWatched,
			LastWatchSeconds:           lastWatchSeconds,
			VideoId:                    row.PublicID,
		}
	}
	return views, nil
}

func (c *channelQueryServiceImpl) ListChannelVideoIDs(ctx context.Context, userID, channelID uuid.UUID, limit int32) (_ []uuid.UUID, err error) {
	defer util.Wrap(&err, "channel.(*channelQueryServiceImpl).ListChannelVideoIDs(userID=%s,channelID=%s)", userID, channelID)

	_ = userID
	rows, err := c.q.ListChannelVideoIDs(ctx, sqlc.ListChannelVideoIDsParams{
		ChannelID:  channelID,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

var _ ChannelQueryService = (*channelQueryServiceImpl)(nil)
