package middleware_d_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/dbtype"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupScreenTimeUser はテスト用のユーザーを作成し、任意の画面時間範囲を
// 設定する。ranges には [start_seconds, end_seconds] のペア列を渡す。
// 空のスライスを渡すと範囲未設定ユーザーとなる。
func setupScreenTimeUser(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	dailyLimitSeconds int,
	ranges [][2]int,
) uuid.UUID {
	t.Helper()
	q := sqlc.New(db)

	authPublicID := uuid.Must(uuid.NewV7())
	authRow, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         "https://accounts.google.com",
		Sub:            "sub-" + uuid.Must(uuid.NewV7()).String(),
		LastLoggedInAt: time.Now(),
		PublicID:       authPublicID,
	})
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())
	mUserID, err := q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User",
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    dailyLimitSeconds,
		JoinedAt:                  time.Now(),
		PublicID:                  userID,
		UserAuthorizationPublicID: authRow.PublicID,
	})
	require.NoError(t, err)

	if len(ranges) > 0 {
		params := make([]sqlc.BulkInsertScreenTimeRangesParams, 0, len(ranges))
		for _, r := range ranges {
			params = append(params, sqlc.BulkInsertScreenTimeRangesParams{
				MUserID:              mUserID,
				ScreenTimeRangeStart: dbtype.Seconds(r[0]),
				ScreenTimeRangeEnd:   dbtype.Seconds(r[1]),
			})
		}
		_, err := q.BulkInsertScreenTimeRanges(ctx, params)
		require.NoError(t, err)
	}
	return userID
}

// fullDayRange は1日全体を覆う範囲。BlockedUntil が常に nil を返すようにする。
var fullDayRange = [2]int{0, 86400}

func TestScreenTimeMiddleware(t *testing.T) {
	ctx := context.Background()

	t.Run("operationID is in ignore list: skip validation and pass through", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		// ignore list対象なのでユーザーを作らなくても動くはず
		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{
			"GetUsersMeStatus": {},
		})
		wrapped := mw(spy.fn(), "GetUsersMeStatus")

		// action: user_id が ctx に居なくても問題なし (ignore listで早期return)
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("no user_id in ctx: pass through without touching DB", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(ctx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("no screen time ranges: blocked with outside_allowed_time_range", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		userID := setupScreenTimeUser(t, ctx, db, 3600, nil)

		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		assert.False(t, spy.called)
		require.Error(t, err)
		var domainErr *core.DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, "outside_allowed_time_range", domainErr.Code())
		assert.Equal(t, core.StatusForbidden, domainErr.StatusCode())
	})

	t.Run("all-day range + within daily limit: passes", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		userID := setupScreenTimeUser(t, ctx, db, 3600, [][2]int{fullDayRange})

		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("all-day range + daily limit 0: screen_time_limit_exceeded", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		// daily_screen_time_seconds=0 → 視聴履歴が無くても remaining=0 で即ブロック
		userID := setupScreenTimeUser(t, ctx, db, 0, [][2]int{fullDayRange})

		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		assert.False(t, spy.called)
		require.Error(t, err)
		var domainErr *core.DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, "screen_time_limit_exceeded", domainErr.Code())
		assert.Equal(t, core.StatusForbidden, domainErr.StatusCode())
	})

	t.Run("unlimited screen time: always passes regardless of consumed", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		// daily_screen_time_seconds >= 86400 で無制限扱い (history.FindTotalWatchSeconds がMaxIntを返す)
		userID := setupScreenTimeUser(t, ctx, db, 86401, [][2]int{fullDayRange})

		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{})
		wrapped := mw(spy.fn(), "GetSomething")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})

	t.Run("ignore list takes precedence over blocked state", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		// 範囲無しユーザー → 通常はブロックされる
		userID := setupScreenTimeUser(t, ctx, db, 0, nil)

		spy := &spyHandler{}
		mw := middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{
			"PatchUsersMeStatus": {},
		})
		wrapped := mw(spy.fn(), "PatchUsersMeStatus")
		reqCtx := hutil.WithUserID(ctx, userID)

		// action: ignore list 対象なので、制限中でもハンドラが呼ばれるはず
		w := httptest.NewRecorder()
		_, err := wrapped(reqCtx, w, httptest.NewRequest(http.MethodGet, "/foo", nil), nil)

		// assert
		require.NoError(t, err)
		assert.True(t, spy.called)
	})
}
