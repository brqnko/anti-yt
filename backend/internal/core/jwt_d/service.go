package jwt_d

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type RegisterClaims struct {
	AuthorizationID string `json:"authorization_id"`
	jwt.RegisteredClaims
}

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type YouTubeImportStateClaims struct {
	UserID              string `json:"user_id"`
	ImportSubscriptions bool   `json:"import_subscriptions"`
	ImportLikes         bool   `json:"import_likes"`
	jwt.RegisteredClaims
}

type OIDCStateClaims struct {
	Platform string `json:"platform"`
	jwt.RegisteredClaims
}

type Service interface {
	SignUserAccessToken(userID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error)
	SignRegisterToken(authorizationID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error)
	SignYouTubeImportStateToken(userID uuid.UUID, importSubscriptions, importLikes bool, serverURL string) (_ string, err error)
	SignOIDCStateToken(platform, serverURL string) (_ string, err error)
	VerifyUserAccessToken(token string) (_, _ uuid.UUID, _ time.Time, err error)
	VerifyRegisterToken(token string) (_, _ uuid.UUID, err error)
	VerifyYouTubeImportStateToken(token string) (_ uuid.UUID, _ bool, _ bool, err error)
	VerifyOIDCStateToken(token string) (_ string, err error)
}

type serviceImpl struct {
	publicKey           ed25519.PublicKey
	privateKey          ed25519.PrivateKey
	accessTokenDuration time.Duration
	serverURL           string
}

func NewService(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, accessTokenDuration time.Duration, serverURL string) Service {
	return &serviceImpl{
		publicKey:           publicKey,
		privateKey:          privateKey,
		accessTokenDuration: accessTokenDuration,
		serverURL:           serverURL,
	}
}

func (s *serviceImpl) SignUserAccessToken(userID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).SignUserAccessToken(userID=%s)", userID)

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, UserClaims{
		UserID: base64.URLEncoding.EncodeToString(userID[:]),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "user_access_token",
			Audience:  []string{serverURL},
			ExpiresAt: new(jwt.NumericDate{Time: expiresAt}),
			NotBefore: new(jwt.NumericDate{Time: now}),
			IssuedAt:  new(jwt.NumericDate{Time: now}),
			ID:        base64.URLEncoding.EncodeToString(jti[:]),
		},
	})
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", expiresAt, err
	}
	return signed, expiresAt, nil
}

func (s *serviceImpl) SignRegisterToken(authorizationID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).SignRegisterToken(authorizationID=%s)", authorizationID)

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, RegisterClaims{
		AuthorizationID: base64.URLEncoding.EncodeToString(authorizationID[:]),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "authorization_token",
			Audience:  []string{serverURL},
			ExpiresAt: new(jwt.NumericDate{Time: expiresAt}),
			NotBefore: new(jwt.NumericDate{Time: now}),
			IssuedAt:  new(jwt.NumericDate{Time: now}),
			ID:        base64.URLEncoding.EncodeToString(jti[:]),
		},
	})
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", expiresAt, err
	}
	return signed, expiresAt, nil
}

func (s *serviceImpl) VerifyRegisterToken(token string) (_ uuid.UUID, _ uuid.UUID, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).VerifyRegisterToken")

	claims := new(RegisterClaims{})

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "authorization_token" {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}

	decoded, err := base64.URLEncoding.DecodeString(claims.AuthorizationID)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	authorizationID := uuid.UUID(decoded)

	jtiDecoded, err := base64.URLEncoding.DecodeString(claims.ID)
	if err != nil || len(jtiDecoded) != 16 {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	jti, err := uuid.FromBytes(jtiDecoded)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}

	return authorizationID, jti, nil
}

func (s *serviceImpl) VerifyUserAccessToken(token string) (_ uuid.UUID, _ uuid.UUID, _ time.Time, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).VerifyUserAccessToken")

	claims := new(UserClaims{})

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "user_access_token" {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	decoded, err := base64.URLEncoding.DecodeString(claims.UserID)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}
	userID, err := uuid.FromBytes(decoded)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	jtiDecoded, err := base64.URLEncoding.DecodeString(claims.ID)
	if err != nil || len(jtiDecoded) != 16 {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}
	jti, err := uuid.FromBytes(jtiDecoded)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}

	return userID, jti, expiresAt, nil
}

func (s *serviceImpl) SignOIDCStateToken(platform, serverURL string) (_ string, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).SignOIDCStateToken(platform=%s)", platform)

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, OIDCStateClaims{
		Platform: platform,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "oidc_state",
			Audience:  []string{serverURL},
			ExpiresAt: new(jwt.NumericDate{Time: now.Add(10 * time.Minute)}),
			NotBefore: new(jwt.NumericDate{Time: now}),
			IssuedAt:  new(jwt.NumericDate{Time: now}),
		},
	})
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (s *serviceImpl) VerifyOIDCStateToken(token string) (_ string, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).VerifyOIDCStateToken")

	claims := new(OIDCStateClaims{})
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return "", ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "oidc_state" {
		return "", ErrInvalidToken
	}
	if iss, _ := claims.GetIssuer(); iss != s.serverURL {
		return "", ErrInvalidToken
	}
	if aud, _ := claims.GetAudience(); len(aud) != 1 || aud[0] != s.serverURL {
		return "", ErrInvalidToken
	}

	return claims.Platform, nil
}

func (s *serviceImpl) SignYouTubeImportStateToken(userID uuid.UUID, importSubscriptions, importLikes bool, serverURL string) (_ string, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).SignYouTubeImportStateToken(userID=%s)", userID)

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, YouTubeImportStateClaims{
		UserID:              base64.URLEncoding.EncodeToString(userID[:]),
		ImportSubscriptions: importSubscriptions,
		ImportLikes:         importLikes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "youtube_import_state",
			Audience:  []string{serverURL},
			ExpiresAt: new(jwt.NumericDate{Time: now.Add(10 * time.Minute)}),
			NotBefore: new(jwt.NumericDate{Time: now}),
			IssuedAt:  new(jwt.NumericDate{Time: now}),
		},
	})
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (s *serviceImpl) VerifyYouTubeImportStateToken(token string) (_ uuid.UUID, _ bool, _ bool, err error) {
	defer util.Wrap(&err, "jwt_d.(*serviceImpl).VerifyYouTubeImportStateToken")

	claims := new(YouTubeImportStateClaims{})
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return uuid.Nil, false, false, ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "youtube_import_state" {
		return uuid.Nil, false, false, ErrInvalidToken
	}
	if iss, _ := claims.GetIssuer(); iss != s.serverURL {
		return uuid.Nil, false, false, ErrInvalidToken
	}
	if aud, _ := claims.GetAudience(); len(aud) != 1 || aud[0] != s.serverURL {
		return uuid.Nil, false, false, ErrInvalidToken
	}

	decoded, err := base64.URLEncoding.DecodeString(claims.UserID)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, false, false, ErrInvalidToken
	}
	userID, err := uuid.FromBytes(decoded)
	if err != nil {
		return uuid.Nil, false, false, ErrInvalidToken
	}

	return userID, claims.ImportSubscriptions, claims.ImportLikes, nil
}
