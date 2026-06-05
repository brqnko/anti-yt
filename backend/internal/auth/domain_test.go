package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewRefreshToken(t *testing.T) {
	t.Parallel()

	type arg struct {
		userAgent   string
		ipAddress   string
		countryCode string
		cityName    string
		expiresAt   time.Time
		opts        []auth.RefreshTokenOption
	}

	type want struct {
		ipAddress   string
		userAgent   string
		countryCode string
		cityName    string
		browserName string
		deviceType  string
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg: arg{
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				ipAddress:   "docker",
				countryCode: "jp",
				cityName:    "docker",
				expiresAt:   time.Now(),
				opts:        []auth.RefreshTokenOption{auth.WithRefreshTokenRaw("raw")},
			},
			want: &want{
				ipAddress:   "docker",
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				countryCode: "jp",
				cityName:    "docker",
				browserName: "Chrome:122.0.0.0",
				deviceType:  "Windows 10",
			},
			wantErr: nil,
		},
		"ipAddress truncated": {
			arg: arg{
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				ipAddress:   strings.Repeat("あ", 65),
				countryCode: "jp",
				cityName:    "docker",
				expiresAt:   time.Now(),
				opts:        []auth.RefreshTokenOption{auth.WithRefreshTokenRaw("raw")},
			},
			want: &want{
				ipAddress:   strings.Repeat("あ", 64),
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				countryCode: "jp",
				cityName:    "docker",
				browserName: "Chrome:122.0.0.0",
				deviceType:  "Windows 10",
			},
			wantErr: nil,
		},
		"userAgent truncated": {
			arg: arg{
				userAgent:   strings.Repeat("あ", 513),
				ipAddress:   "docker",
				countryCode: "jp",
				cityName:    "docker",
				expiresAt:   time.Now(),
				opts:        []auth.RefreshTokenOption{auth.WithRefreshTokenRaw("raw")},
			},
			want: &want{
				ipAddress:   "docker",
				userAgent:   strings.Repeat("あ", 512),
				countryCode: "jp",
				cityName:    "docker",
				browserName: strings.Repeat("あ", 64),
				deviceType:  "",
			},
			wantErr: nil,
		},
		"countryCode truncated": {
			arg: arg{
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				ipAddress:   "docker",
				countryCode: "jpn",
				cityName:    "docker",
				expiresAt:   time.Now(),
				opts:        []auth.RefreshTokenOption{auth.WithRefreshTokenRaw("raw")},
			},
			want: &want{
				ipAddress:   "docker",
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				countryCode: "jp",
				cityName:    "docker",
				browserName: "Chrome:122.0.0.0",
				deviceType:  "Windows 10",
			},
			wantErr: nil,
		},
		"cityName truncated": {
			arg: arg{
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				ipAddress:   "docker",
				countryCode: "jp",
				cityName:    strings.Repeat("あ", 129),
				expiresAt:   time.Now(),
				opts:        []auth.RefreshTokenOption{auth.WithRefreshTokenRaw("raw")},
			},
			want: &want{
				ipAddress:   "docker",
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				countryCode: "jp",
				cityName:    strings.Repeat("あ", 128),
				browserName: "Chrome:122.0.0.0",
				deviceType:  "Windows 10",
			},
			wantErr: nil,
		},
		"token hash not set": {
			arg: arg{
				userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
				ipAddress:   "docker",
				countryCode: "jp",
				cityName:    "docker",
				expiresAt:   time.Now(),
				opts:        []auth.RefreshTokenOption{},
			},
			want:    nil,
			wantErr: auth.ErrRefreshTokenHashNotSet,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange

			// act
			got, err := auth.NewRefreshToken(
				c.arg.userAgent,
				c.arg.ipAddress,
				c.arg.countryCode,
				c.arg.cityName,
				c.arg.expiresAt,
				c.arg.opts...,
			)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.ipAddress, got.IpAddress)
				assert.Equal(t, c.want.userAgent, got.UserAgent)
				assert.Equal(t, c.want.countryCode, got.CountryCode)
				assert.Equal(t, c.want.cityName, got.CityName)
				assert.Equal(t, c.want.browserName, got.BrowserName)
				assert.Equal(t, c.want.deviceType, got.DeviceType)
			}
		})
	}
}

func TestNewAuthorization(t *testing.T) {
	t.Parallel()

	type arg struct {
		issuer string
		sub    string
		opts   []auth.AuthorizationOption
	}

	type want struct {
		issuer string
		sub    string
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg: arg{
				issuer: "4qtgw5rehtyrj",
				sub:    "24gh356j4",
				opts:   []auth.AuthorizationOption{},
			},
			want: &want{
				issuer: "4qtgw5rehtyrj",
				sub:    "24gh356j4",
			},
			wantErr: nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange

			// act
			got, err := auth.NewAuthorization(c.arg.issuer, c.arg.sub, c.arg.opts...)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.arg.issuer, got.Issuer)
				assert.Equal(t, c.arg.sub, got.Sub)
			}
		})
	}
}
