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

type Authorization struct {
	ID             uuid.UUID
	Issuer         string
	Sub            string
	LastLoggedInAt time.Time
}

type AuthorizationOption func(*Authorization)

func WithLastLoggedInAt(lastLoggedInAt time.Time) AuthorizationOption {
	return func(a *Authorization) {
		a.LastLoggedInAt = lastLoggedInAt
	}
}

func WithID(id uuid.UUID) AuthorizationOption {
	return func(a *Authorization) {
		a.ID = id
	}
}

func NewAuthorization(issuer, sub string, options ...AuthorizationOption) (Authorization, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Authorization{}, err
	}

	authorization := Authorization{
		ID:             id,
		Issuer:         issuer,
		Sub:            sub,
		LastLoggedInAt: time.Now().UTC(),
	}

	for _, opt := range options {
		opt(&authorization)
	}

	return authorization, nil
}
