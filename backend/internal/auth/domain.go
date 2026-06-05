package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
)

var ErrRefreshTokenHashNotSet = errors.New("refresh token token hash is not set")

type RefreshToken struct {
	ID             uuid.UUID
	ActivatedAt    time.Time
	TokenHash      string
	IpAddress      string
	UserAgent      string
	CountryCode    string
	CityName       string
	BrowserName    string
	DeviceType     string
	ExpiresAt      time.Time
	AccessTokenJTI uuid.UUID
	LastLoggedInAt time.Time
}

type RefreshTokenOption func(*RefreshToken)

func WithRefreshTokenRaw(tokenRaw string) RefreshTokenOption {
	return func(rt *RefreshToken) {
		rt.TokenHash = util.Sha256Hex(tokenRaw)
	}
}

func NewRefreshToken(
	userAgent,
	ipAddress,
	countryCode,
	cityName string,
	expiresAt time.Time,
	opts ...RefreshTokenOption,
) (_ *RefreshToken, err error) {
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

	rt := new(RefreshToken{
		ID:             id,
		ActivatedAt:    now,
		TokenHash:      "",
		IpAddress:      util.Truncate(ipAddress, 64),
		UserAgent:      util.Truncate(userAgent, 512),
		CountryCode:    util.Truncate(countryCode, 2),
		CityName:       util.Truncate(cityName, 128),
		BrowserName:    util.Truncate(fmt.Sprintf("%s:%s", browserName, browserVersion), 64),
		DeviceType:     util.Truncate(ua.OSInfo().FullName, 32),
		ExpiresAt:      expiresAt,
		AccessTokenJTI: accessTokenJTI,
		LastLoggedInAt: now,
	})
	for _, opt := range opts {
		opt(rt)
	}

	if rt.TokenHash == "" {
		return nil, ErrRefreshTokenHashNotSet
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

func NewAuthorization(issuer, sub string, opts ...AuthorizationOption) (_ *Authorization, err error) {
	defer util.Wrap(&err, "auth.NewAuthorization")

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	authorization := new(Authorization{
		ID:             id,
		Issuer:         issuer,
		Sub:            sub,
		LastLoggedInAt: time.Now().UTC(),
	})

	for _, opt := range opts {
		opt(authorization)
	}

	return authorization, nil
}
