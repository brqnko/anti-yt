package util

import (
	"context"
	"errors"
	"fmt"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
)

func TryAdLock(ctx context.Context, q sqlc.Querier, key int64) error {
	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to tryAcquireAdvisoryXactLock: %w", err)
	}
	if !acquired {
		return errors.New("failed to tryAcquireAdvisoryXactLock: lock not acquired")
	}

	return nil
}
