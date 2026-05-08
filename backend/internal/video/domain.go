package video

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

type Video struct {
	ID        uuid.UUID
	ChannelID uuid.UUID
	FetchedAt time.Time

	Video youtube_d.Video
}

func NewVideo(
	channelID uuid.UUID,
	fetchedAt time.Time,

	video youtube_d.Video,
) (_ *Video, err error) {
	defer util.Wrap(&err, "video.NewVideo(channelID=%s)", channelID)

	id, err := util.NewUUIDv7WithTime(video.CreatedAt)
	if err != nil {
		return nil, err
	}

	return new(Video{
		ID:        id,
		ChannelID: channelID,
		Video:     video,
		FetchedAt: fetchedAt,
	}), nil
}
