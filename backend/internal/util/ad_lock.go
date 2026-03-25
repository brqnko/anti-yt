package util

import (
	"context"
	"errors"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
)

func TryAdLock(ctx context.Context, q sqlc.Querier, key []byte) (err error) {
	lockKey := Sha256Int64(key)
	defer Wrap(&err, "TryAdLock(key=%d)", lockKey)

	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, lockKey)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("lock not acquired")
	}

	return nil
}
