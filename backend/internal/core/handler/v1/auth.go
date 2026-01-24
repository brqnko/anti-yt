package v1

import (
	"github.com/labstack/echo/v4"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

func (h *Handler) GetAuthGoogle(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthLogout(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthGoogle(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetAuthGoogleCallback(ctx echo.Context, params GetAuthGoogleCallbackParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthRefresh(ctx echo.Context, params PostAuthRefreshParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetUsersMeSessions(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeleteUsersMeSessionsSessionId(ctx echo.Context, sessionId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}
