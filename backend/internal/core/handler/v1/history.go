package v1

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetHistory(ctx context.Context, request GetHistoryRequestObject) (GetHistoryResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	views, hasNext, err := h.historyService.GetHistory(ctx, userID, request.Params.Limit, request.Params.Cursor, hutil.TimezoneFromContext(ctx))
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ChannelId                  uuid.UUID `json:"channel_id"`
		ExternalChannelDisplayName string    `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
		ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string    `json:"external_video_title"`
		VideoId                    uuid.UUID `json:"video_id"`
		WatchId                    uuid.UUID `json:"watch_id"`
		WatchPositionSeconds       int       `json:"watch_position_seconds"`
		WatchedAt                  time.Time `json:"watched_at"`
	}, len(views))

	for i, v := range views {
		items[i].WatchId = v.WatchId
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
		return nil, err
	}

	remaining, err := h.historyService.Heartbeat(ctx, userID, request.VideoId, request.Body.CurrentPositionSeconds, request.Body.PlaylistId, hutil.TimezoneFromContext(ctx))
	if err != nil {
		return nil, err
	}

	return PostVideosVideoIdHeartbeats200JSONResponse{
		DailyRemainingSeconds: remaining,
	}, nil
}

func (h *APIHandler) PostVideosVideoIdWatched(ctx context.Context, request PostVideosVideoIdWatchedRequestObject) (PostVideosVideoIdWatchedResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.historyService.MarkVideoWatched(ctx, userID, request.VideoId); err != nil {
		return nil, err
	}

	return PostVideosVideoIdWatched201Response{}, nil
}

func (h *APIHandler) DeleteVideosVideoIdWatched(ctx context.Context, request DeleteVideosVideoIdWatchedRequestObject) (DeleteVideosVideoIdWatchedResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.historyService.UnmarkVideoWatched(ctx, userID, request.VideoId); err != nil {
		return nil, err
	}

	return DeleteVideosVideoIdWatched204Response{}, nil
}

func (h *APIHandler) GetStatisticsWeekly(ctx context.Context, request GetStatisticsWeeklyRequestObject) (GetStatisticsWeeklyResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	loc := hutil.TimezoneFromContext(ctx)
	targetWeek := request.Params.TargetWeek.In(loc)

	aiSummary, views, err := h.historyService.GetStatisticsByWeek(ctx, userID, targetWeek, loc)
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		TargetDay         time.Time `json:"target_day"`
		VideoWatchCount   int       `json:"video_watch_count"`
		VideoWatchSeconds int       `json:"video_watch_seconds"`
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
		AiSummary:  aiSummary,
	}, nil
}
