package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultJWTMock は jwt_d.Service のデフォルトモックを返す
func DefaultJWTMock() *ServiceMock {
	return &ServiceMock{
		SignUserAccessTokenFunc: func(userID, jti uuid.UUID, serverURL string) (string, time.Time, error) {
			return "access-token", time.Now().UTC().Add(30 * time.Minute), nil
		},
		SignRegisterTokenFunc: func(authorizationID, jti uuid.UUID, serverURL string) (string, time.Time, error) {
			return "register-token", time.Now().UTC().Add(30 * time.Minute), nil
		},
		TokenDurationFunc: func() time.Duration {
			return 30 * time.Minute
		},
		VerifyUserAccessTokenFunc: func(token string) (uuid.UUID, uuid.UUID, time.Time, error) {
			return uuid.Nil, uuid.Nil, time.Time{}, nil
		},
		VerifyRegisterTokenFunc: func(token string) (uuid.UUID, uuid.UUID, error) {
			return uuid.Nil, uuid.Nil, nil
		},
	}
}

// SeedAuthorization は m_user_authorization にレコードを挿入し public_id を返す
func SeedAuthorization(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	q := sqlc.New(pool)
	pubID := uuid.Must(uuid.NewV7())
	_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         "https://accounts.google.com",
		Sub:            "test-sub-" + pubID.String(),
		LastLoggedInAt: time.Now().UTC(),
		PublicID:       pubID,
	})
	if err != nil {
		t.Fatalf("SeedAuthorization: %v", err)
	}
	return pubID
}

// SeedUser は authorization + user を作成し user の public_id を返す
func SeedUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	authPubID := SeedAuthorization(t, pool)
	userPubID := uuid.Must(uuid.NewV7())
	q := sqlc.New(pool)
	_, err := q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User",
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    86401,
		JoinedAt:                  time.Now().UTC(),
		PublicID:                  userPubID,
		UserAuthorizationPublicID: authPubID,
	})
	if err != nil {
		t.Fatalf("SeedUser: %v", err)
	}
	return userPubID
}

// SeedUserWithScreenLimit は指定したスクリーン制限でユーザーを作成する。nil は無制限(86401)。
func SeedUserWithScreenLimit(t *testing.T, pool *pgxpool.Pool, screenLimitSeconds *int) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	authPubID := SeedAuthorization(t, pool)
	userPubID := uuid.Must(uuid.NewV7())
	q := sqlc.New(pool)

	dailySeconds := 86401 // unlimited sentinel
	if screenLimitSeconds != nil {
		dailySeconds = *screenLimitSeconds
	}

	_, err := q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User",
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    dailySeconds,
		JoinedAt:                  time.Now().UTC(),
		PublicID:                  userPubID,
		UserAuthorizationPublicID: authPubID,
	})
	if err != nil {
		t.Fatalf("SeedUserWithScreenLimit: %v", err)
	}
	return userPubID
}
