package middleware_d_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyHandler は StrictHandlerFunc をキャプチャするヘルパー。
// middlewareが f を呼び出したか、呼び出し時の ctx に何が入っていたかを記録する。
type spyHandler struct {
	called bool
	ctx    context.Context
}

func (s *spyHandler) fn() v1.StrictHandlerFunc {
	return func(ctx context.Context, _ http.ResponseWriter, _ *http.Request, _ interface{}) (interface{}, error) {
		s.called = true
		s.ctx = ctx
		return "ok", nil
	}
}

func newRequestWithAccessTokenCookie(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/some", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	}
	return req
}

func decodeErrorTitle(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	return body.Title
}

func TestAccessTokenMiddleware(t *testing.T) {
	ctx := context.Background()

	t.Run("operationID is in ignore list: skip validation and pass through", func(t *testing.T) {
		// arrange
		jtiRepo := &JtiBlacklistRepositoryMock{
			IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
				t.Fatal("IsJtiExist should not be called for ignored operations")
				return false, nil
			},
		}
		jwtMock := &ServiceMock{
			VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
				t.Fatal("VerifyUserAccessToken should not be called for ignored operations")
				return uuid.Nil, uuid.Nil, time.Time{}, nil
			},
		}
		spy := &spyHandler{}
		mw := middleware_d.AccessTokenMiddleware(jwtMock, jtiRepo, map[string]struct{}{
			"GetHealth": {},
		})
		wrapped := mw(spy.fn(), "GetHealth")

		// action
		w := httptest.NewRecorder()
		resp, err := wrapped(ctx, w, newRequestWithAccessTokenCookie(""), nil)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.True(t, spy.called)
		_, uerr := hutil.UserIDFromContext(spy.ctx)
		assert.ErrorIs(t, uerr, hutil.ErrUserIDNotFoundInContext,
			"ignore list pass-through must not inject user_id into ctx")
	})

	t.Run("success: jti not in blacklist", func(t *testing.T) {
		// arrange
		userID := uuid.Must(uuid.NewV7())
		jti := uuid.Must(uuid.NewV7())
		jtiRepo := &JtiBlacklistRepositoryMock{
			IsJtiExistFunc: func(_ context.Context, gotJti uuid.UUID) (bool, error) {
				assert.Equal(t, jti, gotJti)
				return false, nil
			},
		}
		jwtMock := &ServiceMock{
			VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
				return userID, jti, time.Now().Add(time.Hour), nil
			},
		}
		spy := &spyHandler{}
		mw := middleware_d.AccessTokenMiddleware(jwtMock, jtiRepo, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")

		// action
		w := httptest.NewRecorder()
		resp, err := wrapped(ctx, w, newRequestWithAccessTokenCookie("valid-token"), nil)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.True(t, spy.called)
		gotUserID, uerr := hutil.UserIDFromContext(spy.ctx)
		require.NoError(t, uerr)
		assert.Equal(t, userID, gotUserID)
	})

	t.Run("unauthorized: access_token cookie is missing", func(t *testing.T) {
		// arrange
		jtiRepo := &JtiBlacklistRepositoryMock{}
		jwtMock := &ServiceMock{
			VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
				t.Fatal("VerifyUserAccessToken should not be called when cookie is missing")
				return uuid.Nil, uuid.Nil, time.Time{}, nil
			},
		}
		spy := &spyHandler{}
		mw := middleware_d.AccessTokenMiddleware(jwtMock, jtiRepo, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, newRequestWithAccessTokenCookie(""), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "unauthorized", decodeErrorTitle(t, w))
	})

	t.Run("unauthorized: VerifyUserAccessToken fails", func(t *testing.T) {
		// arrange
		jtiRepo := &JtiBlacklistRepositoryMock{}
		jwtMock := &ServiceMock{
			VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
				return uuid.Nil, uuid.Nil, time.Time{}, errors.New("invalid signature")
			},
		}
		spy := &spyHandler{}
		mw := middleware_d.AccessTokenMiddleware(jwtMock, jtiRepo, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, newRequestWithAccessTokenCookie("broken-token"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "unauthorized", decodeErrorTitle(t, w))
	})

	t.Run("unauthorized: jti is blacklisted", func(t *testing.T) {
		// arrange
		userID := uuid.Must(uuid.NewV7())
		jti := uuid.Must(uuid.NewV7())
		jtiRepo := &JtiBlacklistRepositoryMock{
			IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return true, nil
			},
		}
		jwtMock := &ServiceMock{
			VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
				return userID, jti, time.Now().Add(time.Hour), nil
			},
		}
		spy := &spyHandler{}
		mw := middleware_d.AccessTokenMiddleware(jwtMock, jtiRepo, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, newRequestWithAccessTokenCookie("valid-token"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "unauthorized", decodeErrorTitle(t, w))
	})

	t.Run("internal server error: jti blacklist lookup fails", func(t *testing.T) {
		// arrange
		userID := uuid.Must(uuid.NewV7())
		jti := uuid.Must(uuid.NewV7())
		jtiRepo := &JtiBlacklistRepositoryMock{
			IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return false, errors.New("redis down")
			},
		}
		jwtMock := &ServiceMock{
			VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
				return userID, jti, time.Now().Add(time.Hour), nil
			},
		}
		spy := &spyHandler{}
		mw := middleware_d.AccessTokenMiddleware(jwtMock, jtiRepo, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, newRequestWithAccessTokenCookie("valid-token"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal_server_error", decodeErrorTitle(t, w))
	})
}
