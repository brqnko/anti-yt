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

type OAuthStateClaims struct {
	UserID   string `json:"user_id"`
	Provider string `json:"provider"`
	jwt.RegisteredClaims
}

type Service interface {
	SignUserAccessToken(userID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error)
	SignRegisterToken(authorizationID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error)
	SignOAuthStateToken(userID uuid.UUID, provider, serverURL string) (_ string, err error)
	TokenDuration() time.Duration
	VerifyUserAccessToken(token string) (_, _ uuid.UUID, _ time.Time, err error)
	VerifyRegisterToken(token string) (_, _ uuid.UUID, err error)
	VerifyOAuthStateToken(token string) (_ uuid.UUID, _ string, err error)
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

func (s *serviceImpl) TokenDuration() time.Duration {
	return s.accessTokenDuration
}

func (s *serviceImpl) SignUserAccessToken(userID, jti uuid.UUID, serverURL string) (_ string, _ time.Time, err error) {
	defer util.Wrap(&err, "jwtService.SignUserAccessToken(userID=%s)", userID)

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, UserClaims{
		UserID: base64.URLEncoding.EncodeToString(userID[:]),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "user_access_token",
			Audience:  []string{serverURL},
			ExpiresAt: &jwt.NumericDate{Time: expiresAt},
			NotBefore: &jwt.NumericDate{Time: now},
			IssuedAt:  &jwt.NumericDate{Time: now},
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
	defer util.Wrap(&err, "jwtService.SignRegisterToken(authorizationID=%s)", authorizationID)

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, RegisterClaims{
		AuthorizationID: base64.URLEncoding.EncodeToString(authorizationID[:]),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "authorization_token",
			Audience:  []string{serverURL},
			ExpiresAt: &jwt.NumericDate{Time: expiresAt},
			NotBefore: &jwt.NumericDate{Time: now},
			IssuedAt:  &jwt.NumericDate{Time: now},
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
	defer util.Wrap(&err, "jwtService.VerifyRegisterToken")

	claims := &RegisterClaims{}

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
	defer util.Wrap(&err, "jwtService.VerifyUserAccessToken")

	claims := &UserClaims{}

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

func (s *serviceImpl) SignOAuthStateToken(userID uuid.UUID, provider, serverURL string) (_ string, err error) {
	defer util.Wrap(&err, "jwtService.SignOAuthStateToken(userID=%s, provider=%s)", userID, provider)

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, OAuthStateClaims{
		UserID:   base64.URLEncoding.EncodeToString(userID[:]),
		Provider: provider,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serverURL,
			Subject:   "oauth_state",
			Audience:  []string{serverURL},
			ExpiresAt: &jwt.NumericDate{Time: now.Add(10 * time.Minute)},
			NotBefore: &jwt.NumericDate{Time: now},
			IssuedAt:  &jwt.NumericDate{Time: now},
		},
	})
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (s *serviceImpl) VerifyOAuthStateToken(token string) (_ uuid.UUID, _ string, err error) {
	defer util.Wrap(&err, "jwtService.VerifyOAuthStateToken")

	claims := &OAuthStateClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "oauth_state" {
		return uuid.Nil, "", ErrInvalidToken
	}
	if iss, _ := claims.GetIssuer(); iss != s.serverURL {
		return uuid.Nil, "", ErrInvalidToken
	}
	if aud, _ := claims.GetAudience(); len(aud) != 1 || aud[0] != s.serverURL {
		return uuid.Nil, "", ErrInvalidToken
	}

	decoded, err := base64.URLEncoding.DecodeString(claims.UserID)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, "", ErrInvalidToken
	}
	userID, err := uuid.FromBytes(decoded)
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}

	return userID, claims.Provider, nil
}
