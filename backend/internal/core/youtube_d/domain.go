package youtube_d

import (
	"strings"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

var (
	ErrChannelIDEmpty              = core.NewDomainError("youtube.channel_id_empty", "channel ID must not be empty")
	ErrChannelIDInvalidPrefix      = core.NewDomainError("youtube.channel_id_invalid_prefix", "channel ID must start with 'UC'")
	ErrChannelUploadsPlaylistEmpty = core.NewDomainError("youtube.channel_uploads_playlist_empty", "channel uploads playlist ID must not be empty")
	ErrVideoIDEmpty                = core.NewDomainError("youtube.video_id_empty", "video ID must not be empty")
	ErrVideoTooShort               = core.NewDomainError("youtube.video_too_short", "video must be longer than 60 seconds")
)

type ChannelID string

func NewChannelID(id string) (_ ChannelID, err error) {
	defer util.Wrap(&err, "youtube_d.NewChannelID")

	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrChannelIDEmpty
	}
	if !strings.HasPrefix(id, "UC") {
		return "", ErrChannelIDInvalidPrefix
	}
	return ChannelID(id), nil
}

func (c ChannelID) String() string {
	return string(c)
}

type ChannelUploadsPlaylistID string

func NewChannelUploadsPlaylistID(id string) (_ ChannelUploadsPlaylistID, err error) {
	defer util.Wrap(&err, "youtube_d.NewChannelUploadsPlaylistID")

	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrChannelUploadsPlaylistEmpty
	}
	return ChannelUploadsPlaylistID(id), nil
}

func (c ChannelUploadsPlaylistID) String() string {
	return string(c)
}

type Channel struct {
	ID                ChannelID
	DisplayName       string
	CustomID          string
	Description       string
	IconURL           string
	SubscribersCount  uint64
	UploadsPlaylistID ChannelUploadsPlaylistID
	CreatedAt         time.Time
}

func NewChannel(id, displayName, customID, description, iconURL string, subscribersCount uint64, uploadsPlaylistID string, createdAt time.Time) (_ Channel, err error) {
	defer util.Wrap(&err, "youtube_d.NewChannel")

	channelID, err := NewChannelID(id)
	if err != nil {
		return Channel{}, err
	}

	channelUploadsPlaylistID, err := NewChannelUploadsPlaylistID(uploadsPlaylistID)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ID:                channelID,
		DisplayName:       displayName,
		CustomID:          customID,
		Description:       description,
		IconURL:           iconURL,
		SubscribersCount:  subscribersCount,
		UploadsPlaylistID: channelUploadsPlaylistID,
		CreatedAt:         createdAt,
	}, nil
}

type Playlist struct {
	ID          string
	Title       string
	Description string
	CreatedAt   time.Time
}

type WatchHistory struct {
	VideoID     VideoID
	WatchedAt   time.Time
}

type VideoID string

func NewVideoID(id string) (_ VideoID, err error) {
	defer util.Wrap(&err, "youtube_d.NewVideoID")

	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrVideoIDEmpty
	}
	return VideoID(id), nil
}

func (v VideoID) String() string {
	return string(v)
}

type Video struct {
	ID            VideoID
	ChannelID     ChannelID
	Title         string
	Description   string
	ThumbnailURL  string
	LengthSeconds int
	CreatedAt     time.Time
}

func NewVideo(id, channelID, title, description, thumbnailURL string, lengthSeconds int, createdAt time.Time) (_ Video, err error) {
	defer util.Wrap(&err, "youtube_d.NewVideo")

	videoID, err := NewVideoID(id)
	if err != nil {
		return Video{}, err
	}

	channelIDd, err := NewChannelID(channelID)
	if err != nil {
		return Video{}, err
	}

	if lengthSeconds <= 60 {
		return Video{}, ErrVideoTooShort
	}

	return Video{
		ID:            videoID,
		ChannelID:     channelIDd,
		Title:         title,
		Description:   description,
		ThumbnailURL:  thumbnailURL,
		LengthSeconds: lengthSeconds,
		CreatedAt:     createdAt,
	}, nil
}
