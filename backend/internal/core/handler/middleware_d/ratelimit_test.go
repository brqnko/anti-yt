package middleware_d_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRatelimitRepo は RatelimitRepository の in-memory 実装。
// 本番の Redis 実装と同じく「Consume は無条件に加算し、加算後の累計を返す」
// セマンティクスを再現する。
type fakeRatelimitRepo struct {
	mu       sync.Mutex
	consumed map[uuid.UUID]int
	err      error
}

func newFakeRatelimitRepo() *fakeRatelimitRepo {
	return &fakeRatelimitRepo{consumed: map[uuid.UUID]int{}}
}

func (r *fakeRatelimitRepo) Consume(_ context.Context, userID uuid.UUID, quota int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	r.consumed[userID] += quota
	return r.consumed[userID], nil
}

func (r *fakeRatelimitRepo) set(userID uuid.UUID, v int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.consumed[userID] = v
}

func (r *fakeRatelimitRepo) get(userID uuid.UUID) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.consumed[userID]
}

var _ database_d.RatelimitRepository = (*fakeRatelimitRepo)(nil)

func TestUserRatelimitMiddleware(t *testing.T) {
	ctx := context.Background()

	t.Run("no user_id in ctx: pass through without touching repo", func(t *testing.T) {
		// arrange
		repo := newFakeRatelimitRepo()
		repo.err = errors.New("must not be called") // 呼ばれたら即失敗させる
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 10, map[string]int{})
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
		repo := newFakeRatelimitRepo()
		userID := uuid.Must(uuid.NewV7())
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 1000, map[string]int{
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
		assert.Equal(t, 100, repo.get(userID))
	})

	t.Run("operationID not in quota map: skips rate limit check", func(t *testing.T) {
		// arrange
		repo := newFakeRatelimitRepo()
		userID := uuid.Must(uuid.NewV7())
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 1000, map[string]int{
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
		assert.Equal(t, 0, repo.get(userID)) // repoに触れていない
	})

	t.Run("within limit: passes", func(t *testing.T) {
		// arrange
		repo := newFakeRatelimitRepo()
		userID := uuid.Must(uuid.NewV7())
		repo.set(userID, 9) // limit 10 の直前

		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 10, map[string]int{"NoopOperation": 1})
		wrapped := mw(spy.fn(), "NoopOperation")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action: pre-request = 9 < 10 なので通る
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("over limit: returns 429 and does not call handler", func(t *testing.T) {
		// arrange
		repo := newFakeRatelimitRepo()
		userID := uuid.Must(uuid.NewV7())
		repo.set(userID, 10) // 既に limit ちょうどまで消費している状態

		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 10, map[string]int{"NoopOperation": 1})
		wrapped := mw(spy.fn(), "NoopOperation")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action: pre-request = 10 >= 10 なので拒否
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Equal(t, "too_many_requests", decodeErrorTitle(t, w))
	})

	t.Run("accumulates across multiple requests", func(t *testing.T) {
		// arrange
		repo := newFakeRatelimitRepo()
		userID := uuid.Must(uuid.NewV7())
		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 1000, map[string]int{
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
		assert.Equal(t, 9, repo.get(userID))
	})

	t.Run("repo error: returns 500 and does not call handler", func(t *testing.T) {
		// arrange
		repo := newFakeRatelimitRepo()
		repo.err = errors.New("redis down")
		userID := uuid.Must(uuid.NewV7())

		spy := &spyHandler{}
		mw := middleware_d.UserRatelimitMiddleware(repo, 10, map[string]int{"NoopOperation": 1})
		wrapped := mw(spy.fn(), "NoopOperation")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.False(t, spy.called)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal_server_error", decodeErrorTitle(t, w))
	})
}
