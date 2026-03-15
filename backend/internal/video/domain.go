package video

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrResourceTypeInvalidString = errors.New("invalid resource type string(video or channel)")

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
