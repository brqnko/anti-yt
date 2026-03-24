package v1

import (
	"context"
)

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
