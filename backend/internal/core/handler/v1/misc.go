package v1

import (
	"context"
)

func (h *APIHandler) GetHealth(ctx context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "OK"}, nil
}
