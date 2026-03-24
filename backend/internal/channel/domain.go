package channel

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/google/uuid"
)

type SubscribedChannel struct {
	SubscribedAt time.Time
	SubscriberID uuid.UUID
	ChannelID    uuid.UUID
}

type SubscribedChannelOption func(*SubscribedChannel)

func WithSubscribedChannelSubscribedAt(subscribedAt time.Time) SubscribedChannelOption {
	return func(sc *SubscribedChannel) {
		sc.SubscribedAt = subscribedAt
	}
}

func NewSubscribedChannel(channelID, subscriberID uuid.UUID, opts ...SubscribedChannelOption) (*SubscribedChannel, error) {
	sc := &SubscribedChannel{
		SubscribedAt: time.Now(),
		ChannelID:    channelID,
		SubscriberID: subscriberID,
	}

	for _, opt := range opts {
		opt(sc)
	}

	return sc, nil
}

type Channel struct {
	ID           uuid.UUID
	FetchedAt    time.Time
	RSSFetchedAt time.Time
	Channel      youtube_d.Channel
}

type ChannelOption func(*Channel)

func WithChannelID(id uuid.UUID) ChannelOption {
	return func(c *Channel) {
		c.ID = id
	}
}

func NewChannel(fetchedAt, rssFetchedAt time.Time, channel youtube_d.Channel, opts ...ChannelOption) (*Channel, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	c := &Channel{
		ID:           id,
		FetchedAt:    fetchedAt,
		RSSFetchedAt: rssFetchedAt,
		Channel:      channel,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (c *Channel) ShouldFetchRSSFeed(fetchDuration time.Duration) bool {
	return time.Now().UTC().Sub(c.RSSFetchedAt) > fetchDuration
}

func (c *Channel) MarkAsRSSFetched() {
	c.RSSFetchedAt = time.Now().UTC()
}
