package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUser(t *testing.T, ctx context.Context, q sqlc.Querier) (authPublicID uuid.UUID, userPublicID uuid.UUID) {
	t.Helper()
	authPublicID = uuid.Must(uuid.NewV7())
	_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         "https://accounts.google.com",
		Sub:            "google-sub-" + uuid.Must(uuid.NewV7()).String(),
		LastLoggedInAt: time.Now(),
		PublicID:       authPublicID,
	})
	require.NoError(t, err)
	userPublicID = uuid.Must(uuid.NewV7())
	_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User",
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    3600,
		JoinedAt:                  time.Now(),
		PublicID:                  userPublicID,
		UserAuthorizationPublicID: authPublicID,
	})
	require.NoError(t, err)
	return authPublicID, userPublicID
}

func TestService_CreateNewUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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
		svc := user.NewService(
			db,
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
				SignUserAccessTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "access-token", time.Now().Add(time.Hour), nil
				},
				TokenDurationFunc: func() time.Duration { return 15 * time.Minute },
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
					return false, nil
				},
				InsertJTIFunc: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
					return nil
				},
			},
		)

		screenLimit := 3600
		// action
		u, accessToken, expiresAt, err := svc.CreateNewUser(
			ctx,
			"register-token",
			&screenLimit,
			nil,
			"Test User",
			"ja",
			time.UTC,
		)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "Test User", u.DisplayName.String())
		assert.Equal(t, "ja", u.LanguageCode.String())
		assert.Equal(t, "access-token", accessToken)
		assert.False(t, expiresAt.IsZero())
	})

	t.Run("success with screen time ranges", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authPublicID := uuid.Must(uuid.NewV7())
		_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-456",
			LastLoggedInAt: time.Now(),
			PublicID:       authPublicID,
		})
		require.NoError(t, err)

		jti := uuid.Must(uuid.NewV7())
		svc := user.NewService(
			db,
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
				SignUserAccessTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "access-token", time.Now().Add(time.Hour), nil
				},
				TokenDurationFunc: func() time.Duration { return 15 * time.Minute },
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
					return false, nil
				},
				InsertJTIFunc: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
					return nil
				},
			},
		)

		screenLimit := 7200
		ranges := []struct{ Start, End int }{
			{Start: 3600, End: 7200},
			{Start: 36000, End: 72000},
		}

		// action
		u, _, _, err := svc.CreateNewUser(
			ctx, "register-token", &screenLimit, ranges, "Ranged User", "en", time.UTC,
		)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "Ranged User", u.DisplayName.String())
		assert.Equal(t, "en", u.LanguageCode.String())
	})

	t.Run("unlimited screen time", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authPublicID := uuid.Must(uuid.NewV7())
		_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-unlimited",
			LastLoggedInAt: time.Now(),
			PublicID:       authPublicID,
		})
		require.NoError(t, err)

		jti := uuid.Must(uuid.NewV7())
		svc := user.NewService(
			db,
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
				SignUserAccessTokenFunc: func(_, _ uuid.UUID, _ string) (string, time.Time, error) {
					return "access-token", time.Now().Add(time.Hour), nil
				},
				TokenDurationFunc: func() time.Duration { return 15 * time.Minute },
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
					return false, nil
				},
				InsertJTIFunc: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
					return nil
				},
			},
		)

		// action
		u, _, _, err := svc.CreateNewUser(
			ctx, "register-token", nil, nil, "Unlimited User", "ja", time.UTC,
		)

		// assert
		require.NoError(t, err)
		assert.True(t, u.ScreenTimeLimit.IsUnlimited())
	})

	t.Run("invalid register token returns error", func(t *testing.T) {
		// arrange
		tokenErr := errors.New("invalid token")
		svc := user.NewService(
			testutil.NewTestPool(t),
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return uuid.Nil, uuid.Nil, tokenErr
				},
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{},
		)

		// action
		_, _, _, err := svc.CreateNewUser(
			ctx, "invalid-token", nil, nil, "Test User", "ja", time.UTC,
		)

		// assert
		assert.ErrorIs(t, err, tokenErr)
	})

	t.Run("invalid display name returns error", func(t *testing.T) {
		// arrange
		svc := user.NewService(
			testutil.NewTestPool(t),
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil
				},
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{},
		)

		// action
		_, _, _, err := svc.CreateNewUser(
			ctx, "register-token", nil, nil, "", "ja", time.UTC,
		)

		// assert
		assert.ErrorIs(t, err, user.ErrDisplayNameTooShort)
	})

	t.Run("invalid language code returns error", func(t *testing.T) {
		// arrange
		svc := user.NewService(
			testutil.NewTestPool(t),
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil
				},
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{},
		)

		// action
		_, _, _, err := svc.CreateNewUser(
			ctx, "register-token", nil, nil, "Test User", "xx", time.UTC,
		)

		// assert
		assert.ErrorIs(t, err, user.ErrLanguageCodeNotSupported)
	})

	t.Run("already registered authorization returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authPublicID, _ := setupUser(t, ctx, q)

		jti := uuid.Must(uuid.NewV7())
		svc := user.NewService(
			db,
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
					return false, nil
				},
			},
		)

		// action
		_, _, _, err := svc.CreateNewUser(
			ctx, "register-token", nil, nil, "Duplicate User", "ja", time.UTC,
		)

		// assert
		assert.ErrorIs(t, err, user.ErrInvalidAuthorizationIDRegistered)
	})

	t.Run("blacklisted jti returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		authPublicID := uuid.Must(uuid.NewV7())
		_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
			Issuer:         "https://accounts.google.com",
			Sub:            "google-sub-blacklisted",
			LastLoggedInAt: time.Now(),
			PublicID:       authPublicID,
		})
		require.NoError(t, err)

		jti := uuid.Must(uuid.NewV7())

		svc := user.NewService(
			db,
			&ServiceMock{
				VerifyRegisterTokenFunc: func(_ string) (uuid.UUID, uuid.UUID, error) {
					return authPublicID, jti, nil
				},
			},
			"http://localhost",
			&JtiBlacklistRepositoryMock{
				IsJtiExistFunc: func(_ context.Context, gotJti uuid.UUID) (bool, error) {
					assert.Equal(t, jti, gotJti)
					return true, nil
				},
			},
		)

		// action
		_, _, _, err = svc.CreateNewUser(
			ctx, "register-token", nil, nil, "Test User", "ja", time.UTC,
		)

		// assert
		assert.ErrorIs(t, err, core.ErrJTIBlacklisted)
	})
}

func TestService_EditUser(t *testing.T) {
	ctx := context.Background()

	t.Run("update display name", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		newName := "Updated Name"

		// action
		u, err := svc.EditUser(ctx, userPublicID, &newName, nil, nil, nil, time.UTC)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", u.DisplayName.String())
	})

	t.Run("update language code", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		newLang := "en"

		// action
		u, err := svc.EditUser(ctx, userPublicID, nil, &newLang, nil, nil, time.UTC)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "en", u.LanguageCode.String())
	})

	t.Run("update screen time limit", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		newLimit := 7200

		// action
		u, err := svc.EditUser(ctx, userPublicID, nil, nil, &newLimit, nil, time.UTC)

		// assert
		require.NoError(t, err)
		assert.Equal(t, 7200, u.ScreenTimeLimit.Seconds())
	})

	t.Run("update screen time ranges", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		newRanges := []struct{ Start, End int }{
			{Start: 0, End: 3600},
			{Start: 7200, End: 10800},
		}

		// action
		_, err := svc.EditUser(ctx, userPublicID, nil, nil, nil, &newRanges, time.UTC)

		// assert
		require.NoError(t, err)
	})

	t.Run("update all fields", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		newName := "New Name"
		newLang := "en"
		newLimit := 1800
		newRanges := []struct{ Start, End int }{
			{Start: 0, End: 3600},
		}

		// action
		u, err := svc.EditUser(ctx, userPublicID, &newName, &newLang, &newLimit, &newRanges, time.UTC)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "New Name", u.DisplayName.String())
		assert.Equal(t, "en", u.LanguageCode.String())
		assert.Equal(t, 1800, u.ScreenTimeLimit.Seconds())
	})

	t.Run("nonexistent user returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		newName := "Updated Name"

		// action
		_, err := svc.EditUser(ctx, uuid.Must(uuid.NewV7()), &newName, nil, nil, nil, time.UTC)

		// assert
		assert.ErrorIs(t, err, core.ErrNotFound)
	})

	t.Run("invalid display name returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		emptyName := ""

		// action
		_, err := svc.EditUser(ctx, userPublicID, &emptyName, nil, nil, nil, time.UTC)

		// assert
		assert.ErrorIs(t, err, user.ErrDisplayNameTooShort)
	})

	t.Run("invalid language code returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		invalidLang := "xx"

		// action
		_, err := svc.EditUser(ctx, userPublicID, nil, &invalidLang, nil, nil, time.UTC)

		// assert
		assert.ErrorIs(t, err, user.ErrLanguageCodeNotSupported)
	})
}

func TestService_GetUserStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		// action
		view, err := svc.GetUserStatus(ctx, userPublicID)

		// assert
		require.NoError(t, err)
		assert.Equal(t, userPublicID, view.UserID)
		assert.Equal(t, "Test User", view.DisplayName)
		assert.Equal(t, "ja", view.LanguageCode)
		assert.NotNil(t, view.DailyScreenSeconds)
		assert.Equal(t, 3600, *view.DailyScreenSeconds)
	})

	t.Run("nonexistent user returns error", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		// action
		_, err := svc.GetUserStatus(ctx, uuid.Must(uuid.NewV7()))

		// assert
		assert.ErrorIs(t, err, core.ErrNotFound)
	})
}

func TestService_RemoveUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		_, userPublicID := setupUser(t, ctx, q)

		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		// action
		err := svc.RemoveUser(ctx, userPublicID)

		// assert
		require.NoError(t, err)

		// verify user is gone from GetUserStatus
		_, err = svc.GetUserStatus(ctx, userPublicID)
		assert.ErrorIs(t, err, core.ErrNotFound)
	})

	t.Run("nonexistent user succeeds silently", func(t *testing.T) {
		// arrange
		db := testutil.NewTestPool(t)
		svc := user.NewService(db, &ServiceMock{}, "http://localhost", &JtiBlacklistRepositoryMock{})

		// action
		err := svc.RemoveUser(ctx, uuid.Must(uuid.NewV7()))

		// assert
		assert.NoError(t, err)
	})
}
