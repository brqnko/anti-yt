package user

import (
	"context"
	"errors"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestService_CreateNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		displayName      string
		languageCode     string
		dailyScreenLimit *int
		screenLimits     []struct{ Start, End int }
		verifyErr        error
		wantErr          bool
	}{
		{
			name:         "success_unlimited",
			displayName:  "Test User",
			languageCode: "ja",
		},
		{
			name:             "success_with_screen_limit",
			displayName:      "Test User",
			languageCode:     "ja",
			dailyScreenLimit: intPtr(3600),
			screenLimits:     []struct{ Start, End int }{{Start: 0, End: 3600}},
		},
		{
			name:         "invalid_register_token",
			displayName:  "Test User",
			languageCode: "ja",
			verifyErr:    errors.New("invalid token"),
			wantErr:      true,
		},
		{
			name:         "invalid_display_name_empty",
			displayName:  "",
			languageCode: "ja",
			wantErr:      true,
		},
		{
			name:         "invalid_language_code",
			displayName:  "Test User",
			languageCode: "xx",
			wantErr:      true,
		},
		{
			name:             "invalid_screen_time_out_of_range",
			displayName:      "Test User",
			languageCode:     "ja",
			dailyScreenLimit: intPtr(-1),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			pool := testutil.NewTestDB(t)
			authPubID := testutil.SeedAuthorization(t, pool)

			jwtMock := testutil.DefaultJWTMock()
			jti := uuid.Must(uuid.NewV7())
			if tt.verifyErr != nil {
				jwtMock.VerifyRegisterTokenFunc = func(token string) (uuid.UUID, uuid.UUID, error) {
					return uuid.Nil, uuid.Nil, tt.verifyErr
				}
			} else {
				jwtMock.VerifyRegisterTokenFunc = func(token string) (uuid.UUID, uuid.UUID, error) {
					return authPubID, jti, nil
				}
			}

			svc := newTestService(t, pool, jwtMock)

			user, accessToken, expiresAt, err := svc.CreateNewUser(ctx, "register-token", tt.dailyScreenLimit, tt.screenLimits, tt.displayName, tt.languageCode)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Fatal("user should not be nil")
			}
			if user.DisplayName.String() != tt.displayName {
				t.Fatalf("expected display name %q, got %q", tt.displayName, user.DisplayName.String())
			}
			if user.LanguageCode.String() != tt.languageCode {
				t.Fatalf("expected language code %q, got %q", tt.languageCode, user.LanguageCode.String())
			}
			if accessToken == "" {
				t.Fatal("access token should not be empty")
			}
			if expiresAt.IsZero() {
				t.Fatal("expiresAt should not be zero")
			}
		})
	}
}

func TestService_CreateNewUser_AlreadyRegistered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pool := testutil.NewTestDB(t)
	authPubID := testutil.SeedAuthorization(t, pool)
	jti := uuid.Must(uuid.NewV7())

	jwtMock := testutil.DefaultJWTMock()
	jwtMock.VerifyRegisterTokenFunc = func(token string) (uuid.UUID, uuid.UUID, error) {
		return authPubID, jti, nil
	}

	svc := newTestService(t, pool, jwtMock)

	// 1回目: 成功
	_, _, _, err := svc.CreateNewUser(ctx, "register-token", nil, nil, "Test User", "ja")
	if err != nil {
		t.Fatalf("first CreateNewUser failed: %v", err)
	}

	// 2回目: 既に登録済み
	_, _, _, err = svc.CreateNewUser(ctx, "register-token", nil, nil, "Test User 2", "ja")
	if err == nil {
		t.Fatal("expected error for already registered, got nil")
	}
}

func TestService_EditUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		newDisplayName  *string
		newLanguageCode *string
		newScreenLimit  *int
		newScreenRanges *[]struct{ Start, End int }
		wantErr         bool
	}{
		{
			name:           "update_display_name",
			newDisplayName: strPtr("New Name"),
		},
		{
			name:           "update_screen_limit",
			newScreenLimit: intPtr(7200),
		},
		{
			name:            "update_screen_ranges",
			newScreenRanges: &[]struct{ Start, End int }{{Start: 0, End: 3600}, {Start: 7200, End: 10800}},
		},
		{
			name: "no_changes",
		},
		{
			name:           "invalid_display_name_too_long",
			newDisplayName: strPtr("あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほ"),
			wantErr:        true,
		},
		{
			name:            "invalid_language_code",
			newLanguageCode: strPtr("xx"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			pool := testutil.NewTestDB(t)
			authPubID := testutil.SeedAuthorization(t, pool)
			jti := uuid.Must(uuid.NewV7())

			jwtMock := testutil.DefaultJWTMock()
			jwtMock.VerifyRegisterTokenFunc = func(token string) (uuid.UUID, uuid.UUID, error) {
				return authPubID, jti, nil
			}

			svc := newTestService(t, pool, jwtMock)

			// セットアップ: ユーザー作成
			created, _, _, err := svc.CreateNewUser(ctx, "register-token", nil, nil, "Original Name", "ja")
			if err != nil {
				t.Fatalf("setup: CreateNewUser failed: %v", err)
			}

			updated, err := svc.EditUser(ctx, created.ID, tt.newDisplayName, tt.newLanguageCode, tt.newScreenLimit, tt.newScreenRanges)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.newDisplayName != nil {
				if updated.DisplayName.String() != *tt.newDisplayName {
					t.Fatalf("expected display name %q, got %q", *tt.newDisplayName, updated.DisplayName.String())
				}
			}
			if tt.newScreenLimit != nil {
				if updated.ScreenTimeLimit.Seconds() != *tt.newScreenLimit {
					t.Fatalf("expected screen limit %d, got %d", *tt.newScreenLimit, updated.ScreenTimeLimit.Seconds())
				}
			}
		})
	}
}

func TestService_EditUser_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pool := testutil.NewTestDB(t)
	jwtMock := testutil.DefaultJWTMock()
	svc := newTestService(t, pool, jwtMock)

	name := "New Name"
	_, err := svc.EditUser(ctx, uuid.Must(uuid.NewV7()), &name, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestService_GetUserStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		screenLimit  *int
		screenRanges []struct{ Start, End int }
		wantRanges   int
	}{
		{
			name:       "unlimited",
			wantRanges: 0,
		},
		{
			name:         "with_screen_limit",
			screenLimit:  intPtr(3600),
			screenRanges: []struct{ Start, End int }{{Start: 0, End: 3600}},
			wantRanges:   1,
		},
		{
			name:         "with_multiple_ranges",
			screenLimit:  intPtr(7200),
			screenRanges: []struct{ Start, End int }{{Start: 0, End: 3600}, {Start: 7200, End: 10800}},
			wantRanges:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			pool := testutil.NewTestDB(t)
			authPubID := testutil.SeedAuthorization(t, pool)
			jti := uuid.Must(uuid.NewV7())

			jwtMock := testutil.DefaultJWTMock()
			jwtMock.VerifyRegisterTokenFunc = func(token string) (uuid.UUID, uuid.UUID, error) {
				return authPubID, jti, nil
			}

			svc := newTestService(t, pool, jwtMock)

			created, _, _, err := svc.CreateNewUser(ctx, "register-token", tt.screenLimit, tt.screenRanges, "Test User", "ja")
			if err != nil {
				t.Fatalf("setup: CreateNewUser failed: %v", err)
			}

			view, err := svc.GetUserStatus(ctx, created.ID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if view.UserID != created.ID {
				t.Fatalf("expected user ID %s, got %s", created.ID, view.UserID)
			}
			if view.DisplayName != "Test User" {
				t.Fatalf("expected display name %q, got %q", "Test User", view.DisplayName)
			}
			if len(view.ScreenTimeLimitRange) != tt.wantRanges {
				t.Fatalf("expected %d ranges, got %d", tt.wantRanges, len(view.ScreenTimeLimitRange))
			}
			if tt.screenLimit != nil {
				if view.DailyScreenSeconds == nil {
					t.Fatal("expected DailyScreenSeconds to be non-nil")
				}
				if *view.DailyScreenSeconds != *tt.screenLimit {
					t.Fatalf("expected screen seconds %d, got %d", *tt.screenLimit, *view.DailyScreenSeconds)
				}
			} else {
				if view.DailyScreenSeconds != nil {
					t.Fatalf("expected DailyScreenSeconds to be nil, got %d", *view.DailyScreenSeconds)
				}
			}
		})
	}
}

func TestService_GetUserStatus_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pool := testutil.NewTestDB(t)
	jwtMock := testutil.DefaultJWTMock()
	svc := newTestService(t, pool, jwtMock)

	_, err := svc.GetUserStatus(ctx, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestService_RemoveUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "success", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			pool := testutil.NewTestDB(t)
			authPubID := testutil.SeedAuthorization(t, pool)
			jti := uuid.Must(uuid.NewV7())

			jwtMock := testutil.DefaultJWTMock()
			jwtMock.VerifyRegisterTokenFunc = func(token string) (uuid.UUID, uuid.UUID, error) {
				return authPubID, jti, nil
			}

			svc := newTestService(t, pool, jwtMock)

			created, _, _, err := svc.CreateNewUser(ctx, "register-token", nil, nil, "Test User", "ja")
			if err != nil {
				t.Fatalf("setup: CreateNewUser failed: %v", err)
			}

			err = svc.RemoveUser(ctx, created.ID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 削除後は取得できないことを確認
			_, err = svc.GetUserStatus(ctx, created.ID)
			if err == nil {
				t.Fatal("expected error after removal, got nil")
			}
		})
	}
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
