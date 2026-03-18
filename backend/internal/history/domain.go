package history

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
)

type HistoryItem struct {
	VideoID                    uuid.UUID
	ExternalVideoID            *video.ExternalVideoID
	ExternalVideoTitle         string
	ExternalVideoThumbnailURL  string
	ExternalVideoLengthSeconds int
	ExternalVideoCreatedAt     time.Time
	WatchPositionSeconds       int
	WatchedAt                  time.Time
	ChannelID                  uuid.UUID
	ExternalChannelID          *channel.ChannelID
	ExternalChannelDisplayName string
	ExternalChannelIconURL     string
}

func NewHistoryItem(
	videoID uuid.UUID,
	externalVideoID string,
	externalVideoTitle string,
	externalVideoThumbnailURL string,
	externalVideoLengthSeconds int,
	externalVideoCreatedAt time.Time,
	watchPositionSeconds int,
	watchedAt time.Time,
	channelID uuid.UUID,
	externalChannelID string,
	externalChannelDisplayName string,
	externalChannelIconURL string,
) (*HistoryItem, error) {
	extVideoID, err := video.NewExternalVideoID(externalVideoID)
	if err != nil {
		return nil, err
	}

	extChannelID, err := channel.NewChannelID(externalChannelID)
	if err != nil {
		return nil, err
	}

	return &HistoryItem{
		VideoID:                    videoID,
		ExternalVideoID:            extVideoID,
		ExternalVideoTitle:         externalVideoTitle,
		ExternalVideoThumbnailURL:  externalVideoThumbnailURL,
		ExternalVideoLengthSeconds: externalVideoLengthSeconds,
		ExternalVideoCreatedAt:     externalVideoCreatedAt,
		WatchPositionSeconds:       watchPositionSeconds,
		WatchedAt:                  watchedAt,
		ChannelID:                  channelID,
		ExternalChannelID:          extChannelID,
		ExternalChannelDisplayName: externalChannelDisplayName,
		ExternalChannelIconURL:     externalChannelIconURL,
	}, nil
}

type DailyStatistics struct {
	WatchDate  time.Time
	VideoCount int
	WatchSum   int64 // seconds
}

func NewDailyStatistics(watchDate time.Time, videoCount int64, watchSum int64) *DailyStatistics {
	return &DailyStatistics{
		WatchDate:  watchDate,
		VideoCount: int(videoCount),
		WatchSum:   watchSum,
	}
}

type WeeklyStatistics struct {
	StartDate      time.Time
	DailyBreakdown []*DailyStatistics
	AIComment      *string
}

func NewWeeklyStatistics(startDate time.Time, daily []*DailyStatistics, aiComment *string) *WeeklyStatistics {
	return &WeeklyStatistics{
		StartDate:      startDate,
		DailyBreakdown: daily,
		AIComment:      aiComment,
	}
}
