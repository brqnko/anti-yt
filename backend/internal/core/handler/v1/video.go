package v1

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetChannelsChannelIdVideos(ctx context.Context, request GetChannelsChannelIdVideosRequestObject) (GetChannelsChannelIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	videos, hasNext, err := h.videoService.GetChannelUploads(ctx, userID, request.ChannelId, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string    `json:"external_video_title"`
		LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
		VideoId                    uuid.UUID `json:"video_id"`
	}, len(videos))

	for i, v := range videos {
		items[i].VideoId = v.VideoId
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		items[i].ExternalVideoTitle = v.ExternalVideoTitle
		items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt
		items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		items[i].LastWatchSeconds = v.LastWatchSeconds
	}

	return GetChannelsChannelIdVideos200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(videos),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetFeed(ctx context.Context, request GetFeedRequestObject) (GetFeedResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.channelService.SyncRSSFeeds(ctx, userID); err != nil {
		hutil.LogError(ctx, err)
	}

	videos, hasNext, err := h.videoService.GetFeed(ctx, userID, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ChannelId                  uuid.UUID `json:"channel_id"`
		ExternalChannelDisplayName string    `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
		ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string    `json:"external_video_title"`
		LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
		VideoId                    uuid.UUID `json:"video_id"`
	}, len(videos))

	for i, v := range videos {
		items[i].VideoId = v.VideoId
		items[i].ChannelId = v.ChannelId
		items[i].ExternalChannelIconUrl = v.ExternalChannelIconUrl
		items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayName
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		items[i].ExternalVideoTitle = v.ExternalVideoTitle
		items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt
		items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		items[i].LastWatchSeconds = v.LastWatchSeconds
	}

	return GetFeed200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetSearch(ctx context.Context, request GetSearchRequestObject) (GetSearchResponseObject, error) {
	// TODO
	return GetSearch500JSONResponse{
		InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
			Detail: internalErrorDetail,
			Title:  internalErrorTitle,
		},
	}, nil
}

func (h *APIHandler) GetVideosVideoId(ctx context.Context, request GetVideosVideoIdRequestObject) (GetVideosVideoIdResponseObject, error) {
	videoDetail, err := h.videoService.GetVideoDetail(ctx, request.VideoId)
	if err != nil {
		return nil, err
	}

	return GetVideosVideoId200JSONResponse{
		VideoId:                         videoDetail.VideoId,
		ExternalVideoId:                 videoDetail.ExternalVideoId,
		ExternalVideoTitle:              videoDetail.ExternalVideoTitle,
		ExternalVideoDescription:        videoDetail.ExternalVideoDescription,
		ExternalVideoThumbnailUrl:       videoDetail.ExternalVideoThumbnailUrl,
		ChannelId:                       videoDetail.ChannelId,
		ExternalChannelDisplayName:      videoDetail.ExternalChannelDisplayName,
		ChannelCustomId:                 videoDetail.ChannelCustomId,
		ExternalChannelIconUrl:          videoDetail.ExternalChannelIconUrl,
		ExternalChannelSubscribersCount: int(videoDetail.ExternalChannelSubscribersCount),
	}, nil
}
