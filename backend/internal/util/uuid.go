package util

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
)

// UUIDv7MinForTime は指定した時刻のUUID v7の最小値を返す（下位ビットは0埋め）。
// 範囲クエリに使う。
func UUIDv7MinForTime(t time.Time) uuid.UUID {
	var u uuid.UUID

	ms := uint64(t.UnixMilli())

	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	u[6] = 0x70
	u[8] = 0x80

	return u
}

// TimeFromUUIDv7 はUUID v7の先頭48ビット(ミリ秒)から時刻を復元する。
func TimeFromUUIDv7(u uuid.UUID) time.Time {
	ms := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 | int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
	return time.UnixMilli(ms).UTC()
}

// NewUUIDv7WithTime は指定した時刻でUUID v7を生成する。
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

	u[6] = (u[6] & 0x0F) | 0x70 // バージョン7
	u[8] = (u[8] & 0x3F) | 0x80 // バリアント RFC 4122

	return u, nil
}
