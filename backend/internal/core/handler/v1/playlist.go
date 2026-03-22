package v1

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/util"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *APIHandler) GetPlaylists(c context.Context, request GetPlaylistsRequestObject) (GetPlaylistsResponseObject, error) {
	playlists, err := h.playlistService.GetPlaylists(c, request.Params.Cursor, request.Params.Limit+1)
	if err != nil {
		util.LogError(c, err)
		return GetPlaylists500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	hasNext := len(playlists) > request.Params.Limit
	if hasNext {
		playlists = playlists[:request.Params.Limit]
	}

	items := make([]struct {
		PlaylistCreatedAt   time.Time          `json:"playlist_created_at"`
		PlaylistDescription string             `json:"playlist_description"`
		PlaylistId          openapi_types.UUID `json:"playlist_id"`
		PlaylistTitle       string             `json:"playlist_title"`
		PlaylistType        PlaylistType       `json:"playlist_type"`
		PlaylistUpdatedAt   time.Time          `json:"playlist_updated_at"`
		PlaylistVideoCount  int                `json:"playlist_video_count"`
		PlaylistVisibility  PlaylistVisibility `json:"playlist_visibility"`
		TopVideoThumbnailUrl *string           `json:"top_video_thumbnail_url,omitempty"`
	}, len(playlists))

	for i, pl := range playlists {
		items[i].PlaylistId = pl.ID
		items[i].PlaylistTitle = string(pl.Title)
		items[i].PlaylistDescription = string(pl.Description)
		items[i].PlaylistType = PlaylistType(pl.PlaylistCode.String())
		items[i].PlaylistVisibility = PlaylistVisibility(pl.VisibilityCode.String())
		items[i].PlaylistVideoCount = pl.VideoCount
		items[i].PlaylistCreatedAt = pl.CreatedAt
		items[i].PlaylistUpdatedAt = pl.CreatedAt
		if pl.TopVideoThumbnailURL != "" {
			items[i].TopVideoThumbnailUrl = &pl.TopVideoThumbnailURL
		}
	}

	return GetPlaylists200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) PostPlaylists(c context.Context, request PostPlaylistsRequestObject) (PostPlaylistsResponseObject, error) {
	created, err := h.playlistService.CreatePlaylist(
		c,
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

		util.LogError(c, err)
		return PostPlaylists500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostPlaylists201JSONResponse{
		PlaylistId:          created.ID,
		PlaylistType:        PlaylistType(created.PlaylistCode.String()),
		PlaylistVisibility:  PlaylistVisibility(created.VisibilityCode.String()),
		PlaylistTitle:       string(created.Title),
		PlaylistDescription: string(created.Description),
		PlaylistVideoCount:  created.VideoCount,
		PlaylistCreatedAt:   created.CreatedAt,
		PlaylistUpdatedAt:   created.CreatedAt,
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistId(c context.Context, request DeletePlaylistsPlaylistIdRequestObject) (DeletePlaylistsPlaylistIdResponseObject, error) {
	err := h.playlistService.DeletePlaylist(c, request.PlaylistId)
	if err != nil {
		if errors.Is(err, playlist.ErrPlaylistNotFound) {
			return DeletePlaylistsPlaylistId404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		util.LogError(c, err)
		return DeletePlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeletePlaylistsPlaylistId204Response{}, nil
}

func (h *APIHandler) GetPlaylistsPlaylistId(c context.Context, request GetPlaylistsPlaylistIdRequestObject) (GetPlaylistsPlaylistIdResponseObject, error) {
	pl, err := h.playlistService.GetPlaylistInfo(c, request.PlaylistId)
	if err != nil {
		if errors.Is(err, playlist.ErrPlaylistNotFound) {
			return GetPlaylistsPlaylistId404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		util.LogError(c, err)
		return GetPlaylistsPlaylistId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return GetPlaylistsPlaylistId200JSONResponse{
		PlaylistId:           pl.ID,
		PlaylistTitle:        string(pl.Title),
		PlaylistDescription:  string(pl.Description),
		PlaylistType:         PlaylistType(pl.PlaylistCode.String()),
		PlaylistVisibility:   PlaylistVisibility(pl.VisibilityCode.String()),
		PlaylistVideoCount:   pl.VideoCount,
		PlaylistCreatedAt:    pl.CreatedAt,
		PlaylistUpdatedAt:    pl.CreatedAt,
		TopVideoThumbnailUrl: func() *string {
			if pl.TopVideoThumbnailURL == "" {
				return nil
			}
			return &pl.TopVideoThumbnailURL
		}(),
	}, nil
}

func (h *APIHandler) GetPlaylistsPlaylistIdVideos(c context.Context, request GetPlaylistsPlaylistIdVideosRequestObject) (GetPlaylistsPlaylistIdVideosResponseObject, error) {
	videos, err := h.playlistService.GetPlaylistItems(c, request.PlaylistId, request.Params.Cursor, request.Params.Limit+1)
	if err != nil {
		util.LogError(c, err)
		return GetPlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	hasNext := len(videos) > request.Params.Limit
	if hasNext {
		videos = videos[:request.Params.Limit]
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
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailURL
		items[i].ExternalVideoTitle = v.ExternalVideoTitle
		items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt
		items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		items[i].ExternalChannelIconUrl = v.ExternalChannelIconURL
		items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayname
		if v.LastWatchSeconds != 0 {
			items[i].LastWatchSeconds = &v.LastWatchSeconds
		}
	}

	return GetPlaylistsPlaylistIdVideos200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) PatchPlaylistsPlaylistId(c context.Context, request PatchPlaylistsPlaylistIdRequestObject) (PatchPlaylistsPlaylistIdResponseObject, error) {
	updated, err := h.playlistService.UpdatePlaylist(c, request.PlaylistId, request.Body.PlaylistTitle, request.Body.PlaylistDescription)
	if err != nil {
		if errors.Is(err, playlist.ErrPlaylistNotFound) {
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

		util.LogError(c, err)
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
		PlaylistUpdatedAt:   updated.CreatedAt,
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistIdVideos(c context.Context, request DeletePlaylistsPlaylistIdVideosRequestObject) (DeletePlaylistsPlaylistIdVideosResponseObject, error) {
	err := h.playlistService.RemoveVideoFromPlaylist(c, request.PlaylistId, request.Params.VideoId)
	if err != nil {
		if errors.Is(err, playlist.ErrPlaylistNotFound) {
			return DeletePlaylistsPlaylistIdVideos404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}
		if errors.Is(err, playlist.ErrVideoNotInPlaylist) {
			return DeletePlaylistsPlaylistIdVideos404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		util.LogError(c, err)
		return DeletePlaylistsPlaylistIdVideos500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeletePlaylistsPlaylistIdVideos204Response{}, nil
}

func (h *APIHandler) PostPlaylistsPlaylistIdVideos(c context.Context, request PostPlaylistsPlaylistIdVideosRequestObject) (PostPlaylistsPlaylistIdVideosResponseObject, error) {
	insertedAt, err := h.playlistService.InsertVideoIntoPlaylist(c, request.PlaylistId, request.Body.VideoId)
	if err != nil {
		if errors.Is(err, playlist.ErrPlaylistNotFound) {
			return PostPlaylistsPlaylistIdVideos404JSONResponse{NotFoundJSONResponse{
				Detail: err.Error(),
				Title:  "Not Found",
			}}, nil
		}

		util.LogError(c, err)
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
		InsertedAt: insertedAt,
	}, nil
}
