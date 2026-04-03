package channel

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

var (
	ErrInvalidValuableDescription = core.NewDomainError("valuable_channel.invalid_valuable_reason", "invalid valuable description")
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

func NewSubscribedChannel(channelID, subscriberID uuid.UUID, opts ...SubscribedChannelOption) (_ *SubscribedChannel, err error) {
	defer util.Wrap(&err, "channel.NewSubscribedChannel")

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
	ID            uuid.UUID
	FetchedAt     time.Time
	RSSFetchedAt  time.Time
	BulkFetchedAt time.Time
	Channel       youtube_d.Channel
}

type ChannelOption func(*Channel)

func WithChannelID(id uuid.UUID) ChannelOption {
	return func(c *Channel) {
		c.ID = id
	}
}

func WithBulkFetchedAt(bulkFetchedAt time.Time) ChannelOption {
	return func(c *Channel) {
		c.BulkFetchedAt = bulkFetchedAt
	}
}

func NewChannel(fetchedAt, rssFetchedAt time.Time, channel youtube_d.Channel, opts ...ChannelOption) (_ *Channel, err error) {
	defer util.Wrap(&err, "channel.NewChannel")

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	c := &Channel{
		ID:            id,
		FetchedAt:     fetchedAt,
		RSSFetchedAt:  rssFetchedAt,
		BulkFetchedAt: time.Now().UTC().AddDate(-1, 0, 0),
		Channel:       channel,
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

func (c *Channel) MarkAsBulkFetched() {
	c.BulkFetchedAt = time.Now().UTC()
}

type ValuableDescription string

func NewValuableDescription(description string) (_ ValuableDescription, err error) {
	defer util.Wrap(&err, "channel.NewValuableDescription")

	if len(description) >= 256 {
		return "", ErrInvalidValuableDescription
	}

	return ValuableDescription(description), nil
}

func (v ValuableDescription) String() string {
	return string(v)
}

type ValuableCategoryCode int

var (
	ErrInvalidValuableCategoryCode = core.NewDomainError("valuable_channel.invalid_valuable_category_code", "invalid valuable category code")

	valuableCategoryCodeMap = []struct {
		code ValuableCategoryCode
		str  string
	}{
		{
			code: 0,
			str:  "unknown",
		},
		{
			code: 1,
			str:  "education",
		},
		{
			code: 2,
			str:  "technology",
		},
		{
			code: 3,
			str:  "economy",
		},
		{
			code: 4,
			str:  "politics",
		},
		{
			code: 5,
			str:  "music",
		},
	}
)

func NewValuableCategoryCode(str string) (_ ValuableCategoryCode, err error) {
	defer util.Wrap(&err, "channel.NewValuableCategoryCode")

	for _, c := range valuableCategoryCodeMap {
		if str == c.str {
			return c.code, nil
		}
	}

	return -1, ErrInvalidValuableCategoryCode
}

func (v ValuableCategoryCode) String() string {
	for _, c := range valuableCategoryCodeMap {
		if c.code == v {
			return c.str
		}
	}

	return "unknown"
}

type ValuableChannel struct {
	ChannelID           uuid.UUID
	ValuableReasonCode  ValuableCategoryCode
	ValuableDescription ValuableDescription
}

// これEntityでいいのかな...?
func NewValuableChannel(channelID uuid.UUID, reasonCode, valuableDescription string) (_ *ValuableChannel, err error) {
	defer util.Wrap(&err, "channel.NewValuableChannel")

	rc, err := NewValuableCategoryCode(reasonCode)
	if err != nil {
		return nil, err
	}

	description, err := NewValuableDescription(valuableDescription)
	if err != nil {
		return nil, err
	}

	return &ValuableChannel{
		ChannelID:           channelID,
		ValuableReasonCode:  rc,
		ValuableDescription: description,
	}, nil
}

func (vc *ValuableChannel) SetReasonCode(reasonCode *string) (err error) {
	if reasonCode == nil {
		return nil
	}
	defer util.Wrap(&err, "channel.(*ValuableChannel).SetReasonCode")

	rc, err := NewValuableCategoryCode(*reasonCode)
	if err != nil {
		return err
	}
	vc.ValuableReasonCode = rc
	return nil
}

func (vc *ValuableChannel) SetDescription(description *string) (err error) {
	if description == nil {
		return nil
	}
	defer util.Wrap(&err, "channel.(*ValuableChannel).SetDescription")

	d, err := NewValuableDescription(*description)
	if err != nil {
		return err
	}
	vc.ValuableDescription = d
	return nil
}
