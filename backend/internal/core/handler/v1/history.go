package v1

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetHistory(ctx context.Context, request GetHistoryRequestObject) (GetHistoryResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetHistory500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	views, hasNext, err := h.historyService.GetHistory(ctx, userID, request.Params.Limit, request.Params.Cursor)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetHistory500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		ChannelId                  uuid.UUID `json:"channel_id"`
		ExternalChannelDisplayName string             `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string             `json:"external_channel_icon_url"`
		ExternalVideoLengthSeconds int                `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string             `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string             `json:"external_video_title"`
		VideoId                    uuid.UUID `json:"video_id"`
		WatchPositionSeconds       int                `json:"watch_position_seconds"`
		WatchedAt                  time.Time          `json:"watched_at"`
	}, len(views))

	for i, v := range views {
		items[i].VideoId = v.VideoId
		items[i].ExternalVideoTitle = v.ExternalVideoTitle
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		items[i].WatchPositionSeconds = v.WatchPositionSeconds
		items[i].WatchedAt = v.WatchedAt
		items[i].ChannelId = v.ChannelId
		items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayName
		items[i].ExternalChannelIconUrl = v.ExternalChannelIconUrl
	}

	return GetHistory200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) PostVideosVideoIdHeartbeats(ctx context.Context, request PostVideosVideoIdHeartbeatsRequestObject) (PostVideosVideoIdHeartbeatsResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return PostVideosVideoIdHeartbeats500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	remaining, err := h.historyService.Heartbeat(ctx, userID, request.VideoId, request.Body.CurrentPositionSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostVideosVideoIdHeartbeats404JSONResponse{
				NotFoundJSONResponse: NotFoundJSONResponse{
					Detail: "video not found",
					Title:  "Not Found",
				},
			}, nil
		}
		hutil.LogError(ctx, err)
		return PostVideosVideoIdHeartbeats500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostVideosVideoIdHeartbeats200JSONResponse{
		DailyRemainingSeconds: remaining,
	}, nil
}

func (h *APIHandler) GetStatisticsWeekly(ctx context.Context, request GetStatisticsWeeklyRequestObject) (GetStatisticsWeeklyResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetStatisticsWeekly500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	targetWeek := request.Params.TargetWeek

	views, err := h.historyService.GetStatisticsByWeek(ctx, userID, targetWeek)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetStatisticsWeekly500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		TargetDay         time.Time `json:"target_day"`
		VideoWatchCount   int                `json:"video_watch_count"`
		VideoWatchSeconds int                `json:"video_watch_seconds"`
	}, len(views))

	for i, v := range views {
		items[i].TargetDay = v.WatchDate
		items[i].VideoWatchCount = v.VideoCount
		items[i].VideoWatchSeconds = int(v.WatchSum)
	}

	return GetStatisticsWeekly200JSONResponse{
		TargetWeek: targetWeek,
		ItemCount:  len(items),
		Items:      items,
	}, nil
}
