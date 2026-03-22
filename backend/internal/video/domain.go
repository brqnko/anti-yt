package video

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrResourceTypeInvalidString = errors.New("invalid resource type string(video or channel)")
	ErrInvalidExternalVideoID    = errors.New("invalid external video id")
)

const (
	ResourceTypeVideo   = "video"
	ResourceTypeChannel = "channel"
)

type ResourceType string

func NewResourceType(str string) (ResourceType, error) {
	if str != ResourceTypeChannel && str != ResourceTypeVideo {
		return "", ErrResourceTypeInvalidString
	}
	return ResourceType(str), nil
}

type Video struct {
	ID                         uuid.UUID
	ExternalVideoThumbnailURL  string
	ExternalVideoTitle         string
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
	LastWatchSeconds           int

	ChannelID                  uuid.UUID
	ExternalChannelIconURL     string
	ExternalChannelDisplayname string
}

func NewVideo(
	id uuid.UUID,
	channelID uuid.UUID,
	externalVideoThumbnailURL string,
	externalChannelIconURL string,
	externalVideoTitle string,
	externalChannelDisplayname string,
	externalVideoCreatedAt time.Time,
	externalVideoLengthSeconds int,
	lastWatchSeconds int,
) Video {
	return Video{
		ID:                         id,
		ChannelID:                  channelID,
		ExternalVideoThumbnailURL:  externalVideoThumbnailURL,
		ExternalChannelIconURL:     externalChannelIconURL,
		ExternalVideoTitle:         externalVideoTitle,
		ExternalChannelDisplayname: externalChannelDisplayname,
		ExternalVideoCreatedAt:     externalVideoCreatedAt,
		ExternalVideoLengthSeconds: externalVideoLengthSeconds,
		LastWatchSeconds:           lastWatchSeconds,
	}
}

type ExternalVideoID string

func NewExternalVideoID(id string) (ExternalVideoID, error) {
	if len(id) != 11 {
		return "", ErrInvalidExternalVideoID
	}
	return ExternalVideoID(id), nil
}

type VideoDetail struct {
	ID                              uuid.UUID
	ExternalVideoID                 ExternalVideoID
	ExternalVideoTitle              string
	ExternalVideoDescription        string
	ExternalVideoThumbnailURL       string
	ChannelID                       uuid.UUID
	ExternalChannelID               string
	ExternalChannelDisplayName      string
	ChannelCustomID                 string
	ExternalChannelIconURL          string
	ExternalChannelSubscribersCount int
}

func NewVideoDetail(
	id uuid.UUID,
	externalID,
	externalVideoTitle,
	externalVideoDescription,
	externalVideoThumbnailURL string,
	channelID uuid.UUID,
	externalChannelID,
	externalChannelDisplayName,
	channelCustomID,
	externalChannelIconURL string,
	externalChannelSubscribersCount int,
) (VideoDetail, error) {
	extVideoID, err := NewExternalVideoID(externalID)
	if err != nil {
		return VideoDetail{}, err
	}

	return VideoDetail{
		ID:              id,
		ExternalVideoID: extVideoID,
		ExternalVideoTitle:              externalVideoTitle,
		ExternalVideoDescription:        externalVideoDescription,
		ExternalVideoThumbnailURL:       externalVideoThumbnailURL,
		ChannelID:                       channelID,
		ExternalChannelID:               externalChannelID,
		ExternalChannelDisplayName:      externalChannelDisplayName,
		ChannelCustomID:                 channelCustomID,
		ExternalChannelIconURL:          externalChannelIconURL,
		ExternalChannelSubscribersCount: externalChannelSubscribersCount,
	}, nil
}
