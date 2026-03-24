package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChannelRepository interface {
	Save(ctx context.Context, channel *Channel) (int64, error)
	FindForUpdate(ctx context.Context, id uuid.UUID) (*Channel, error)
	FindByIdOrHandle(ctx context.Context, idOrHandle string) (*Channel, error)
	FindToFetchRSSForUpdate(ctx context.Context, userID uuid.UUID, rssFetchDuration time.Duration, limit int32) ([]*Channel, error)
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

func (c *channelRepositoryImpl) Save(ctx context.Context, channel *Channel) (int64, error) {
	if err := c.q.ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
		ExternalCustomID: channel.Channel.CustomID,
		ExternalID:       string(channel.Channel.ID),
	}); err != nil {
		return 0, fmt.Errorf("failed to clearStaleChannelCustomID(channelRepository.Save): %w", err)
	}

	saveChannel, err := c.q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
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
	})
	if err != nil {
		return 0, fmt.Errorf("failed to saveChannel(channelRepository.Save): %w", err)
	}

	return saveChannel, nil
}

func (c *channelRepositoryImpl) FindForUpdate(ctx context.Context, id uuid.UUID) (*Channel, error) {
	row, err := c.q.GetChannelForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to getChannelForUpdate(channelRepository.FindForUpdate): %w", err)
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
		return nil, fmt.Errorf("failed to newChannel(channelRepository.FindForUpdate): %w", err)
	}

	ch, err := NewChannel(row.FetchedAt, row.RssFetchedAt, channelDetail, WithChannelID(row.PublicID))
	if err != nil {
		return nil, fmt.Errorf("failed to newChannel(channelRepository.FindForUpdate): %w", err)
	}

	return ch, nil
}

func (c *channelRepositoryImpl) FindByIdOrHandle(ctx context.Context, idOrHandle string) (*Channel, error) {
	row, err := c.q.FindChannelByExternalID(ctx, sqlc.FindChannelByExternalIDParams{
		ExternalID:       idOrHandle,
		ExternalCustomID: idOrHandle,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to getChannelByIdOrHandle: %w", err)
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
		return nil, fmt.Errorf("failed to newChannel: %w", err)
	}

	ch, err := NewChannel(row.FetchedAt, row.RssFetchedAt, ytCh, WithChannelID(row.PublicID))
	if err != nil {
		return nil, fmt.Errorf("failed to newChannel: %w", err)
	}

	return ch, nil
}

func (c *channelRepositoryImpl) FindToFetchRSSForUpdate(ctx context.Context, userID uuid.UUID, rssFetchDuration time.Duration, limit int32) ([]*Channel, error) {
	rows, err := c.q.ListStaleRSSChannelsForUpdate(ctx, sqlc.ListStaleRSSChannelsForUpdateParams{
		UserID:     userID,
		RssFetch:   time.Now().UTC().Add(-rssFetchDuration),
		QueryLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to getChannelsToFetchRSSForUpdate(channelRepository.FindToFetchRSSForUpdate): %w", err)
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
			return nil, fmt.Errorf("failed to newChannel(channelRepository.FindToFetchRSSForUpdate): %w", err)
		}

		channel, err := NewChannel(row.FetchedAt, row.RssFetchedAt, channelDetail, WithChannelID(row.PublicID))
		if err != nil {
			return nil, fmt.Errorf("failed to newChannel(channelRepository.FindToFetchRSSForUpdate): %w", err)
		}

		channels[i] = channel
	}

	return channels, nil
}

func (s *channelRepositoryImpl) SaveSubscription(ctx context.Context, subscribedChannel *SubscribedChannel) (int64, error) {
	id, err := s.q.InsertSubscription(ctx, sqlc.InsertSubscriptionParams{
		UserPublicID: subscribedChannel.SubscriberID,
		ChannelID:    subscribedChannel.ChannelID,
		SubscribedAt: subscribedChannel.SubscribedAt,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to saveChannelSubscription(channelRepository.SaveSubscription): %w", err)
	}

	return id, nil
}

func (s *channelRepositoryImpl) RemoveSubscription(ctx context.Context, userID, channelID uuid.UUID) (rowAffected int64, err error) {
	row, err := s.q.DeleteSubscription(ctx, sqlc.DeleteSubscriptionParams{
		UserPublicID: userID,
		ChannelID:    channelID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to deleteChannelSubscription(channelRepository.RemoveSubscription): %w", err)
	}

	return row, nil
}

var _ ChannelRepository = (*channelRepositoryImpl)(nil)
