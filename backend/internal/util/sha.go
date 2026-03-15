package util

import (
	"crypto/sha256"
	"encoding/binary"
)

func Sha256Int64(bytes []byte) int64 {
	hash := sha256.Sum256(bytes)
	return int64(binary.BigEndian.Uint64(hash[:8]))
}
