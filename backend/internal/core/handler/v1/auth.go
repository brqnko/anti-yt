package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/util"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *APIHandler) GetAuthGoogle(c context.Context, request GetAuthGoogleRequestObject) (GetAuthGoogleResponseObject, error) {
	url, csrf, err := h.authService.CreateAuthCode(c)
	if err != nil {
		util.LogError(c, err)
		return GetAuthGoogle500JSONResponse{
			InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "csrf",
		Value:    csrf,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}).String())

	return GetAuthGoogle302Response{
		Headers: GetAuthGoogle302ResponseHeaders{
			Location: url,
		},
	}, nil
}

func (h *APIHandler) GetAuthGoogleCallback(c context.Context, request GetAuthGoogleCallbackRequestObject) (GetAuthGoogleCallbackResponseObject, error) {
	result, err := h.authService.GoogleOIDCCallback(c, auth.GoogleOIDCCallbackParams{
		CSRF:        request.Params.Csrf,
		State:       request.Params.State,
		Code:        request.Params.Code,
		IPAddress:   request.Params.XRealIP,
		CountryCode: request.Params.CfIpcountry,
		UserAgent:   request.Params.UserAgent,
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCSRFOrState) {
			return GetAuthGoogleCallback400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "CSRF is missing",
					Title:  "CSRF is missing",
				},
			}, nil
		}
		if errors.Is(err, auth.ErrIDTokenNotFound) {
			return GetAuthGoogleCallback400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "Id token not found",
					Title:  "Id token not found",
				},
			}, nil
		}
		if errors.Is(err, auth.ErrInvalidCSRF) {
			return GetAuthGoogleCallback400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "CSRF is wrong",
					Title:  "CSRF is wrong",
				},
			}, nil
		}

		util.LogError(c, err)
		return GetAuthGoogleCallback500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Path:     "/",
		Expires:  result.AccessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		Expires:  result.RefreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "csrf_token",
		Value:    result.CSRFToken,
		Path:     "/",
		Expires:  time.Now().AddDate(100, 0, 0),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())

	return GetAuthGoogleCallback302Response{
		Headers: GetAuthGoogleCallback302ResponseHeaders{
			Location: fmt.Sprintf("%s/%s", h.frontendURL, result.RedirectPath),
		},
	}, nil
}

func (h *APIHandler) PostAuthLogout(c context.Context, request PostAuthLogoutRequestObject) (PostAuthLogoutResponseObject, error) {
	refreshToken, ok := util.RefreshTokenFromContext(c)
	if !ok {
		return PostAuthLogout400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Detail: "refresh token not found",
				Title:  "refresh token not found",
			},
		}, nil
	}
	accessToken, ok := util.AccessTokenFromContext(c)
	if !ok {
		return PostAuthLogout500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: "access token not found",
				Title:  "access token not found",
			},
		}, nil
	}

	if err := h.authService.Logout(c, accessToken, refreshToken); err != nil {
		util.LogError(c, err)
		return PostAuthLogout500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())

	return PostAuthLogout200Response{}, nil
}

func (h *APIHandler) PostAuthRefresh(c context.Context, request PostAuthRefreshRequestObject) (PostAuthRefreshResponseObject, error) {
	refreshToken, ok := util.RefreshTokenFromContext(c)
	if !ok {
		return PostAuthRefresh400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Detail: "refresh token not found",
				Title:  "refresh token not found",
			},
		}, nil
	}

	newRefreshToken, newAccessToken, accessTokenExpiresAt, refreshTokenExpiresAt, err := h.authService.RefreshToken(c, refreshToken, request.Params.XRealIP, request.Params.CfIpcountry, request.Params.XDeviceFingerprint, request.Params.UserAgent)
	if err != nil {
		util.LogError(c, err)
		return PostAuthRefresh500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		Expires:  accessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	util.AddResponseCookie(c, (&http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		Expires:  refreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())

	return PostAuthRefresh200Response{}, nil
}

func (h *APIHandler) GetUsersMeSessions(c context.Context, request GetUsersMeSessionsRequestObject) (GetUsersMeSessionsResponseObject, error) {
	userID, err := util.UserIDFromContext(c)
	if err != nil {
		return GetUsersMeSessions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	sessions, err := h.authService.GetSessions(c, userID)
	if err != nil {
		util.LogError(c, err)
		return GetUsersMeSessions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	resp := GetUsersMeSessions200JSONResponse{
		ItemCount: len(sessions),
		Items: make([]struct {
			BrowserName    string             `json:"browser_name"`
			CityName       string             `json:"city_name"`
			CountryCode    string             `json:"country_code"`
			CreatedAt      time.Time          `json:"created_at"`
			Id             openapi_types.UUID `json:"id"`
			LastLoggedInAt time.Time          `json:"last_logged_in_at"`
		}, len(sessions)),
	}

	for i, session := range sessions {
		resp.Items[i].BrowserName = session.BrowserName
		resp.Items[i].CityName = session.CityName
		resp.Items[i].CountryCode = session.CountryCode
		resp.Items[i].CreatedAt = session.CreatedAt
		resp.Items[i].Id = session.ID
		resp.Items[i].LastLoggedInAt = session.LastLoggedInAt
	}

	return resp, nil
}

func (h *APIHandler) DeleteUsersMeSessionsSessionId(c context.Context, request DeleteUsersMeSessionsSessionIdRequestObject) (DeleteUsersMeSessionsSessionIdResponseObject, error) {
	if _, err := h.authService.RemoveSession(c, request.SessionId); err != nil {
		if errors.Is(err, auth.ErrNoSuchRefreshToken) {
			return DeleteUsersMeSessionsSessionId400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "no such refresh token",
					Title:  "no such refresh token",
				},
			}, nil
		}
		util.LogError(c, err)
		return DeleteUsersMeSessionsSessionId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeleteUsersMeSessionsSessionId204Response{}, nil
}
