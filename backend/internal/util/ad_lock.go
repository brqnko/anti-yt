package util

import (
	"context"
	"errors"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
)

func TryAdLock(ctx context.Context, q sqlc.Querier, key int64) (err error) {
	defer Wrap(&err, "TryAdLock(key=%d)", key)

	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, key)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("lock not acquired")
	}

	return nil
}
