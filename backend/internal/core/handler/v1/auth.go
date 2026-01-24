package v1

import (
	"context"
	"net/http"
	"time"
)

func (h *Handler) GetAuthGoogle(c context.Context, request GetAuthGoogleRequestObject) (GetAuthGoogleResponseObject, error) {
	ctx, cancel := newContext()
	defer cancel()
	url, state, err := h.authService.CreateAuthCode(ctx)
	if err != nil {
		return GetAuthGoogle500JSONResponse{
			InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return GetAuthGoogle302Response{
		Headers: GetAuthGoogle302ResponseHeaders{
			Location: url,
			SetCookie: (&http.Cookie{
				Name:     "state",
				Value:    state,
				Path:     "/",
				Expires:  time.Now().Add(10 * time.Minute),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}).String(),
		},
	}, nil
}

func (h *Handler) GetAuthGoogleCallback(c context.Context, request GetAuthGoogleCallbackRequestObject) (GetAuthGoogleCallbackResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthLogout(c context.Context, request PostAuthLogoutRequestObject) (PostAuthLogoutResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthRefresh(c context.Context, request PostAuthRefreshRequestObject) (PostAuthRefreshResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetUsersMeSessions(c context.Context, request GetUsersMeSessionsRequestObject) (GetUsersMeSessionsResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeleteUsersMeSessionsSessionId(c context.Context, request DeleteUsersMeSessionsSessionIdRequestObject) (DeleteUsersMeSessionsSessionIdResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
