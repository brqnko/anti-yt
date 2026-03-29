package util

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
)

// NewUUIDv7WithTime generates a UUIDv7 with the specified timestamp.
func NewUUIDv7WithTime(t time.Time) (uuid.UUID, error) {
	var u uuid.UUID
	if _, err := rand.Read(u[:]); err != nil {
		return uuid.Nil, err
	}

	ms := t.UnixMilli()
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	u[6] = (u[6] & 0x0F) | 0x70 // version 7
	u[8] = (u[8] & 0x3F) | 0x80 // variant RFC 4122

	return u, nil
}
