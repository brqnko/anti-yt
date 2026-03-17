package v1

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/util"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const defaultGetChannelVideosLimit = 20

func (h *APIHandler) GetChannelsChannelIdVideos(c context.Context, request GetChannelsChannelIdVideosRequestObject) (GetChannelsChannelIdVideosResponseObject, error) {
	videos, hasNext, err := h.channelService.GetChannelUploads(c, request.ChannelId, request.Params.Cursor, defaultGetChannelVideosLimit)
	if err != nil {
		if errors.Is(err, channel.ErrInvalidGetUploadLimit) {
			return GetChannelsChannelIdVideos400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		util.LogError(c, err)
		return GetChannelsChannelIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		ExternalVideoCreatedAt     time.Time          `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int                `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string             `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string             `json:"external_video_title"`
		LastWatchSeconds           *int               `json:"last_watch_seconds,omitempty"`
		VideoId                    openapi_types.UUID `json:"video_id"`
	}, len(videos))

	for i, v := range videos {
		items[i].VideoId = v.ID
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailURL
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

func (h *APIHandler) GetFeedChannels(c context.Context, request GetFeedChannelsRequestObject) (GetFeedChannelsResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
