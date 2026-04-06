package middleware_d_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// returningHandler は固定の (response, err) を返すだけのStrictHandlerFunc。
// WrapErrorMiddleware の挙動をテストするための簡易ハンドラ。
func returningHandler(resp interface{}, err error) v1.StrictHandlerFunc {
	return func(_ context.Context, _ http.ResponseWriter, _ *http.Request, _ interface{}) (interface{}, error) {
		return resp, err
	}
}

// noopDiscord は呼び出しを無視するDiscordモック (呼ばれないことを期待するテスト用)。
func noopDiscord(t *testing.T) *DiscordServiceMock {
	t.Helper()
	return &DiscordServiceMock{
		SendWebhookMessageFunc: func(_ context.Context, _ string) error {
			t.Fatal("Discord.SendWebhookMessage should not be called")
			return nil
		},
	}
}

func TestWrapErrorMiddleware(t *testing.T) {
	ctx := context.Background()

	t.Run("handler success: passthrough response and no write to ResponseWriter", func(t *testing.T) {
		// arrange
		reportSvc := report.NewService(report.WithDiscord(noopDiscord(t)))
		mw := middleware_d.WrapErrorMiddleware(reportSvc)
		wrapped := mw(returningHandler("ok", nil), "GetSomething")

		// action
		w := httptest.NewRecorder()
		resp, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.Empty(t, w.Body.String()) // middleware は何も書き込まないはず
	})

	t.Run("DomainError (404): writes correct JSON with NotFound status", func(t *testing.T) {
		// arrange
		reportSvc := report.NewService(report.WithDiscord(noopDiscord(t)))
		mw := middleware_d.WrapErrorMiddleware(reportSvc)
		wrapped := mw(returningHandler(nil, core.ErrNotFound), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "not_found", decodeErrorTitle(t, w))
	})

	t.Run("DomainError (403): writes correct JSON with Forbidden status", func(t *testing.T) {
		// arrange
		customErr := core.NewDomainError("custom_forbidden", "custom detail message", core.StatusForbidden)
		reportSvc := report.NewService(report.WithDiscord(noopDiscord(t)))
		mw := middleware_d.WrapErrorMiddleware(reportSvc)
		wrapped := mw(returningHandler(nil, customErr), "PostSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodPost, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "custom_forbidden", decodeErrorTitle(t, w))
	})

	t.Run("wrapped DomainError (fmt.Errorf %w): still detected by errors.As", func(t *testing.T) {
		// arrange
		wrapped := fmt.Errorf("context: %w", core.ErrNotFound)
		reportSvc := report.NewService(report.WithDiscord(noopDiscord(t)))
		mw := middleware_d.WrapErrorMiddleware(reportSvc)
		handler := mw(returningHandler(nil, wrapped), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := handler(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "not_found", decodeErrorTitle(t, w))
	})

	t.Run("non-DomainError: returns 500 and reports to Discord asynchronously", func(t *testing.T) {
		// arrange
		unexpectedErr := errors.New("db connection lost")
		// 非同期で呼ばれるのでchannelで受け取る
		sent := make(chan string, 1)
		discordMock := &DiscordServiceMock{
			SendWebhookMessageFunc: func(_ context.Context, message string) error {
				sent <- message
				return nil
			},
		}
		reportSvc := report.NewService(report.WithDiscord(discordMock))
		mw := middleware_d.WrapErrorMiddleware(reportSvc)
		wrapped := mw(returningHandler(nil, unexpectedErr), "GetFailing")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert: HTTP response はDiscord呼び出しを待たずに返るはず
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal_server_error", decodeErrorTitle(t, w))

		// assert: Discord にエラー通知が送られている(非同期のため少し待つ)
		select {
		case msg := <-sent:
			assert.Contains(t, msg, "GetFailing")         // operationIDが含まれる
			assert.Contains(t, msg, "db connection lost") // 元errorのメッセージも含まれる
		case <-time.After(2 * time.Second):
			t.Fatal("Discord report was not sent within 2s")
		}
	})

	t.Run("non-DomainError: works even when Discord is not configured", func(t *testing.T) {
		// arrange: Discord未設定 (report.WithDiscord を呼ばない)
		reportSvc := report.NewService()
		mw := middleware_d.WrapErrorMiddleware(reportSvc)
		wrapped := mw(returningHandler(nil, errors.New("boom")), "GetFailing")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert: Discord通知はスキップされるが500レスポンスは正しく返る
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal_server_error", decodeErrorTitle(t, w))
	})
}
