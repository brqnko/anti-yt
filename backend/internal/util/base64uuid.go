package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Base64UUID uuid.UUID

func (b Base64UUID) encode() string {
	return base64.RawURLEncoding.EncodeToString(b[:])
}

func (b Base64UUID) String() string {
	return b.encode()
}

func (b Base64UUID) UUID() uuid.UUID {
	return uuid.UUID(b)
}

func (b Base64UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.encode())
}

func (b *Base64UUID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return b.parse(s)
}

func (b Base64UUID) MarshalText() ([]byte, error) {
	return []byte(b.encode()), nil
}

func (b *Base64UUID) UnmarshalText(data []byte) error {
	return b.parse(string(data))
}

func (b *Base64UUID) parse(s string) error {
	switch len(s) {
	case 22:
		decoded, err := base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			return fmt.Errorf("invalid base64url UUID: %w", err)
		}
		if len(decoded) != 16 {
			return fmt.Errorf("invalid base64url UUID: decoded length %d, want 16", len(decoded))
		}
		copy(b[:], decoded)
		return nil
	case 36:
		u, err := uuid.Parse(s)
		if err != nil {
			return fmt.Errorf("invalid UUID: %w", err)
		}
		*b = Base64UUID(u)
		return nil
	default:
		return fmt.Errorf("invalid UUID format (length %d): %s", len(s), s)
	}
}
