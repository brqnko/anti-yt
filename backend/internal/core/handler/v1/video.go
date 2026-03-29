package v1

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetVideosVideoId(ctx context.Context, request GetVideosVideoIdRequestObject) (GetVideosVideoIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	videoDetail, err := h.videoService.GetVideoDetail(ctx, userID, request.VideoId)
	if err != nil {
		return nil, err
	}

	return GetVideosVideoId200JSONResponse{
		VideoId:                         videoDetail.VideoId,
		ExternalVideoId:                 videoDetail.ExternalVideoId,
		ExternalVideoTitle:              videoDetail.ExternalVideoTitle,
		ExternalVideoDescription:        videoDetail.ExternalVideoDescription,
		ExternalVideoThumbnailUrl:       videoDetail.ExternalVideoThumbnailUrl,
		ExternalVideoCreatedAt:          videoDetail.ExternalVideoCreatedAt,
		ChannelId:                       videoDetail.ChannelId,
		ExternalChannelDisplayName:      videoDetail.ExternalChannelDisplayName,
		ChannelCustomId:                 videoDetail.ChannelCustomId,
		ExternalChannelIconUrl:          videoDetail.ExternalChannelIconUrl,
		ExternalChannelSubscribersCount: int(videoDetail.ExternalChannelSubscribersCount),
		IsWatched:                       videoDetail.IsWatched,
	}, nil
}
