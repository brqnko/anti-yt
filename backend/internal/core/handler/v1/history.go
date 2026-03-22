package v1

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *APIHandler) GetHistory(c context.Context, request GetHistoryRequestObject) (GetHistoryResponseObject, error) {
	historyItems, hasNext, err := h.historyService.GetHistory(c, request.Params.Limit, request.Params.Cursor)
	if err != nil {
		util.LogError(c, err)
		return GetHistory500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		ChannelId                  openapi_types.UUID `json:"channel_id"`
		ExternalChannelDisplayName string             `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string             `json:"external_channel_icon_url"`
		ExternalChannelId          string             `json:"external_channel_id"`
		ExternalVideoCreatedAt     time.Time          `json:"external_video_created_at"`
		ExternalVideoId            string             `json:"external_video_id"`
		ExternalVideoLengthSeconds int                `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string             `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string             `json:"external_video_title"`
		VideoId                    openapi_types.UUID `json:"video_id"`
		WatchPositionSeconds       int                `json:"watch_position_seconds"`
		WatchedAt                  time.Time          `json:"watched_at"`
	}, len(historyItems))

	for i, item := range historyItems {
		items[i].VideoId = item.VideoID
		items[i].ExternalVideoId = string(item.ExternalVideoID)
		items[i].ExternalVideoTitle = item.ExternalVideoTitle
		items[i].ExternalVideoThumbnailUrl = item.ExternalVideoThumbnailURL
		items[i].ExternalVideoLengthSeconds = item.ExternalVideoLengthSeconds
		items[i].ExternalVideoCreatedAt = item.ExternalVideoCreatedAt
		items[i].WatchPositionSeconds = item.WatchPositionSeconds
		items[i].WatchedAt = item.WatchedAt
		items[i].ChannelId = item.ChannelID
		items[i].ExternalChannelId = string(item.ExternalChannelID)
		items[i].ExternalChannelDisplayName = item.ExternalChannelDisplayName
		items[i].ExternalChannelIconUrl = item.ExternalChannelIconURL
	}

	return GetHistory200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetStatisticsWeekly(c context.Context, request GetStatisticsWeeklyRequestObject) (GetStatisticsWeeklyResponseObject, error) {
	targetWeek := request.Params.TargetWeek.Time

	stats, err := h.historyService.GetStatisticsByWeek(c, targetWeek)
	if err != nil {
		util.LogError(c, err)
		return GetStatisticsWeekly500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		TargetDay         openapi_types.Date `json:"target_day"`
		VideoWatchCount   int                `json:"video_watch_count"`
		VideoWatchSeconds int                `json:"video_watch_seconds"`
	}, len(stats.DailyBreakdown))

	for i, d := range stats.DailyBreakdown {
		items[i].TargetDay = openapi_types.Date{Time: d.WatchDate}
		items[i].VideoWatchCount = d.VideoCount
		items[i].VideoWatchSeconds = int(d.WatchSum)
	}

	return GetStatisticsWeekly200JSONResponse{
		TargetWeek: openapi_types.Date{Time: targetWeek},
		ItemCount:  len(items),
		Items:      items,
		AiSummary: func() *string {
			if stats.AIComment == "" {
				return nil
			}
			return &stats.AIComment
		}(),
	}, nil
}
