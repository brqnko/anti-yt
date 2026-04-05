package v1

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func (h *APIHandler) GetVideosVideoId(ctx context.Context, request GetVideosVideoIdRequestObject) (GetVideosVideoIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	videoDetail, err := h.videoService.GetVideoDetail(ctx, userID, request.VideoId.UUID())
	if err != nil {
		return nil, err
	}

	return GetVideosVideoId200JSONResponse{
		VideoId:                         util.Base64UUID(videoDetail.VideoId),
		ExternalVideoId:                 videoDetail.ExternalVideoId,
		ExternalVideoTitle:              videoDetail.ExternalVideoTitle,
		ExternalVideoDescription:        videoDetail.ExternalVideoDescription,
		ExternalVideoThumbnailUrl:       videoDetail.ExternalVideoThumbnailUrl,
		ExternalVideoCreatedAt:          videoDetail.ExternalVideoCreatedAt,
		ChannelId:                       util.Base64UUID(videoDetail.ChannelId),
		ExternalChannelDisplayName:      videoDetail.ExternalChannelDisplayName,
		ChannelCustomId:                 videoDetail.ChannelCustomId,
		ExternalChannelIconUrl:          videoDetail.ExternalChannelIconUrl,
		ExternalChannelSubscribersCount: int(videoDetail.ExternalChannelSubscribersCount),
		IsWatched:                       videoDetail.IsWatched,
		IsInWatchLater:                  videoDetail.IsInWatchLater,
	}, nil
}

func (h *APIHandler) PostVideosVideoIdWatchLater(ctx context.Context, request PostVideosVideoIdWatchLaterRequestObject) (PostVideosVideoIdWatchLaterResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.playlistService.MarkAsWatchLater(ctx, userID, request.VideoId.UUID()); err != nil {
		return nil, err
	}

	return PostVideosVideoIdWatchLater201Response{}, nil
}

func (h *APIHandler) DeleteVideosVideoIdWatchLater(ctx context.Context, request DeleteVideosVideoIdWatchLaterRequestObject) (DeleteVideosVideoIdWatchLaterResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.playlistService.UnmarkAsWatchLater(ctx, userID, request.VideoId.UUID()); err != nil {
		return nil, err
	}

	return DeleteVideosVideoIdWatchLater204Response{}, nil
}
