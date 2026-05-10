package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateAuthCode(t *testing.T) {
	ctx := context.Background()

	t.Run("success with platform", func(t *testing.T) {
		// arrange
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{
				AuthCodeURLFunc: func(state string) string {
					return "https://accounts.google.com/o/oauth2/auth?state=" + state
				},
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				SignOIDCStateTokenFunc: func(platform, _ string) (string, error) {
					return "state-token-" + platform, nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		authURL, stateToken, err := svc.CreateAuthCode(ctx, "ios")

		// assert
		require.NoError(t, err)
		assert.Equal(t, "state-token-ios", stateToken)
		assert.Contains(t, authURL, "state-token-ios")
	})

	t.Run("empty platform defaults to web", func(t *testing.T) {
		// arrange
		var capturedPlatform string
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{
				AuthCodeURLFunc: func(state string) string { return "https://example.com" },
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				SignOIDCStateTokenFunc: func(platform, _ string) (string, error) {
					capturedPlatform = platform
					return "state-token", nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, _, err := svc.CreateAuthCode(ctx, "")

		// assert
		require.NoError(t, err)
		assert.Equal(t, "web", capturedPlatform)
	})

	t.Run("jwt sign error", func(t *testing.T) {
		// arrange
		signErr := errors.New("sign failed")
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				SignOIDCStateTokenFunc: func(_, _ string) (string, error) {
					return "", signErr
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, _, err := svc.CreateAuthCode(ctx, "web")

		// assert
		assert.ErrorIs(t, err, signErr)
	})
}

func TestService_GoogleOIDCCallback(t *testing.T) {
	ctx := context.Background()
	const sub = "google-sub-123"

	t.Run("empty csrf returns error", func(t *testing.T) {
		// arrange
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, err := svc.GoogleOIDCCallback(
			ctx, "", "state", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		assert.ErrorIs(t, err, auth.ErrInvalidCSRFOrState)
	})

	t.Run("csrf mismatch returns error", func(t *testing.T) {
		// arrange
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, err := svc.GoogleOIDCCallback(
			ctx, "a", "b", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		assert.ErrorIs(t, err, auth.ErrInvalidCSRF)
	})

	t.Run("invalid state token returns error", func(t *testing.T) {
		// arrange
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyOIDCStateTokenFunc: func(_ string) (string, error) {
					return "", auth.ErrInvalidCSRFOrState
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, err := svc.GoogleOIDCCallback(
			ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		assert.ErrorIs(t, err, auth.ErrInvalidCSRFOrState)
	})

	t.Run("invalid oidc code returns error", func(t *testing.T) {
		// arrange
		exchangeErr := errors.New("invalid code")
		svc := auth.NewService(
			testutil.NewTestPool(t),
			new(GoogleClientMock{
				ExchangeAndVerifyFunc: func(_ context.Context, _ string) (string, error) {
					return "", exchangeErr
				},
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyOIDCStateTokenFunc: func(_ string) (string, error) { return "web", nil },
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, err := svc.GoogleOIDCCallback(
			ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		assert.ErrorIs(t, err, exchangeErr)
	})

	t.Run("new user redirects to register", func(t *testing.T) {
		// arrange
		svc := auth.NewService(
			testutil.NewTestPool(t),
			new(GoogleClientMock{
				ExchangeAndVerifyFunc: func(_ context.Context, _ string) (string, error) { return sub, nil },
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyOIDCStateTokenFunc: func(_ string) (string, error) { return "web", nil },
				SignRegisterTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "register-token", time.Now().Add(time.Hour), nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		result, err := svc.GoogleOIDCCallback(
			ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "register", result.RedirectPath)
		assert.Equal(t, "web", result.Platform)
	})

	t.Run("existing user redirects to dashboard", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            sub,
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  uuid.Must(uuid.NewV7()),
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{
				ExchangeAndVerifyFunc: func(_ context.Context, _ string) (string, error) { return sub, nil },
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyOIDCStateTokenFunc: func(_ string) (string, error) { return "web", nil },
				SignUserAccessTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "access-token", time.Now().Add(time.Hour), nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		result, err := svc.GoogleOIDCCallback(
			ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "", result.RedirectPath)
		assert.Equal(t, "web", result.Platform)
	})

	t.Run("same user login twice succeeds", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            sub,
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  uuid.Must(uuid.NewV7()),
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{
				ExchangeAndVerifyFunc: func(_ context.Context, _ string) (string, error) { return sub, nil },
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyOIDCStateTokenFunc: func(_ string) (string, error) { return "web", nil },
				SignUserAccessTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "access-token", time.Now().Add(time.Hour), nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action & assert
		for range 2 {
			result, err := svc.GoogleOIDCCallback(
				ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "ua",
			)
			require.NoError(t, err)
			assert.Equal(t, "", result.RedirectPath)
		}
	})

	t.Run("deactivated user redirects to reactivation", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            sub,
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		require.NoError(t, q.ArchiveUser(ctx, sqlc.ArchiveUserParams{
			LeaveReasonCode: 1,
			UserPublicID:    userPublicID,
		}))

		svc := auth.NewService(
			db,
			new(GoogleClientMock{
				ExchangeAndVerifyFunc: func(_ context.Context, _ string) (string, error) { return sub, nil },
			}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyOIDCStateTokenFunc: func(_ string) (string, error) { return "web", nil },
				SignRegisterTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "register-token", time.Now().Add(time.Hour), nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		result, err := svc.GoogleOIDCCallback(
			ctx, "csrf", "csrf", "code", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "reactivation", result.RedirectPath)
		assert.Equal(t, "web", result.Platform)
	})
}

func TestService_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		rawToken := "raw-refresh-token"
		_, err = q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
			MUserAuthorizationID: authResult.MUserAuthorizationID,
			TokenHash:            util.Sha256Hex(rawToken),
			Generation:           1,
			PublicID:             uuid.Must(uuid.NewV7()),
			IpAddress:            "127.0.0.1",
			DeviceFingerprint:    "fp",
			UserAgent:            "ua",
			CountryCode:          "JP",
			CityName:             "",
			BrowserName:          "test",
			DeviceType:           "test",
			ExpiresAt:            time.Now().Add(7 * 24 * time.Hour),
			AccessTokenJti:       uuid.Must(uuid.NewV7()),
			ActivatedAt:          time.Now(),
			LastLoggedInAt:       time.Now(),
		})
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
					return userPublicID, uuid.Nil, time.Time{}, nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{
				InsertJTIFunc: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
					return nil
				},
			}),
		)

		// action
		err = svc.Logout(ctx, "access-token", rawToken)

		// assert
		require.NoError(t, err)
	})

	t.Run("nonexistent refresh token returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyUserAccessTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, time.Time, error) {
					return userPublicID, uuid.Nil, time.Time{}, nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		err = svc.Logout(ctx, "access-token", "nonexistent-token")

		// assert
		assert.ErrorIs(t, err, core.ErrNotFound)
	})
}

func TestService_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		rawToken := "raw-refresh-token"
		_, err = q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
			MUserAuthorizationID: authResult.MUserAuthorizationID,
			TokenHash:            util.Sha256Hex(rawToken),
			Generation:           1,
			PublicID:             uuid.Must(uuid.NewV7()),
			IpAddress:            "127.0.0.1",
			DeviceFingerprint:    "fp",
			UserAgent:            "ua",
			CountryCode:          "JP",
			CityName:             "",
			BrowserName:          "test",
			DeviceType:           "test",
			ExpiresAt:            time.Now().Add(7 * 24 * time.Hour),
			AccessTokenJti:       uuid.Must(uuid.NewV7()),
			ActivatedAt:          time.Now(),
			LastLoggedInAt:       time.Now(),
		})
		require.NoError(t, err)

		// RotateRefreshTokenのWHERE条件 "updated_at < now - TokenDuration" を満たすために
		// updated_atを過去にずらす
		_, err = db.Exec(ctx, "UPDATE m_refresh_token SET updated_at = $1 WHERE token_hash = $2",
			time.Now().Add(-30*time.Minute), util.Sha256Hex(rawToken))
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				SignUserAccessTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "new-access-token", time.Now().Add(15 * time.Minute), nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		newRefreshToken, accessToken, _, _, err := svc.RefreshToken(
			ctx, rawToken, "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		require.NoError(t, err)
		assert.NotEmpty(t, newRefreshToken)
		assert.Equal(t, "new-access-token", accessToken)
	})

	t.Run("nonexistent token returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, _, _, _, err := svc.RefreshToken(
			ctx, "nonexistent-token", "127.0.0.1", "JP", "fp", "ua",
		)

		// assert
		assert.ErrorIs(t, err, core.ErrNotFound)
	})
}

func TestService_GetSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("returns sessions", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		for range 3 {
			_, err = q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
				MUserAuthorizationID: authResult.MUserAuthorizationID,
				TokenHash:            util.Sha256Hex(uuid.Must(uuid.NewV7()).String()),
				Generation:           1,
				PublicID:             uuid.Must(uuid.NewV7()),
				IpAddress:            "127.0.0.1",
				DeviceFingerprint:    "fp",
				UserAgent:            "ua",
				CountryCode:          "JP",
				CityName:             "",
				BrowserName:          "test",
				DeviceType:           "test",
				ExpiresAt:            time.Now().Add(7 * 24 * time.Hour),
				AccessTokenJti:       uuid.Must(uuid.NewV7()),
				ActivatedAt:          time.Now(),
				LastLoggedInAt:       time.Now(),
			})
			require.NoError(t, err)
		}

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		sessions, hasMore, err := svc.GetSessions(ctx, userPublicID, nil, 10)

		// assert
		require.NoError(t, err)
		assert.Len(t, sessions, 3)
		assert.False(t, hasMore)
	})

	t.Run("pagination with hasMore", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		for range 3 {
			_, err = q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
				MUserAuthorizationID: authResult.MUserAuthorizationID,
				TokenHash:            util.Sha256Hex(uuid.Must(uuid.NewV7()).String()),
				Generation:           1,
				PublicID:             uuid.Must(uuid.NewV7()),
				IpAddress:            "127.0.0.1",
				DeviceFingerprint:    "fp",
				UserAgent:            "ua",
				CountryCode:          "JP",
				CityName:             "",
				BrowserName:          "test",
				DeviceType:           "test",
				ExpiresAt:            time.Now().Add(7 * 24 * time.Hour),
				AccessTokenJti:       uuid.Must(uuid.NewV7()),
				ActivatedAt:          time.Now(),
				LastLoggedInAt:       time.Now(),
			})
			require.NoError(t, err)
		}

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		sessions, hasMore, err := svc.GetSessions(ctx, userPublicID, nil, 2)

		// assert
		require.NoError(t, err)
		assert.Len(t, sessions, 2)
		assert.True(t, hasMore)
	})

	t.Run("no sessions", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		sessions, hasMore, err := svc.GetSessions(ctx, userPublicID, nil, 10)

		// assert
		require.NoError(t, err)
		assert.Empty(t, sessions)
		assert.False(t, hasMore)
	})
}

func TestService_RemoveSession(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		sessionPublicID := uuid.Must(uuid.NewV7())
		accessTokenJti := uuid.Must(uuid.NewV7())
		_, err = q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
			MUserAuthorizationID: authResult.MUserAuthorizationID,
			TokenHash:            util.Sha256Hex("token"),
			Generation:           1,
			PublicID:             sessionPublicID,
			IpAddress:            "127.0.0.1",
			DeviceFingerprint:    "fp",
			UserAgent:            "ua",
			CountryCode:          "JP",
			CityName:             "",
			BrowserName:          "test",
			DeviceType:           "test",
			ExpiresAt:            time.Now().Add(7 * 24 * time.Hour),
			AccessTokenJti:       accessTokenJti,
			ActivatedAt:          time.Now(),
			LastLoggedInAt:       time.Now(),
		})
		require.NoError(t, err)

		jtiRepo := new(JtiBlacklistRepositoryMock{
			InsertJTIFunc: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
				return nil
			},
		})
		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
			}),
			15*time.Minute, 7*24*time.Hour,
			jtiRepo,
		)

		// action
		removedID, err := svc.RemoveSession(ctx, userPublicID, sessionPublicID)

		// assert
		require.NoError(t, err)
		assert.Equal(t, sessionPublicID, removedID)
		// 削除した session の access_token_jti が blacklist 入りしていることを確認
		insertCalls := jtiRepo.InsertJTICalls()
		require.Len(t, insertCalls, 1)
		assert.Equal(t, accessTokenJti, insertCalls[0].Jti)
	})

	t.Run("nonexistent session returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       uuid.Must(uuid.NewV7()),
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, err = svc.RemoveSession(ctx, userPublicID, uuid.Must(uuid.NewV7()))

		// assert
		assert.ErrorIs(t, err, core.ErrNotFound)
	})
}

func TestService_CreateYouTubeAuthCode(t *testing.T) {
	ctx := context.Background()

	t.Run("success with subscriptions only", func(t *testing.T) {
		// arrange
		userID := uuid.Must(uuid.NewV7())
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{}),
			new(YouTubeClientMock{
				OAuthAuthCodeURLFunc: func(state string) string {
					return "https://youtube.com/oauth?state=" + state
				},
			}),
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				SignYouTubeImportStateTokenFunc: func(_ uuid.UUID, importSubs, importLikes bool, _ string) (string, error) {
					assert.True(t, importSubs)
					assert.False(t, importLikes)
					return "yt-state-token", nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		authURL, err := svc.CreateYouTubeAuthCode(ctx, userID, true, false)

		// assert
		require.NoError(t, err)
		assert.Contains(t, authURL, "yt-state-token")
	})

	t.Run("no import option returns error", func(t *testing.T) {
		// arrange
		svc := auth.NewService(
			(*pgxpool.Pool)(nil),
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		_, err := svc.CreateYouTubeAuthCode(ctx, uuid.Must(uuid.NewV7()), false, false)

		// assert
		assert.ErrorIs(t, err, auth.ErrNoImportOptionSelected)
	})
}

func TestService_ReactivateAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authPublicID := uuid.Must(uuid.NewV7())
		authResult, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       authPublicID,
		})
		require.NoError(t, err)
		userPublicID := uuid.Must(uuid.NewV7())
		_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
			DisplayName:               "Test User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    3600,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authResult.PublicID,
		})
		require.NoError(t, err)
		require.NoError(t, q.ArchiveUser(ctx, sqlc.ArchiveUserParams{
			LeaveReasonCode: 1,
			UserPublicID:    userPublicID,
		}))

		jti := uuid.Must(uuid.NewV7())
		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
					return false, nil
				},
				InsertJTIFunc: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
					return nil
				},
			}),
		)

		// action
		err = svc.ReactivateAccount(ctx, "register-token")

		// assert
		require.NoError(t, err)
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		// arrange
		tokenErr := errors.New("invalid token")
		svc := auth.NewService(
			testutil.NewTestPool(t),
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return uuid.Nil, uuid.Nil, tokenErr
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{}),
		)

		// action
		err := svc.ReactivateAccount(ctx, "invalid-token")

		// assert
		assert.ErrorIs(t, err, tokenErr)
	})

	t.Run("blacklisted jti returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authPublicID := uuid.Must(uuid.NewV7())
		_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-123",
			LastLoggedInAt: time.Now(),
			PublicID:       authPublicID,
		})
		require.NoError(t, err)

		jti := uuid.Must(uuid.NewV7())

		svc := auth.NewService(
			db,
			new(GoogleClientMock{}),
			nil,
			(*channel.Service)(nil),
			(*playlist.Service)(nil),
			"http://localhost",
			new(ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
			}),
			15*time.Minute, 7*24*time.Hour,
			new(JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, gotJti uuid.UUID) (bool, error) {
					assert.Equal(t, jti, gotJti)
					return true, nil
				},
			}),
		)

		// action
		err = svc.ReactivateAccount(ctx, "register-token")

		// assert
		assert.ErrorIs(t, err, core.ErrJTIBlacklisted)
	})
}
