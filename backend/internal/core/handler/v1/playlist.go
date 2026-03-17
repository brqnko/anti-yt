package v1

import (
	"context"
	"errors"

	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func (h *APIHandler) GetPlaylists(c context.Context, request GetPlaylistsRequestObject) (GetPlaylistsResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *APIHandler) PostPlaylists(c context.Context, request PostPlaylistsRequestObject) (PostPlaylistsResponseObject, error) {
	created, err := h.playlistService.CreatePlaylist(
		c,
		request.Body.PlaylistTitle,
		request.Body.PlaylistDescription,
		string(request.Body.PlaylistVisibility),
		string(request.Body.PlaylistType),
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
		PlaylistTitle:       string(*created.Title),
		PlaylistDescription: string(*created.Description),
		PlaylistVideoCount:  created.VideoCount,
		PlaylistCreatedAt:   created.CreatedAt,
		PlaylistUpdatedAt:   created.UpdatedAt,
	}, nil
}

func (h *APIHandler) DeletePlaylistsPlaylistId(c context.Context, request DeletePlaylistsPlaylistIdRequestObject) (DeletePlaylistsPlaylistIdResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *APIHandler) GetPlaylistsPlaylistId(c context.Context, request GetPlaylistsPlaylistIdRequestObject) (GetPlaylistsPlaylistIdResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *APIHandler) PatchPlaylistsPlaylistId(c context.Context, request PatchPlaylistsPlaylistIdRequestObject) (PatchPlaylistsPlaylistIdResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *APIHandler) DeletePlaylistsPlaylistIdVideos(c context.Context, request DeletePlaylistsPlaylistIdVideosRequestObject) (DeletePlaylistsPlaylistIdVideosResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *APIHandler) PostPlaylistsPlaylistIdVideos(c context.Context, request PostPlaylistsPlaylistIdVideosRequestObject) (PostPlaylistsPlaylistIdVideosResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
