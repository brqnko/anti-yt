package middleware_d_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRatelimitMiddleware(t *testing.T) {
	ctx := context.Background()

	t.Run("no user_id in ctx: pass through without touching DB", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(db, 10, map[string]int{})
		wrapped := mw(spy.fn(), "AnyOperation")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("operationID has custom quota: consumes specified amount", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := uuid.Must(uuid.NewV7())
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(db, 1000, map[string]int{
			"GetSearch": 100,
		})
		wrapped := mw(spy.fn(), "GetSearch")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)

		// DBに100消費した状態が記録されているはず
		row, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{UserID: userID, Quota: 0})
		require.NoError(t, err)
		assert.Equal(t, 100, row.ConsumedQuota)
	})

	t.Run("operationID not in quota map: falls back to default quota 1", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := uuid.Must(uuid.NewV7())
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(db, 1000, map[string]int{
			"GetSearch": 100,
		})
		wrapped := mw(spy.fn(), "GetSomethingElse")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)

		// 未登録operationなのでdefault=1だけ消費されているはず
		row, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{UserID: userID, Quota: 0})
		require.NoError(t, err)
		assert.Equal(t, 1, row.ConsumedQuota)
	})

	t.Run("within limit: passes", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := uuid.Must(uuid.NewV7())
		// 既に limit - 1 まで消費している状態を作る
		_, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{
			UserID: userID,
			Quota:  9, // limit 10 の直前
		})
		require.NoError(t, err)

		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(db, 10, map[string]int{})
		wrapped := mw(spy.fn(), "NoopOperation") // default quota=1
		reqCtx := hutil.WithUserID(ctx, userID)

		// action: pre-request = 9 < 10 なので通る
		w := httptest.NewRecorder()
		_, err = wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("over limit: returns 429 and does not call handler", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := uuid.Must(uuid.NewV7())
		// 既に limit ちょうどまで消費している状態
		_, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{
			UserID: userID,
			Quota:  10, // limit と同値
		})
		require.NoError(t, err)

		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(db, 10, map[string]int{})
		wrapped := mw(spy.fn(), "NoopOperation") // default quota=1
		reqCtx := hutil.WithUserID(ctx, userID)

		// action: pre-request = 10 >= 10 なので拒否
		w := httptest.NewRecorder()
		_, err = wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Equal(t, "too_many_requests", decodeErrorTitle(t, w))
	})

	t.Run("accumulates across multiple requests", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := uuid.Must(uuid.NewV7())
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(db, 1000, map[string]int{
			"GetFoo": 3,
		})
		wrapped := mw(spy.fn(), "GetFoo")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action: 3回呼ぶ → 累計9
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)
			require.NoError(t, err)
		}

		// assert
		row, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{UserID: userID, Quota: 0})
		require.NoError(t, err)
		assert.Equal(t, 9, row.ConsumedQuota)
	})
}
