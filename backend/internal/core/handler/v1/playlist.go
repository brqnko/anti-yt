package v1

import (
	"github.com/labstack/echo/v4"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

func (h *Handler) GetPlaylists(ctx echo.Context, params GetPlaylistsParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostPlaylists(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeletePlaylistsPlaylistId(ctx echo.Context, playlistId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetPlaylistsPlaylistId(ctx echo.Context, playlistId openapitypes.UUID, params GetPlaylistsPlaylistIdParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PatchPlaylistsPlaylistId(ctx echo.Context, playlistId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeletePlaylistsPlaylistIdVideos(ctx echo.Context, playlistId openapitypes.UUID, params DeletePlaylistsPlaylistIdVideosParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostPlaylistsPlaylistIdVideos(ctx echo.Context, playlistId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}
