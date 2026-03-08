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

func NewSession(id uuid.UUID, createdAt, lastLoggedInAt time.Time, countryCode, cityName, browserName string) Session {
	return Session{
		ID:             id,
		CreatedAt:      createdAt,
		LastLoggedInAt: lastLoggedInAt,
		CountryCode:    countryCode,
		CityName:       cityName,
		BrowserName:    browserName,
	}
}
