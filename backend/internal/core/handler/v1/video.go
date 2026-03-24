package v1

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (h *APIHandler) GetSearch(ctx context.Context, request GetSearchRequestObject) (GetSearchResponseObject, error) {
	items, hasNext, _, err := h.feedService.Search(ctx, request.Params.Query, request.Params.Limit, request.Params.Cursor)
	if err != nil {
		return nil, err
	}

	resp := GetSearch200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items: make([]struct {
			ChannelId                  uuid.UUID `json:"channel_id"`
			ExternalChannelDisplayName string    `json:"external_channel_display_name"`
			ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
			ExternalChannelId          string    `json:"external_channel_id"`
			ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
			ExternalVideoDescription   string    `json:"external_video_description"`
			ExternalVideoId            string    `json:"external_video_id"`
			ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
			ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
			ExternalVideoTitle         string    `json:"external_video_title"`
			LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
			VideoId                    uuid.UUID `json:"video_id"`
		}, len(items)),
	}

	for i, item := range items {
		resp.Items[i].VideoId = item.VideoID
		resp.Items[i].ChannelId = item.ChannelID
		resp.Items[i].ExternalVideoId = item.ExternalVideoID
		resp.Items[i].ExternalChannelId = item.ExternalChannelID
		resp.Items[i].ExternalVideoTitle = item.ExternalVideoTitle
		resp.Items[i].ExternalVideoDescription = item.ExternalVideoDescription
		resp.Items[i].ExternalVideoThumbnailUrl = item.ExternalVideoThumbnailUrl
		resp.Items[i].ExternalChannelDisplayName = item.ExternalChannelDisplayName
		resp.Items[i].ExternalChannelIconUrl = item.ExternalChannelIconUrl
		resp.Items[i].ExternalVideoCreatedAt = item.ExternalVideoCreatedAt
		resp.Items[i].ExternalVideoLengthSeconds = item.ExternalVideoLengthSeconds
	}

	return resp, nil
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
