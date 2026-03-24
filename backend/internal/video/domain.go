package video

import (
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/google/uuid"
)

type Video struct {
	ID        uuid.UUID
	ChannelID uuid.UUID
	FetchedAt time.Time

	Video youtube_d.Video
}

type VideoOption func(*Video)

func VideoWithID(id uuid.UUID) VideoOption {
	return func(v *Video) {
		v.ID = id
	}
}

func NewVideo(
	channelID uuid.UUID,
	fetchedAt time.Time,

	video youtube_d.Video,
	opts ...VideoOption,
) (*Video, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid v7(newVideo): %w", err)
	}

	v := Video{
		ID:        id,
		ChannelID: channelID,
		Video:     video,
		FetchedAt: fetchedAt,
	}

	for _, opt := range opts {
		opt(&v)
	}

	return &v, nil
}
