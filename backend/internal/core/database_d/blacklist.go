package database_d

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type JtiBlacklistRepository interface {
	IsJtiExist(ctx context.Context, jti uuid.UUID) (found bool, err error)
	InsertJTI(ctx context.Context, jti uuid.UUID, expiresAt time.Time) (err error)
}

type jtiBlacklistRepositoryImpl struct {
	client redis.Cmdable
}

func jtiBlacklistKey(jti uuid.UUID) string {
	return "jti:" + jti.String()
}

func (j *jtiBlacklistRepositoryImpl) InsertJTI(ctx context.Context, jti uuid.UUID, expiresAt time.Time) (err error) {
	defer util.Wrap(&err, "database_d.(*jtiBlacklistRepositoryImpl).InsertJTI(jti=%s)", jti)

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return j.client.Set(ctx, jtiBlacklistKey(jti), 1, ttl).Err()
}

func (j *jtiBlacklistRepositoryImpl) IsJtiExist(ctx context.Context, jti uuid.UUID) (found bool, err error) {
	defer util.Wrap(&err, "database_d.(*jtiBlacklistRepositoryImpl).IsJtiExist(jti=%s)", jti)

	n, err := j.client.Exists(ctx, jtiBlacklistKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func NewJtiBlacklistRepository(client redis.Cmdable) JtiBlacklistRepository {
	return &jtiBlacklistRepositoryImpl{
		client: client,
	}
}

var _ JtiBlacklistRepository = (*jtiBlacklistRepositoryImpl)(nil)
