package v1

import (
	"context"
)

func (h *Handler) GetHealth(c context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "OK"}, nil
}
