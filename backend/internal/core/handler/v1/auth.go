package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/core/handler"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *Handler) GetAuthGoogle(c context.Context, request GetAuthGoogleRequestObject) (GetAuthGoogleResponseObject, error) {
	url, csrf, err := h.authService.CreateAuthCode(c)
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
			SetCookie: []string{
				(&http.Cookie{
					Name:     "csrf",
					Value:    csrf,
					Path:     "/",
					Expires:  time.Now().Add(10 * time.Minute),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				}).String(),
			},
		},
	}, nil
}

func (h *Handler) GetAuthGoogleCallback(c context.Context, request GetAuthGoogleCallbackRequestObject) (GetAuthGoogleCallbackResponseObject, error) {
	result, err := h.authService.GoogleOIDCCallback(c, auth.GoogleOIDCCallbackParams{
		CSRF:              request.Params.Csrf,
		State:             request.Params.State,
		Code:              request.Params.Code,
		IPAddress:         request.Params.XRealIP,
		CountryCode:       request.Params.CfIpcountry,
		DeviceFingerprint: request.Params.XDeviceFingerprint,
		UserAgent:         request.Params.UserAgent,
	})
	if err != nil {
		if errors.Is(err, auth.ErrCSRFOrStateIsEmpty) {
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
		if errors.Is(err, auth.ErrCSRFIsWrong) {
			return GetAuthGoogleCallback400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "CSRF is wrong",
					Title:  "CSRF is wrong",
				},
			}, nil
		}

		return GetAuthGoogleCallback500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	var location string
	if result.IsCreated {
		location = fmt.Sprintf("%s/home", h.frontendURL)
	} else {
		location = fmt.Sprintf("%s/register", h.frontendURL)
	}
	return GetAuthGoogleCallback302Response{
		Headers: GetAuthGoogleCallback302ResponseHeaders{
			Location: location,
			SetCookie: []string{
				(&http.Cookie{
					Name:     "access_token",
					Value:    result.AccessToken,
					Path:     "/",
					Expires:  time.Now().Add(h.authService.AccessTokenDuration),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
				(&http.Cookie{
					Name:     "refresh_token",
					Value:    result.RefreshToken,
					Path:     "/",
					Expires:  time.Now().Add(h.authService.RefreshTokenDuration),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
				(&http.Cookie{
					Name:     "csrf_token",
					Value:    result.CSRFToken,
					Path:     "/",
					Expires:  time.Now().AddDate(100, 0, 0),
					HttpOnly: false,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
			},
		},
	}, nil
}

func (h *Handler) PostAuthLogout(c context.Context, request PostAuthLogoutRequestObject) (PostAuthLogoutResponseObject, error) {
	refreshToken, ok := c.Value(handler.RefreshTokenKey{}).(string)
	if !ok {
		return PostAuthLogout400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Detail: "refresh token not found",
				Title:  "refresh token not found",
			},
		}, nil
	}
	accessToken, ok := c.Value(handler.AccessTokenKey{}).(string)
	if !ok {
		return PostAuthLogout500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: "access token not found",
				Title:  "access token not found",
			},
		}, nil
	}

	if err := h.authService.Logout(c, accessToken, refreshToken); err != nil {
		return PostAuthLogout500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostAuthLogout200Response{
		Headers: PostAuthLogout200ResponseHeaders{
			SetCookie: []string{
				(&http.Cookie{
					Name:     "refresh_token",
					Value:    "",
					Path:     "/",
					Expires:  time.Unix(0, 0),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
				(&http.Cookie{
					Name:     "csrf_token",
					Value:    "",
					Path:     "/",
					Expires:  time.Unix(0, 0),
					HttpOnly: false,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
				(&http.Cookie{
					Name:     "access_token",
					Value:    "",
					Path:     "/",
					Expires:  time.Unix(0, 0),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
			},
		},
	}, nil
}

func (h *Handler) PostAuthRefresh(c context.Context, request PostAuthRefreshRequestObject) (PostAuthRefreshResponseObject, error) {
	refreshToken, ok := c.Value(handler.RefreshTokenKey{}).(string)
	if !ok {
		return PostAuthRefresh400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Detail: "refresh token not found",
				Title:  "refresh token not found",
			},
		}, nil
	}

	newRefreshToken, newAccessToken, err := h.authService.RefreshToken(c, refreshToken, request.Params.XRealIP, request.Params.CfIpcountry, request.Params.XDeviceFingerprint, request.Params.UserAgent)
	if err != nil {
		return PostAuthRefresh500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostAuthRefresh200Response{
		Headers: PostAuthRefresh200ResponseHeaders{
			SetCookie: []string{
				(&http.Cookie{
					Name:     "access_token",
					Value:    newAccessToken,
					Path:     "/",
					Expires:  time.Now().Add(h.authService.AccessTokenDuration),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
				(&http.Cookie{
					Name:     "refresh_token",
					Value:    newRefreshToken,
					Path:     "/",
					Expires:  time.Now().Add(h.authService.RefreshTokenDuration),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}).String(),
			},
		},
	}, nil
}

func (h *Handler) GetUsersMeSessions(c context.Context, request GetUsersMeSessionsRequestObject) (GetUsersMeSessionsResponseObject, error) {
	userID, ok := c.Value(handler.UserIDKey{}).(uuid.UUID)
	if !ok {
		return GetUsersMeSessions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	sessions, err := h.authService.GetSessions(c, userID)
	if err != nil {
		return GetUsersMeSessions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		BrowserName    string             `json:"browser_name"`
		CityName       string             `json:"city_name"`
		CountryCode    string             `json:"country_code"`
		CreatedAt      time.Time          `json:"created_at"`
		Id             openapi_types.UUID `json:"id"`
		LastLoggedInAt time.Time          `json:"last_logged_in_at"`
	}, len(sessions))

	for i, session := range sessions {
		items[i] = struct {
			BrowserName    string             `json:"browser_name"`
			CityName       string             `json:"city_name"`
			CountryCode    string             `json:"country_code"`
			CreatedAt      time.Time          `json:"created_at"`
			Id             openapi_types.UUID `json:"id"`
			LastLoggedInAt time.Time          `json:"last_logged_in_at"`
		}{
			BrowserName:    session.BrowserName,
			CityName:       session.CityName,
			CountryCode:    session.CountryCode,
			CreatedAt:      session.CreatedAt,
			Id:             session.ID,
			LastLoggedInAt: session.LastLoggedInAt,
		}
	}

	return GetUsersMeSessions200JSONResponse{
		ItemCount: len(sessions),
		Items:     items,
	}, nil
}

func (h *Handler) DeleteUsersMeSessionsSessionId(c context.Context, request DeleteUsersMeSessionsSessionIdRequestObject) (DeleteUsersMeSessionsSessionIdResponseObject, error) {
	if _, err := h.authService.RemoveSession(c, request.SessionId); err != nil {
		if errors.Is(err, auth.ErrNoSuchRefreshToken) {
			return DeleteUsersMeSessionsSessionId400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "no such refresh token",
					Title:  "no such refresh token",
				},
			}, nil
		}
		return DeleteUsersMeSessionsSessionId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeleteUsersMeSessionsSessionId204Response{}, nil
}
