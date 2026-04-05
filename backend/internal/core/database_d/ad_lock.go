package database_d

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func TryAdLock(ctx context.Context, q sqlc.Querier, key []byte) (err error) {
	hash := sha256.Sum256(key)
	lockKey := int64(binary.BigEndian.Uint64(hash[:8]))
	defer util.Wrap(&err, "database_d.TryAdLock(key=%d)", lockKey)

	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, lockKey)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("lock not acquired")
	}

	return nil
}

func TryAdLockSession(ctx context.Context, q sqlc.Querier, key []byte) (err error) {
	hash := sha256.Sum256(key)
	lockKey := int64(binary.BigEndian.Uint64(hash[:8]))
	defer util.Wrap(&err, "database_d.TryAdLockSession(key=%d)", lockKey)

	acquired, err := q.TryAcquireAdvisoryLock(ctx, lockKey)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("lock not acquired")
	}

	return nil
}

func ReleaseAdLock(ctx context.Context, q sqlc.Querier, key []byte) (err error) {
	hash := sha256.Sum256(key)
	lockKey := int64(binary.BigEndian.Uint64(hash[:8]))
	defer util.Wrap(&err, "database_d.ReleaseAdLock(key=%d)", lockKey)

	released, err := q.ReleaseAdvisoryLock(ctx, lockKey)
	if err != nil {
		return err
	}
	if !released {
		return errors.New("lock not released")
	}

	return nil
}
