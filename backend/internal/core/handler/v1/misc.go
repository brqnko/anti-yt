package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "OK")
}
