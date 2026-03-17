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
		items[i].ExternalVideoId = item.ExternalVideoID
		items[i].ExternalVideoTitle = item.ExternalVideoTitle
		items[i].ExternalVideoThumbnailUrl = item.ExternalVideoThumbnailURL
		items[i].ExternalVideoLengthSeconds = item.ExternalVideoLengthSeconds
		items[i].WatchPositionSeconds = item.WatchPositionSeconds
		items[i].WatchedAt = item.WatchedAt
		items[i].ChannelId = item.ChannelID
		items[i].ExternalChannelId = item.ExternalChannelID
		items[i].ExternalChannelDisplayName = item.ExternalChannelDisplayName
		items[i].ExternalChannelIconUrl = item.ExternalChannelIconURL
	}

	return GetHistory200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetStatisticsDaily(c context.Context, request GetStatisticsDailyRequestObject) (GetStatisticsDailyResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *APIHandler) GetStatisticsMonthly(c context.Context, request GetStatisticsMonthlyRequestObject) (GetStatisticsMonthlyResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
