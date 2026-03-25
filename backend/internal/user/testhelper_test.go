package user

import (
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, pool *pgxpool.Pool, jwtMock *testutil.ServiceMock) *Service {
	t.Helper()
	return &Service{
		db:         pool,
		jwtService: jwtMock,
		serverURL:  "https://test.example.com",
		userQS:     NewUserQueryService(pool),
	}
}
