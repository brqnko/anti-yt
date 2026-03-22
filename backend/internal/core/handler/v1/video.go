package v1

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *APIHandler) GetFeed(c context.Context, request GetFeedRequestObject) (GetFeedResponseObject, error) {
	videos, hasNext, err := h.channelService.GetFeed(c, request.Params.Cursor, request.Params.Limit)
	if err != nil {
		util.LogError(c, err)
		return GetFeed500JSONResponse{
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
		ExternalVideoCreatedAt     time.Time          `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int                `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string             `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string             `json:"external_video_title"`
		LastWatchSeconds           *int               `json:"last_watch_seconds,omitempty"`
		VideoId                    openapi_types.UUID `json:"video_id"`
	}, len(videos))

	for i, v := range videos {
		items[i].VideoId = v.ID
		items[i].ChannelId = v.ChannelID
		items[i].ExternalChannelIconUrl = v.ExternalChannelIconURL
		items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayname
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailURL
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

func (h *APIHandler) GetSearch(c context.Context, request GetSearchRequestObject) (GetSearchResponseObject, error) {
	// TODO
	return GetSearch500JSONResponse{
		InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
			Detail: internalErrorDetail,
			Title:  internalErrorTitle,
		},
	}, nil
}

func (h *APIHandler) GetVideosVideoId(c context.Context, request GetVideosVideoIdRequestObject) (GetVideosVideoIdResponseObject, error) {
	videoDetail, err := h.videoService.GetVideoDetail(c, request.VideoId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetVideosVideoId404JSONResponse{
				NotFoundJSONResponse: NotFoundJSONResponse{
					Detail: "video not found",
					Title:  "Not Found",
				},
			}, nil
		}
		util.LogError(c, err)
		return GetVideosVideoId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return GetVideosVideoId200JSONResponse{
		VideoId:                         openapi_types.UUID(videoDetail.ID),
		ExternalVideoId:                 string(*videoDetail.ExternalVideoID),
		ExternalVideoTitle:              videoDetail.ExternalVideoTitle,
		ExternalVideoDescription:        videoDetail.ExternalVideoDescription,
		ExternalVideoThumbnailUrl:       videoDetail.ExternalVideoThumbnailURL,
		ChannelId:                       openapi_types.UUID(videoDetail.ChannelID),
		ExternalChannelId:               videoDetail.ExternalChannelID,
		ExternalChannelDisplayName:      videoDetail.ExternalChannelDisplayName,
		ChannelCustomId:                 videoDetail.ChannelCustomID,
		ExternalChannelIconUrl:          videoDetail.ExternalChannelIconURL,
		ExternalChannelSubscribersCount: videoDetail.ExternalChannelSubscribersCount,
	}, nil
}

func (h *APIHandler) PostVideosVideoIdHeartbeats(c context.Context, request PostVideosVideoIdHeartbeatsRequestObject) (PostVideosVideoIdHeartbeatsResponseObject, error) {
	remaining, err := h.videoService.Heartbeat(c, request.VideoId, request.Body.CurrentPositionSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostVideosVideoIdHeartbeats404JSONResponse{
				NotFoundJSONResponse: NotFoundJSONResponse{
					Detail: "video not found",
					Title:  "Not Found",
				},
			}, nil
		}
		util.LogError(c, err)
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
