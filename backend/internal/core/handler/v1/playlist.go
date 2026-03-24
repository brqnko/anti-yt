package v1

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetPlaylists(ctx context.Context, request GetPlaylistsRequestObject) (GetPlaylistsResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetPlaylists500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	playlists, hasNext, err := h.playlistService.GetPlaylists(ctx, userID, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		hutil.LogError(ctx, err)
		return GetPlaylists500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	resp := GetPlaylists200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(playlists),
		Items: make([]struct {
			PlaylistDescription  string             `json:"playlist_description"`
			PlaylistId           uuid.UUID `json:"playlist_id"`
			PlaylistRegisteredAt time.Time          `json:"playlist_registered_at"`
			PlaylistTitle        string             `json:"playlist_title"`
			PlaylistType         PlaylistType       `json:"playlist_type"`
			PlaylistUpdatedAt    time.Time          `json:"playlist_updated_at"`
			PlaylistVideoCount   int                `json:"playlist_video_count"`
			PlaylistVisibility   PlaylistVisibility `json:"playlist_visibility"`
			TopVideoThumbnailUrl *string            `json:"top_video_thumbnail_url,omitempty"`
		}, len(playlists)),
	}

	for i, pl := range playlists {
		resp.Items[i].PlaylistId = pl.PlaylistId
		resp.Items[i].PlaylistTitle = pl.PlaylistTitle
		resp.Items[i].PlaylistDescription = pl.PlaylistDescription
		resp.Items[i].PlaylistType = PlaylistType(pl.PlaylistType)
		resp.Items[i].PlaylistVisibility = PlaylistVisibility(pl.PlaylistVisibility)
		resp.Items[i].PlaylistVideoCount = pl.PlaylistVideoCount
		resp.Items[i].PlaylistRegisteredAt = pl.PlaylistRegisteredAt
		resp.Items[i].PlaylistUpdatedAt = pl.PlaylistUpdatedAt
		resp.Items[i].TopVideoThumbnailUrl = pl.TopVideoThumbnailUrl
	}

	return resp, nil
}

func (h *APIHandler) PostPlaylists(ctx context.Context, request PostPlaylistsRequestObject) (PostPlaylistsResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return PostPlaylists500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	created, err := h.playlistService.CreatePlaylist(
		ctx,
		userID,
		request.Body.PlaylistTitle,
		request.Body.PlaylistDescription,
		string(request.Body.PlaylistVisibility),
		string(request.Body.PlaylistType),
		request.Body.BasePlaylistUrl,
	)
	if err != nil {
		if errors.Is(err, playlist.ErrInvalidPlaylistTitle) ||
			errors.Is(err, playlist.ErrInvalidPlaylistDescription) ||
			errors.Is(err, playlist.ErrInvalidVisibilityCode) ||
			errors.Is(err, playlist.ErrInvalidPlaylistCode) {
			return PostPlaylists400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return PostPlaylists500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostPlaylists201JSONResponse{
		PlaylistId:           created.ID,
		PlaylistType:         PlaylistType(created.PlaylistCode.String()),
		PlaylistVisibility:   PlaylistVisibility(created.VisibilityCode.String()),
		PlaylistTitle:        string(created.Title),
		PlaylistDescription:  string(created.Description),
		PlaylistVideoCount:   created.VideoCount,
		PlaylistRegisteredAt: created.RegisteredAt,
		// TODO: updated_atは物理カラムとして存在するがドメインには含めていない。
		// QueryServiceではSQLから直接updated_atを取得して誤魔化しているが、ここではドメイン経由のためRegisteredAtで代用している。
		PlaylistUpdatedAt: created.RegisteredAt,
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistId(ctx context.Context, request DeletePlaylistsPlaylistIdRequestObject) (DeletePlaylistsPlaylistIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return DeletePlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	err = h.playlistService.DeletePlaylist(ctx, userID, request.PlaylistId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DeletePlaylistsPlaylistId404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return DeletePlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeletePlaylistsPlaylistId204Response{}, nil
}

func (h *APIHandler) GetPlaylistsPlaylistId(ctx context.Context, request GetPlaylistsPlaylistIdRequestObject) (GetPlaylistsPlaylistIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetPlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	pl, err := h.playlistService.GetPlaylistDetail(ctx, userID, request.PlaylistId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetPlaylistsPlaylistId404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return GetPlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return GetPlaylistsPlaylistId200JSONResponse{
		PlaylistId:           pl.PlaylistId,
		PlaylistTitle:        pl.PlaylistTitle,
		PlaylistDescription:  pl.PlaylistDescription,
		PlaylistType:         PlaylistType(pl.PlaylistType),
		PlaylistVisibility:   PlaylistVisibility(pl.PlaylistVisibility),
		PlaylistVideoCount:   pl.PlaylistVideoCount,
		PlaylistRegisteredAt: pl.PlaylistRegisteredAt,
		PlaylistUpdatedAt:    pl.PlaylistUpdatedAt,
		TopVideoThumbnailUrl: pl.TopVideoThumbnailUrl,
	}, nil
}

func (h *APIHandler) GetPlaylistsPlaylistIdVideos(ctx context.Context, request GetPlaylistsPlaylistIdVideosRequestObject) (GetPlaylistsPlaylistIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetPlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	videos, hasNext, err := h.playlistService.GetPlaylistItems(ctx, userID, request.PlaylistId, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		hutil.LogError(ctx, err)
		return GetPlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	resp := GetPlaylistsPlaylistIdVideos200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(videos),
		Items: make([]struct {
			ChannelId                  uuid.UUID `json:"channel_id"`
			ExternalChannelDisplayName string             `json:"external_channel_display_name"`
			ExternalChannelIconUrl     string             `json:"external_channel_icon_url"`
			ExternalVideoCreatedAt     time.Time          `json:"external_video_created_at"`
			ExternalVideoLengthSeconds int                `json:"external_video_length_seconds"`
			ExternalVideoThumbnailUrl  string             `json:"external_video_thumbnail_url"`
			ExternalVideoTitle         string             `json:"external_video_title"`
			LastWatchSeconds           *int               `json:"last_watch_seconds,omitempty"`
			VideoId                    uuid.UUID `json:"video_id"`
		}, len(videos)),
	}

	for i, v := range videos {
		resp.Items[i].VideoId = v.VideoId
		resp.Items[i].ChannelId = v.ChannelId
		resp.Items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		resp.Items[i].ExternalVideoTitle = v.ExternalVideoTitle
		resp.Items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt
		resp.Items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		resp.Items[i].ExternalChannelIconUrl = v.ExternalChannelIconUrl
		resp.Items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayName
		resp.Items[i].LastWatchSeconds = v.LastWatchSeconds
	}

	return resp, nil
}

func (h *APIHandler) PatchPlaylistsPlaylistId(ctx context.Context, request PatchPlaylistsPlaylistIdRequestObject) (PatchPlaylistsPlaylistIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return PatchPlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	updated, err := h.playlistService.UpdatePlaylist(ctx, userID, request.PlaylistId, request.Body.PlaylistTitle, request.Body.PlaylistDescription)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PatchPlaylistsPlaylistId404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}
		if errors.Is(err, playlist.ErrInvalidPlaylistTitle) ||
			errors.Is(err, playlist.ErrInvalidPlaylistDescription) {
			return PatchPlaylistsPlaylistId400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return PatchPlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PatchPlaylistsPlaylistId200JSONResponse{
		PlaylistId:          updated.ID,
		PlaylistTitle:       string(updated.Title),
		PlaylistDescription: string(updated.Description),
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistIdVideos(ctx context.Context, request DeletePlaylistsPlaylistIdVideosRequestObject) (DeletePlaylistsPlaylistIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return DeletePlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	err = h.playlistService.RemoveVideoFromPlaylist(ctx, userID, request.PlaylistId, request.Params.VideoId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DeletePlaylistsPlaylistIdVideos404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return DeletePlaylistsPlaylistIdVideos404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return DeletePlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeletePlaylistsPlaylistIdVideos204Response{}, nil
}

func (h *APIHandler) PostPlaylistsPlaylistIdVideos(ctx context.Context, request PostPlaylistsPlaylistIdVideosRequestObject) (PostPlaylistsPlaylistIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return PostPlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	err = h.playlistService.InsertVideoIntoPlaylist(ctx, userID, request.PlaylistId, request.Body.VideoId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostPlaylistsPlaylistIdVideos404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return PostPlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostPlaylistsPlaylistIdVideos201JSONResponse{
		PlaylistId: request.PlaylistId,
		VideoId:    request.Body.VideoId,
		InsertedAt: time.Now().UTC(),
	}, nil
}
