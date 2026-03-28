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
	defer util.Wrap(&err, "TryAdLock(key=%d)", lockKey)

	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, lockKey)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("lock not acquired")
	}

	return nil
}
