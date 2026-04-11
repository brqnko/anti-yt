package database_d

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RatelimitRepository interface {
	Consume(ctx context.Context, userID uuid.UUID, quota int) (consumed int, err error)
}

type ratelimitRepositoryImpl struct {
	client redis.Cmdable
	window time.Duration
}

func ratelimitKey(userID uuid.UUID) string {
	return "ratelimit:" + userID.String()
}

func (r *ratelimitRepositoryImpl) Consume(ctx context.Context, userID uuid.UUID, quota int) (consumed int, err error) {
	defer util.Wrap(&err, "database_d.(*ratelimitRepositoryImpl).Consume(userID=%s)", userID)

	key := ratelimitKey(userID)
	if quota == 0 {
		v, err := r.client.Get(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(v)
	}

	n, err := r.client.IncrBy(ctx, key, int64(quota)).Result()
	if err != nil {
		return 0, err
	}
	// 初回インクリメント時のみウィンドウのTTLを設定する
	if n == int64(quota) {
		if err := r.client.Expire(ctx, key, r.window).Err(); err != nil {
			return 0, err
		}
	}
	return int(n), nil
}

func NewRatelimitRepository(client redis.Cmdable, window time.Duration) RatelimitRepository {
	return &ratelimitRepositoryImpl{
		client: client,
		window: window,
	}
}

var _ RatelimitRepository = (*ratelimitRepositoryImpl)(nil)
