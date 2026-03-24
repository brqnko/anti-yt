package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetAuthGoogle(ctx context.Context, request GetAuthGoogleRequestObject) (GetAuthGoogleResponseObject, error) {
	url, csrf, err := h.authService.CreateAuthCode(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetAuthGoogle500JSONResponse{
			InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	hutil.AddResponseCookie(ctx, (&http.Cookie{
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

func (h *APIHandler) GetAuthGoogleCallback(ctx context.Context, request GetAuthGoogleCallbackRequestObject) (GetAuthGoogleCallbackResponseObject, error) {
	resultAccessToken, resultRefreshToken, resultCSRFToken, resultRedirectPath, resultAccessTokenExpiresAt, resultRefreshTokenExpiresAt, err := h.authService.GoogleOIDCCallback(ctx,
		request.Params.Csrf,
		request.Params.State,
		request.Params.Code,
		request.Params.XRealIP,
		request.Params.CfIpcountry,
		"", // NOTE: googleからのリダイレクトなのでfingerpintは取得できない
		request.Params.UserAgent,
	)
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

		hutil.LogError(ctx, err)
		return GetAuthGoogleCallback500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "access_token",
		Value:    resultAccessToken,
		Path:     "/",
		Expires:  resultAccessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "refresh_token",
		Value:    resultRefreshToken,
		Path:     "/",
		Expires:  resultRefreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "csrf_token",
		Value:    resultCSRFToken,
		Path:     "/",
		Expires:  time.Now().AddDate(100, 0, 0),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())

	return GetAuthGoogleCallback302Response{
		Headers: GetAuthGoogleCallback302ResponseHeaders{
			Location: fmt.Sprintf("%s/%s", h.frontendURL, resultRedirectPath),
		},
	}, nil
}

func (h *APIHandler) PostAuthLogout(ctx context.Context, request PostAuthLogoutRequestObject) (PostAuthLogoutResponseObject, error) {
	refreshToken, ok := hutil.RefreshTokenFromContext(ctx)
	if !ok {
		return PostAuthLogout400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Detail: "refresh token not found",
				Title:  "refresh token not found",
			},
		}, nil
	}
	accessToken, ok := hutil.AccessTokenFromContext(ctx)
	if !ok {
		return PostAuthLogout500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: "access token not found",
				Title:  "access token not found",
			},
		}, nil
	}

	if err := h.authService.Logout(ctx, accessToken, refreshToken); err != nil {
		hutil.LogError(ctx, err)
		return PostAuthLogout500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	hutil.AddResponseCookie(ctx, (&http.Cookie{
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

func (h *APIHandler) PostAuthRefresh(ctx context.Context, request PostAuthRefreshRequestObject) (PostAuthRefreshResponseObject, error) {
	refreshToken, ok := hutil.RefreshTokenFromContext(ctx)
	if !ok {
		return PostAuthRefresh400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Detail: "refresh token not found",
				Title:  "refresh token not found",
			},
		}, nil
	}

	newRefreshToken, newAccessToken, accessTokenExpiresAt, refreshTokenExpiresAt, err := h.authService.RefreshToken(ctx, refreshToken, request.Params.XRealIP, request.Params.CfIpcountry, request.Params.XDeviceFingerprint, request.Params.UserAgent)
	if err != nil {
		hutil.LogError(ctx, err)
		return PostAuthRefresh500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	hutil.AddResponseCookie(ctx, (&http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		Expires:  accessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}).String())
	hutil.AddResponseCookie(ctx, (&http.Cookie{
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

func (h *APIHandler) GetUsersMeSessions(ctx context.Context, request GetUsersMeSessionsRequestObject) (GetUsersMeSessionsResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetUsersMeSessions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	sessions, hasNext, err := h.authService.GetSessions(ctx, userID, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		hutil.LogError(ctx, err)
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
			DeviceType     string             `json:"device_type"`
			Id             uuid.UUID `json:"id"`
			IpAddress      string             `json:"ip_address"`
			LastLoggedInAt time.Time          `json:"last_logged_in_at"`
			UserAgent      string             `json:"user_agent"`
		}, len(sessions)),
		HasNext: hasNext,
	}

	for i, session := range sessions {
		resp.Items[i].BrowserName = session.BrowserName
		resp.Items[i].CityName = session.CityName
		resp.Items[i].CountryCode = session.CountryCode
		resp.Items[i].CreatedAt = session.ActivatedAt
		resp.Items[i].DeviceType = session.DeviceType
		resp.Items[i].Id = session.ID
		resp.Items[i].IpAddress = session.IpAddress
		resp.Items[i].LastLoggedInAt = session.LastLoggedInAt
		resp.Items[i].UserAgent = session.UserAgent
	}

	return resp, nil
}

func (h *APIHandler) DeleteUsersMeSessionsSessionId(ctx context.Context, request DeleteUsersMeSessionsSessionIdRequestObject) (DeleteUsersMeSessionsSessionIdResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return DeleteUsersMeSessionsSessionId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	if _, err := h.authService.RemoveSession(ctx, userID, request.SessionId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DeleteUsersMeSessionsSessionId400JSONResponse{
				BadRequestJSONResponse: BadRequestJSONResponse{
					Detail: "no such refresh token",
					Title:  "no such refresh token",
				},
			}, nil
		}
		hutil.LogError(ctx, err)
		return DeleteUsersMeSessionsSessionId500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeleteUsersMeSessionsSessionId204Response{}, nil
}
