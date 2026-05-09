package testutil

import (
	"math"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func NewFeedRepo(t *testing.T, q sqlc.Querier) database_d.FeedRepository {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(new(redis.Options{Addr: mr.Addr()}))
	t.Cleanup(func() { _ = client.Close() })

	return database_d.NewFeedRepository(client, math.MaxInt64, q)
}
