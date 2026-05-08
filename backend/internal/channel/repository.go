package channel

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChannelRepository interface {
	Save(ctx context.Context, channel *Channel) (int64, error)
	FindForUpdate(ctx context.Context, id uuid.UUID) (*Channel, error)
	FindByIdOrHandle(ctx context.Context, idOrHandle string) (*Channel, error)
	FindToFetchRSSForUpdate(ctx context.Context, userID uuid.UUID, rssFetchDuration time.Duration, limit int32) ([]*Channel, error)
	FindBulkFetchedBefore(ctx context.Context, before time.Time) ([]*Channel, error)
	SaveSubscription(ctx context.Context, subscribedChannel *SubscribedChannel) (int64, error)
	RemoveSubscription(ctx context.Context, userID, channelID uuid.UUID) (int64, error)
}

func NewChannelRepository(q sqlc.Querier) ChannelRepository {
	return &channelRepositoryImpl{
		q: q,
	}
}

type channelRepositoryImpl struct {
	q sqlc.Querier
}

func (c *channelRepositoryImpl) Save(ctx context.Context, channel *Channel) (_ int64, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).Save")

	if err := c.q.ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
		ExternalCustomID: channel.Channel.CustomID,
		ExternalID:       string(channel.Channel.ID),
	}); err != nil {
		return 0, err
	}

	row, err := c.q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
		ExternalID:                string(channel.Channel.ID),
		ExternalDisplayName:       channel.Channel.DisplayName,
		ExternalCustomID:          channel.Channel.CustomID,
		ExternalIconUrl:           channel.Channel.IconURL,
		ExternalDescription:       channel.Channel.Description,
		ExternalSubscribersCount:  int64(channel.Channel.SubscribersCount),
		ExternalCreatedAt:         channel.Channel.CreatedAt,
		ExternalUploadsPlaylistID: string(channel.Channel.UploadsPlaylistID),
		PublicID:                  channel.ID,
		RssFetchedAt:              channel.RSSFetchedAt,
		FetchedAt:                 channel.FetchedAt,
		BulkFetchedAt:             channel.BulkFetchedAt,
	})
	if err != nil {
		return 0, err
	}

	channel.ID = row.PublicID

	return row.MChannelID, nil
}

func (c *channelRepositoryImpl) FindForUpdate(ctx context.Context, id uuid.UUID) (_ *Channel, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).FindForUpdate(id=%s)", id)

	row, err := c.q.GetChannelForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	channelDetail, err := youtube_d.NewChannel(
		row.ExternalID,
		row.ExternalDisplayName,
		row.ExternalCustomID,
		row.ExternalDescription,
		row.ExternalIconUrl,
		uint64(row.ExternalSubscribersCount),
		row.ExternalUploadsPlaylistID,
		row.ExternalCreatedAt,
	)
	if err != nil {
		return nil, err
	}

	ch, err := NewChannel(
		row.FetchedAt,
		row.RssFetchedAt,
		channelDetail,
		WithChannelID(row.PublicID),
		WithBulkFetchedAt(row.BulkFetchedAt),
	)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (c *channelRepositoryImpl) FindByIdOrHandle(ctx context.Context, idOrHandle string) (_ *Channel, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).FindByIdOrHandle")

	row, err := c.q.FindChannelByExternalID(ctx, sqlc.FindChannelByExternalIDParams{
		ExternalID:       idOrHandle,
		ExternalCustomID: idOrHandle,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	ytCh, err := youtube_d.NewChannel(
		row.ExternalID,
		row.ExternalDisplayName,
		row.ExternalCustomID,
		row.ExternalDescription,
		row.ExternalIconUrl,
		uint64(row.ExternalSubscribersCount), // NOTE: bigint(sql) -> int64 -> uint64
		row.ExternalUploadsPlaylistID,
		row.ExternalCreatedAt,
	)
	if err != nil {
		return nil, err
	}

	ch, err := NewChannel(row.FetchedAt, row.RssFetchedAt, ytCh, WithChannelID(row.PublicID), WithBulkFetchedAt(row.BulkFetchedAt))
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (c *channelRepositoryImpl) FindToFetchRSSForUpdate(ctx context.Context, userID uuid.UUID, rssFetchDuration time.Duration, limit int32) (_ []*Channel, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).FindToFetchRSSForUpdate(userID=%s)", userID)

	rows, err := c.q.ListStaleRSSChannelsForUpdate(ctx, sqlc.ListStaleRSSChannelsForUpdateParams{
		UserID:     userID,
		RssFetch:   time.Now().UTC().Add(-rssFetchDuration),
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	channels := make([]*Channel, len(rows))
	for i, row := range rows {
		channelDetail, err := youtube_d.NewChannel(
			row.ExternalID,
			row.ExternalDisplayName,
			row.ExternalCustomID,
			row.ExternalDescription,
			row.ExternalIconUrl,
			uint64(row.ExternalSubscribersCount),
			row.ExternalUploadsPlaylistID,
			row.ExternalCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		channel, err := NewChannel(
			row.FetchedAt,
			row.RssFetchedAt,
			channelDetail,
			WithChannelID(row.PublicID),
			WithBulkFetchedAt(row.BulkFetchedAt),
		)
		if err != nil {
			return nil, err
		}

		channels[i] = channel
	}

	return channels, nil
}

func (c *channelRepositoryImpl) FindBulkFetchedBefore(ctx context.Context, before time.Time) (_ []*Channel, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).FindBulkFetchedBefore")

	rows, err := c.q.ListChannelsBulkFetchedBefore(ctx, before)
	if err != nil {
		return nil, err
	}

	channels := make([]*Channel, len(rows))
	for i, row := range rows {
		channelDetail, err := youtube_d.NewChannel(
			row.ExternalID,
			row.ExternalDisplayName,
			row.ExternalCustomID,
			row.ExternalDescription,
			row.ExternalIconUrl,
			uint64(row.ExternalSubscribersCount),
			row.ExternalUploadsPlaylistID,
			row.ExternalCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		channel, err := NewChannel(
			row.FetchedAt,
			row.RssFetchedAt,
			channelDetail,
			WithChannelID(row.PublicID),
			WithBulkFetchedAt(row.BulkFetchedAt),
		)
		if err != nil {
			return nil, err
		}

		channels[i] = channel
	}

	return channels, nil
}

func (s *channelRepositoryImpl) SaveSubscription(ctx context.Context, subscribedChannel *SubscribedChannel) (_ int64, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).SaveSubscription")

	id, err := s.q.InsertSubscription(ctx, sqlc.InsertSubscriptionParams{
		UserPublicID: subscribedChannel.SubscriberID,
		ChannelID:    subscribedChannel.ChannelID,
		SubscribedAt: subscribedChannel.SubscribedAt,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrTooManySubscriptions
		}
		return 0, err
	}

	return id, nil
}

func (s *channelRepositoryImpl) RemoveSubscription(ctx context.Context, userID, channelID uuid.UUID) (rowAffected int64, err error) {
	defer util.Wrap(&err, "channel.(*channelRepositoryImpl).RemoveSubscription(userID=%s, channelID=%s)", userID, channelID)

	row, err := s.q.DeleteSubscription(ctx, sqlc.DeleteSubscriptionParams{
		UserPublicID: userID,
		ChannelID:    channelID,
	})
	if err != nil {
		return 0, err
	}

	return row, nil
}

var _ ChannelRepository = (*channelRepositoryImpl)(nil)

type ValuableChannelRepository interface {
	Save(ctx context.Context, vc *ValuableChannel) (int64, error)
	Remove(ctx context.Context, channelID uuid.UUID) error
	FindForUpdate(ctx context.Context, channelID uuid.UUID) (*ValuableChannel, error)
}

type valuableChannelRepositoryImpl struct {
	q sqlc.Querier
}

func NewValuableChannelRepository(q sqlc.Querier) ValuableChannelRepository {
	return &valuableChannelRepositoryImpl{q: q}
}

func (v *valuableChannelRepositoryImpl) Save(ctx context.Context, vc *ValuableChannel) (_ int64, err error) {
	defer util.Wrap(&err, "channel.(*valuableChannelRepositoryImpl).Save(channelID=%s)", vc.ChannelID)

	id, err := v.q.UpsertValuableChannel(ctx, sqlc.UpsertValuableChannelParams{
		ChannelPublicID:     vc.ChannelID,
		CategoryCode:        int(vc.ValuableReasonCode),
		ValuableDescription: vc.ValuableDescription.String(),
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (v *valuableChannelRepositoryImpl) Remove(ctx context.Context, channelID uuid.UUID) (err error) {
	defer util.Wrap(&err, "channel.(*valuableChannelRepositoryImpl).Remove(channelID=%s)", channelID)

	if err := v.q.DeleteValuableChannel(ctx, channelID); err != nil {
		return err
	}
	return nil
}

func (v *valuableChannelRepositoryImpl) FindForUpdate(ctx context.Context, channelID uuid.UUID) (_ *ValuableChannel, err error) {
	defer util.Wrap(&err, "channel.(*valuableChannelRepositoryImpl).FindForUpdate(channelID=%s)", channelID)

	row, err := v.q.GetValuableChannelForUpdate(ctx, channelID)
	if err != nil {
		return nil, err
	}
	return &ValuableChannel{
		ChannelID:           row.ChannelPublicID,
		ValuableReasonCode:  ValuableCategoryCode(row.CategoryCode),
		ValuableDescription: ValuableDescription(row.ValuableDescription),
	}, nil
}

var _ ValuableChannelRepository = (*valuableChannelRepositoryImpl)(nil)
