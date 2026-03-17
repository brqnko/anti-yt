package channel

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrChannelCustomIDTooShort           = errors.New("the channel custom id(handle) is too short(<=3)")
	ErrChannelCustomIDShouldStartsWithAt = errors.New("the channel custom id should starts with @")

	ErrChannelIDShouldStartsWithUC = errors.New("the channel id should starts with UC")
	ErrInvalidChannelIDLength      = errors.New("invalid channel id length (should be 24)")
)

type ChannelCustomID string

func NewChannelCustomID(id string) (*ChannelCustomID, error) {
	if !strings.HasPrefix(id, "@") {
		return nil, ErrChannelCustomIDShouldStartsWithAt
	}
	if len([]rune(id)) <= 3 {
		return nil, ErrChannelCustomIDTooShort
	}

	c := ChannelCustomID(id)
	return &c, nil
}

type ChannelID string

func NewChannelID(id string) (*ChannelID, error) {
	if !strings.HasPrefix(id, "UC") {
		return nil, ErrChannelIDShouldStartsWithUC
	}
	if len(id) != 24 {
		return nil, ErrInvalidChannelIDLength
	}

	c := ChannelID(id)
	return &c, nil
}

type ExternalChannelInfo struct {
	ID               *ChannelID
	DisplayName      string
	CustomID         *ChannelCustomID
	Description      string
	IconURL          string
	SubscribersCount int
	CreatedAt        time.Time
}

type SubscribedChannel struct {
	SubscriptionID uuid.UUID
	ChannelID      uuid.UUID
	CreatedAt      time.Time
	ExternalChannelInfo
}

func NewSubscribedChannel(
	subscriptionID uuid.UUID,
	channelID uuid.UUID,
	createdAt time.Time,
	extID,
	extDisplayname,
	extCustomID,
	extDescription,
	extIconURL string,
	extSubscribersCnt int,
	extCreatedAt time.Time,
) (*SubscribedChannel, error) {
	chID, err := NewChannelID(extID)
	if err != nil {
		return nil, err
	}
	channelCustomID, err := NewChannelCustomID(extCustomID)
	if err != nil {
		return nil, err
	}

	return &SubscribedChannel{
		SubscriptionID: subscriptionID,
		ChannelID:      channelID,
		CreatedAt:      createdAt,
		ExternalChannelInfo: ExternalChannelInfo{
			ID:               chID,
			DisplayName:      extDisplayname,
			CustomID:         channelCustomID,
			Description:      extDescription,
			IconURL:          extIconURL,
			SubscribersCount: extSubscribersCnt,
			CreatedAt:        extCreatedAt,
		},
	}, nil
}
