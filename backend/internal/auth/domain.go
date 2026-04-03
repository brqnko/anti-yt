package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
)

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

type RefreshToken struct {
	ID                uuid.UUID
	ActivatedAt       time.Time
	TokenHash         string
	IpAddress         string
	DeviceFingerprint string
	UserAgent         string
	CountryCode       string
	CityName          string
	BrowserName       string
	DeviceType        string
	ExpiresAt         time.Time
	AccessTokenJTI    uuid.UUID
	LastLoggedInAt    time.Time
}

type RefreshTokenOption func(*RefreshToken)

func WithRefreshTokenID(id uuid.UUID) RefreshTokenOption {
	return func(rt *RefreshToken) {
		rt.ID = id
	}
}

func WithRefreshTokenActivatedAt(activatedAt time.Time) RefreshTokenOption {
	return func(rt *RefreshToken) {
		rt.ActivatedAt = activatedAt
	}
}

func WithRefreshTokenHash(tokenHash string) RefreshTokenOption {
	return func(rt *RefreshToken) {
		rt.TokenHash = tokenHash
	}
}

func WithRefreshTokenRaw(tokenRaw string) RefreshTokenOption {
	return func(rt *RefreshToken) {
		rt.TokenHash = util.Sha256Hex(tokenRaw)
	}
}

func WithRefreshTokenLastLoggedInAt(lastLoggedInAt time.Time) RefreshTokenOption {
	return func(rt *RefreshToken) {
		rt.LastLoggedInAt = lastLoggedInAt
	}
}

func NewRefreshToken(userAgent, deviceFingerprint, ipAddress, countryCode, cityName string, expiresAt time.Time, opts ...RefreshTokenOption) (_ *RefreshToken, err error) {
	defer util.Wrap(&err, "auth.NewRefreshToken")

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	accessTokenJTI, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	ua := user_agent.New(userAgent)
	browserName, browserVersion := ua.Browser()

	rt := &RefreshToken{
		ID:                id,
		ActivatedAt:       now,
		TokenHash:         "",
		IpAddress:         truncate(ipAddress, 64),
		DeviceFingerprint: truncate(deviceFingerprint, 32),
		UserAgent:         truncate(userAgent, 512),
		CountryCode:       truncate(countryCode, 2),
		CityName:          truncate(cityName, 128),
		BrowserName:       truncate(fmt.Sprintf("%s:%s", browserName, browserVersion), 64),
		DeviceType:        truncate(ua.OSInfo().FullName, 32),
		ExpiresAt:         expiresAt,
		AccessTokenJTI:    accessTokenJTI,
		LastLoggedInAt:    now,
	}
	for _, opt := range opts {
		opt(rt)
	}

	if rt.TokenHash == "" {
		return nil, errors.New("refresh token token hash is not set")
	}

	return rt, nil
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

func WithAuthorizationID(id uuid.UUID) AuthorizationOption {
	return func(a *Authorization) {
		a.ID = id
	}
}

func NewAuthorization(issuer, sub string, options ...AuthorizationOption) (_ *Authorization, err error) {
	defer util.Wrap(&err, "auth.NewAuthorization")

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	authorization := &Authorization{
		ID:             id,
		Issuer:         issuer,
		Sub:            sub,
		LastLoggedInAt: time.Now().UTC(),
	}

	for _, opt := range options {
		opt(authorization)
	}

	return authorization, nil
}
