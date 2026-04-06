package middleware_d_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// csrfRequest はテスト用にメソッド、cookie値、ヘッダ値を指定してリクエストを組み立てる。
// cookieValue が "-" の場合は cookie を一切セットしない、空文字の場合は
// 空の cookie を明示的にセットする。header も同様。
func csrfRequest(method, cookieValue, headerValue string) *http.Request {
	req := httptest.NewRequest(method, "/api/v1/something", nil)
	if cookieValue != "-" {
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: cookieValue})
	}
	if headerValue != "-" {
		req.Header.Set("x-csrf-token", headerValue)
	}
	return req
}

func TestCsrfMiddleware(t *testing.T) {
	ctx := context.Background()

	t.Run("operationID is in ignore list: skip validation and pass through", func(t *testing.T) {
		// arrange
		spy := &spyHandler{}
		mw := middleware_d.CsrfMiddleware(map[string]struct{}{
			"GetAuthGoogle": {},
		})
		wrapped := mw(spy.fn(), "GetAuthGoogle")

		// action: cookie/headerが揃っていなくても素通りするはず
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, csrfRequest(http.MethodPost, "-", "-"), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	safeMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodTrace,
	}
	for _, method := range safeMethods {
		t.Run("safe method "+method+" is skipped", func(t *testing.T) {
			// arrange
			spy := &spyHandler{}
			mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
			wrapped := mw(spy.fn(), "Noop")

			// action: cookie/headerが揃っていなくても素通りするはず
			w := httptest.NewRecorder()
			_, err := wrapped(ctx, w, csrfRequest(method, "-", "-"), nil)

			// assert
			require.NoError(t, err)
			assert.True(t, spy.called)
		})
	}

	t.Run("POST without csrf_token cookie returns csrf_cookie_missing", func(t *testing.T) {
		// arrange
		spy := &spyHandler{}
		mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
		wrapped := mw(spy.fn(), "PostSomething")

		// action: cookieなし、headerあり
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, csrfRequest(http.MethodPost, "-", "abc"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "csrf_cookie_missing", decodeErrorTitle(t, w))
	})

	t.Run("POST without x-csrf-token header returns csrf_header_missing", func(t *testing.T) {
		// arrange
		spy := &spyHandler{}
		mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
		wrapped := mw(spy.fn(), "PostSomething")

		// action: cookieあり、headerなし
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, csrfRequest(http.MethodPost, "abc", "-"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "csrf_header_missing", decodeErrorTitle(t, w))
	})

	t.Run("POST with empty cookie value returns csrf_cookie_missing", func(t *testing.T) {
		// arrange
		spy := &spyHandler{}
		mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
		wrapped := mw(spy.fn(), "PostSomething")

		// action: cookieは存在するが値が空 → missing相当として扱う
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, csrfRequest(http.MethodPost, "", "abc"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "csrf_cookie_missing", decodeErrorTitle(t, w))
	})

	t.Run("POST with mismatched cookie/header returns csrf_mismatch", func(t *testing.T) {
		// arrange
		spy := &spyHandler{}
		mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
		wrapped := mw(spy.fn(), "PostSomething")

		// action: 値が異なる
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, csrfRequest(http.MethodPost, "cookie-value", "header-value"), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "csrf_mismatch", decodeErrorTitle(t, w))
	})

	t.Run("POST with matching cookie/header passes", func(t *testing.T) {
		// arrange
		spy := &spyHandler{}
		mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
		wrapped := mw(spy.fn(), "PostSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, csrfRequest(http.MethodPost, "same", "same"), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	// GET以外のstate-changingメソッドでもCSRF検証が動くことを確認
	stateChangingMethods := []string{
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}
	for _, method := range stateChangingMethods {
		t.Run(method+" with matching cookie/header passes", func(t *testing.T) {
			// arrange
			spy := &spyHandler{}
			mw := middleware_d.CsrfMiddleware(map[string]struct{}{})
			wrapped := mw(spy.fn(), "StateChange")

			// action
			w := httptest.NewRecorder()
			_, err := wrapped(ctx, w, csrfRequest(method, "same", "same"), nil)

			// assert
			require.NoError(t, err)
			assert.True(t, spy.called)
		})
	}
}
