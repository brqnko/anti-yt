package auth

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	LastLoggedInAt time.Time
	CountryCode    string
	CityName       string
	BrowserName    string
}
