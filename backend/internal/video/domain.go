package video

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrResourceTypeInvalidString = errors.New("invalid resource type string(video or channel)")
	ErrInvalidExternalVideoId    = errors.New("invalid external video id")
)

const (
	ResourceTypeVideo   = "video"
	ResourceTypeChannel = "channel"
)

type ResourceType string

func NewResourceType(str string) (*ResourceType, error) {
	if str != ResourceTypeChannel && str != ResourceTypeVideo {
		return nil, ErrResourceTypeInvalidString
	}
	r := ResourceType(str)

	return &r, nil
}

type Video struct {
	Id                         uuid.UUID
	ExternalVideoThumbnailUrl  string
	ExternalVideoTitle         string
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
	LastWatchSeconds           *int

	ChannelID                  uuid.UUID
	ExternalChannelIconUrl     string
	ExternalChannelDisplayname string
}

func NewVideo(
	id uuid.UUID,
	channelId uuid.UUID,
	externalVideoThumbnailUrl string,
	externalChannelIconUrl string,
	externalVideoTitle string,
	externalChannelDisplayname string,
	externalVideoCreatedAt time.Time,
	externalVideoLengthSeconds int,
	lastWatchSeconds int,
) *Video {
	var lastWatchPtr *int
	if lastWatchSeconds != 0 {
		lastWatchPtr = &lastWatchSeconds
	}

	return &Video{
		Id:                         id,
		ChannelID:                  channelId,
		ExternalVideoThumbnailUrl:  externalVideoThumbnailUrl,
		ExternalChannelIconUrl:     externalChannelIconUrl,
		ExternalVideoTitle:         externalVideoTitle,
		ExternalChannelDisplayname: externalChannelDisplayname,
		ExternalVideoCreatedAt:     externalVideoCreatedAt,
		ExternalVideoLengthSeconds: externalVideoLengthSeconds,
		LastWatchSeconds:           lastWatchPtr,
	}
}

type ExternalVideoId string

func NewExternalVideoId(id string) (*ExternalVideoId, error) {
	if len(id) != 11 {
		return nil, ErrInvalidExternalVideoId
	}

	v := ExternalVideoId(id)
	return &v, nil
}

type VideoDetail struct {
	Id                              uuid.UUID
	ExternalVideoId                 *ExternalVideoId
	ExternalVideoTitle              string
	ExternalVideoDescription        string
	ExternalVideoThumbnailUrl       string
	ChannelId                       uuid.UUID
	ExternalChannelId               string
	ExternalChannelDisplayName      string
	ChannelCustomId                 string
	ExternalChannelIconUrl          string
	ExternalChannelSubscribersCount int
}

func NewVideoDetail(
	id uuid.UUID,
	externalId,
	externalVideoTitle,
	externalVideoDescription,
	externalVideoThumbnailUrl string,
	channelId uuid.UUID,
	externalChannelId,
	externalChannelDisplayName,
	channelCustomId,
	externalChannelIconUrl string,
	externalChannelSubscribersCount int,
) (*VideoDetail, error) {
	extVideoId, err := NewExternalVideoId(externalId)
	if err != nil {
		return nil, err
	}

	return &VideoDetail{
		Id:                              id,
		ExternalVideoId:                 extVideoId,
		ExternalVideoTitle:              externalVideoTitle,
		ExternalVideoDescription:        externalVideoDescription,
		ExternalVideoThumbnailUrl:       externalVideoThumbnailUrl,
		ChannelId:                       channelId,
		ExternalChannelId:               externalChannelId,
		ExternalChannelDisplayName:      externalChannelDisplayName,
		ChannelCustomId:                 channelCustomId,
		ExternalChannelIconUrl:          externalChannelIconUrl,
		ExternalChannelSubscribersCount: externalChannelSubscribersCount,
	}, nil
}
