package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, db *pgxpool.Pool, oidcMock *GoogleOIDCServiceMock, jwtMock *testutil.ServiceMock) *Service {
	t.Helper()
	s := &Service{
		db:                   db,
		oidcService:          oidcMock,
		jwtService:           jwtMock,
		serverURL:            "https://test.example.com",
		refreshTokenDuration: 30 * 24 * time.Hour,
	}
	if db != nil {
		s.refreshTokenQS = NewRefreshTokenQueryService(db)
		s.authorizationQS = NewAuthorizationQueryService(db)
		s.userQS = user.NewUserQueryService(db)
	}
	return s
}

func defaultOIDCMock(sub string) *GoogleOIDCServiceMock {
	return &GoogleOIDCServiceMock{
		AuthCodeURLFunc: func(state string) string {
			return fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?state=%s", state)
		},
		ExchangeAndVerifyFunc: func(ctx context.Context, code string) (string, error) {
			return sub, nil
		},
	}
}
