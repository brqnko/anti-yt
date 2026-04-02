package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// getAuthorizationPublicID は issuer+sub から authorization の public_id を取得する
func getAuthorizationPublicID(t *testing.T, pool *pgxpool.Pool, issuer, sub string) uuid.UUID {
	t.Helper()
	var publicID uuid.UUID
	err := pool.QueryRow(context.Background(),
		"SELECT public_id FROM m_user_authorization WHERE issuer = $1 AND sub = $2", issuer, sub,
	).Scan(&publicID)
	if err != nil {
		t.Fatalf("getAuthorizationPublicID: %v", err)
	}
	return publicID
}

// seedUserForAuth は GoogleOIDCCallback で作成された authorization に対してユーザーを作成する
func seedUserForAuth(t *testing.T, pool *pgxpool.Pool, sub string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	authPubID := getAuthorizationPublicID(t, pool, "https://accounts.google.com", sub)
	userPubID := uuid.Must(uuid.NewV7())
	q := sqlc.New(pool)
	_, err := q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User",
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    86400,
		JoinedAt:                  time.Now().UTC(),
		PublicID:                  userPubID,
		UserAuthorizationPublicID: authPubID,
	})
	if err != nil {
		t.Fatalf("seedUserForAuth: InsertUser failed: %v", err)
	}
	return userPubID
}

func TestService_CreateAuthCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "success"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			oidcMock := defaultOIDCMock("sub123")
			svc := newTestService(t, nil, oidcMock, testutil.DefaultJWTMock())

			url, csrf, err := svc.CreateAuthCode(context.Background(), "web")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if csrf == "" {
				t.Fatal("csrf token should not be empty")
			}
			if !strings.Contains(url, csrf) {
				t.Fatalf("url should contain csrf token: url=%s, csrf=%s", url, csrf)
			}
			if len(oidcMock.AuthCodeURLCalls()) != 1 {
				t.Fatalf("AuthCodeURL should be called once, got %d", len(oidcMock.AuthCodeURLCalls()))
			}
		})
	}
}

func TestService_GoogleOIDCCallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		csrf         string
		state        string
		code         string
		needsDB      bool
		seedUser     bool
		deactivate   bool
		exchangeErr  error
		wantErr      error
		wantRedirect string
	}{
		{
			name:    "empty_csrf",
			csrf:    "",
			state:   "some-state",
			code:    "code",
			wantErr: ErrInvalidCSRFOrState,
		},
		{
			name:    "empty_state",
			csrf:    "some-csrf",
			state:   "",
			code:    "code",
			wantErr: ErrInvalidCSRFOrState,
		},
		{
			name:    "csrf_ne_state",
			csrf:    "csrf-a",
			state:   "csrf-b",
			code:    "code",
			wantErr: ErrInvalidCSRF,
		},
		{
			name:        "exchange_error",
			csrf:        "valid",
			state:       "valid",
			code:        "bad-code",
			needsDB:     true,
			exchangeErr: errors.New("exchange failed"),
			wantErr:     errors.New("exchange failed"),
		},
		{
			name:         "new_user_register",
			csrf:         "valid-csrf",
			state:        "valid-csrf",
			code:         "auth-code",
			needsDB:      true,
			wantRedirect: "register",
		},
		{
			name:         "existing_user_dashboard",
			csrf:         "valid-csrf",
			state:        "valid-csrf",
			code:         "auth-code",
			needsDB:      true,
			seedUser:     true,
			wantRedirect: "dashboard",
		},
		{
			name:         "deactivated_user_reactivation",
			csrf:         "valid-csrf",
			state:        "valid-csrf",
			code:         "auth-code",
			needsDB:      true,
			seedUser:     true,
			deactivate:   true,
			wantRedirect: "reactivation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			sub := "google-sub-" + tt.name
			oidcMock := defaultOIDCMock(sub)
			if tt.exchangeErr != nil {
				oidcMock.ExchangeAndVerifyFunc = func(ctx context.Context, code string) (string, error) {
					return "", tt.exchangeErr
				}
			}
			jwtMock := testutil.DefaultJWTMock()

			var svc *Service
			if tt.needsDB {
				pool := testutil.NewTestDB(t)
				svc = newTestService(t, pool, oidcMock, jwtMock)

				if tt.seedUser {
					// まず一回コールバックを呼んで authorization + refresh token を作成
					_, _, _, _, _, _, _, err := svc.GoogleOIDCCallback(ctx, "setup", "setup", "auth-code", "127.0.0.1", "JP", "fp", "Mozilla/5.0")
					if err != nil {
						t.Fatalf("seed: GoogleOIDCCallback failed: %v", err)
					}

					userPubID := seedUserForAuth(t, pool, sub)

					if tt.deactivate {
						q := sqlc.New(pool)
						err := q.ArchiveUser(ctx, sqlc.ArchiveUserParams{
							LeaveReasonCode: 0,
							UserPublicID:    userPubID,
						})
						if err != nil {
							t.Fatalf("seed: ArchiveUser failed: %v", err)
						}
					}
				}
			} else {
				svc = newTestService(t, nil, oidcMock, jwtMock)
			}

			at, rt, _, redirect, _, _, _, err := svc.GoogleOIDCCallback(ctx, tt.csrf, tt.state, tt.code, "192.168.1.1", "US", "fingerprint", "Mozilla/5.0")

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if at == "" {
				t.Fatal("access token should not be empty")
			}
			if rt == "" {
				t.Fatal("refresh token should not be empty")
			}
			if redirect != tt.wantRedirect {
				t.Fatalf("expected redirect %q, got %q", tt.wantRedirect, redirect)
			}
		})
	}
}

func TestService_Logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "success"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)

			sub := "logout-sub"
			oidcMock := defaultOIDCMock(sub)
			jwtMock := testutil.DefaultJWTMock()
			svc := newTestService(t, pool, oidcMock, jwtMock)

			// セットアップ: authorization + refresh token を作成
			_, refreshTokenRaw, _, _, _, _, _, err := svc.GoogleOIDCCallback(ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "Mozilla/5.0")
			if err != nil {
				t.Fatalf("setup: GoogleOIDCCallback failed: %v", err)
			}

			userPubID := seedUserForAuth(t, pool, sub)

			// VerifyUserAccessToken がこのユーザーIDを返すようにモック設定
			jwtMock.VerifyUserAccessTokenFunc = func(token string) (uuid.UUID, uuid.UUID, time.Time, error) {
				return userPubID, uuid.Must(uuid.NewV7()), time.Now().UTC().Add(30 * time.Minute), nil
			}

			err = svc.Logout(ctx, "dummy-access-token", refreshTokenRaw)
			if err != nil {
				t.Fatalf("Logout failed: %v", err)
			}
		})
	}
}

func TestService_RefreshToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "invalid_token", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)

			sub := "refresh-sub-" + tt.name
			oidcMock := defaultOIDCMock(sub)
			jwtMock := testutil.DefaultJWTMock()
			svc := newTestService(t, pool, oidcMock, jwtMock)

			// セットアップ: authorization + refresh token を作成
			_, refreshTokenRaw, _, _, _, _, _, err := svc.GoogleOIDCCallback(ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "Mozilla/5.0")
			if err != nil {
				t.Fatalf("setup: GoogleOIDCCallback failed: %v", err)
			}

			// ユーザーを作成 (RotateRefreshToken が user public_id を返すため)
			seedUserForAuth(t, pool, sub)

			// updated_at を過去に設定 (RotateRefreshToken の updated_at < updated_at_for_check 条件を満たすため)
			_, err = pool.Exec(ctx, "UPDATE m_refresh_token SET updated_at = updated_at - interval '1 hour'")
			if err != nil {
				t.Fatalf("setup: update updated_at failed: %v", err)
			}

			var tokenToUse string
			if tt.wantErr {
				tokenToUse = "invalid-token-that-does-not-exist"
			} else {
				tokenToUse = refreshTokenRaw
			}

			newRT, newAT, _, _, err := svc.RefreshToken(ctx, tokenToUse, "192.168.1.1", "US", "fp2", "Chrome/120")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if newRT == "" {
				t.Fatal("new refresh token should not be empty")
			}
			if newAT == "" {
				t.Fatal("new access token should not be empty")
			}
		})
	}
}

func TestService_GetSessions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenCount  int
		limit       int32
		wantCount   int
		wantHasNext bool
	}{
		{name: "empty", tokenCount: 0, limit: 10, wantCount: 0, wantHasNext: false},
		{name: "returns_sessions", tokenCount: 3, limit: 10, wantCount: 3, wantHasNext: false},
		{name: "pagination_has_next", tokenCount: 3, limit: 2, wantCount: 2, wantHasNext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)

			sub := "sessions-sub-" + tt.name
			oidcMock := defaultOIDCMock(sub)
			jwtMock := testutil.DefaultJWTMock()
			svc := newTestService(t, pool, oidcMock, jwtMock)

			// セットアップ: 複数の refresh token を作成
			for i := 0; i < tt.tokenCount; i++ {
				_, _, _, _, _, _, _, err := svc.GoogleOIDCCallback(ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "Mozilla/5.0")
				if err != nil {
					t.Fatalf("setup: GoogleOIDCCallback[%d] failed: %v", i, err)
				}
			}

			// ユーザーを作成
			var userPubID uuid.UUID
			if tt.tokenCount > 0 {
				userPubID = seedUserForAuth(t, pool, sub)
			} else {
				userPubID = uuid.Must(uuid.NewV7())
			}

			sessions, hasNext, err := svc.GetSessions(ctx, userPubID, nil, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sessions) != tt.wantCount {
				t.Fatalf("expected %d sessions, got %d", tt.wantCount, len(sessions))
			}
			if hasNext != tt.wantHasNext {
				t.Fatalf("expected hasNext=%v, got %v", tt.wantHasNext, hasNext)
			}
		})
	}
}

func TestService_RemoveSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "not_found", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)

			sub := "remove-sub-" + tt.name
			oidcMock := defaultOIDCMock(sub)
			jwtMock := testutil.DefaultJWTMock()
			svc := newTestService(t, pool, oidcMock, jwtMock)

			// セットアップ: refresh token を作成
			_, _, _, _, _, _, _, err := svc.GoogleOIDCCallback(ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "Mozilla/5.0")
			if err != nil {
				t.Fatalf("setup: GoogleOIDCCallback failed: %v", err)
			}

			userPubID := seedUserForAuth(t, pool, sub)

			// セッション一覧を取得
			sessions, _, err := svc.GetSessions(ctx, userPubID, nil, 100)
			if err != nil {
				t.Fatalf("setup: GetSessions failed: %v", err)
			}

			var sessionID uuid.UUID
			if tt.wantErr {
				sessionID = uuid.Must(uuid.NewV7()) // 存在しない ID
			} else {
				if len(sessions) == 0 {
					t.Fatal("setup: no sessions found")
				}
				sessionID = sessions[0].ID
			}

			_, err = svc.RemoveSession(ctx, userPubID, sessionID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if !errors.Is(err, core.ErrNotFound) {
					t.Fatalf("expected core.ErrNotFound, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
