package channel

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrChannelCustomIdTooShort           = errors.New("the channel custom id(handle) is too short(<=3)")
	ErrChannelCustomIdShouldStartsWithAt = errors.New("the channel custom id should starts with @")

	ErrChannelIdShouldStartsWithUC = errors.New("the channel id should starts with UC")
	ErrInvalidChannelIdLength      = errors.New("invalid channel id length (should be 24)")
)

type ChannelCustomId string

func NewChannelCustomId(id string) (*ChannelCustomId, error) {
	if !strings.HasPrefix(id, "@") {
		return nil, ErrChannelCustomIdShouldStartsWithAt
	}
	if len([]rune(id)) <= 3 {
		return nil, ErrChannelCustomIdTooShort
	}

	c := ChannelCustomId(id)
	return &c, nil
}

type ChannelId string

func NewChannelId(id string) (*ChannelId, error) {
	if !strings.HasPrefix(id, "UC") {
		return nil, ErrChannelIdShouldStartsWithUC
	}
	if len(id) != 24 {
		return nil, ErrInvalidChannelIdLength
	}

	c := ChannelId(id)
	return &c, nil
}

type ExternalChannelInfo struct {
	Id               *ChannelId
	DisplayName      string
	CustomId         *ChannelCustomId
	Description      string
	IconUrl          string
	SubscribersCount int
	CreatedAt        time.Time
}

type SubscribedChannel struct {
	SubscriptionId uuid.UUID
	ChannelId      uuid.UUID
	CreatedAt      time.Time
	ExternalChannelInfo
}

func NewSubscribedChannel(
	subscriptionId uuid.UUID,
	channelId uuid.UUID,
	createdAt time.Time,
	extId,
	extDisplayname,
	extCustomId,
	extDescription,
	extIconUrl string,
	extSubscribersCnt int,
	extCreatedAt time.Time,
) (*SubscribedChannel, error) {
	chId, err := NewChannelId(extId)
	if err != nil {
		return nil, err
	}
	channelCustomId, err := NewChannelCustomId(extCustomId)
	if err != nil {
		return nil, err
	}

	return &SubscribedChannel{
		SubscriptionId: subscriptionId,
		ChannelId:      channelId,
		CreatedAt:      createdAt,
		ExternalChannelInfo: ExternalChannelInfo{
			Id:               chId,
			DisplayName:      extDisplayname,
			CustomId:         channelCustomId,
			Description:      extDescription,
			IconUrl:          extIconUrl,
			SubscribersCount: extSubscribersCnt,
			CreatedAt:        extCreatedAt,
		},
	}, nil
}
