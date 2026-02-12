package jwt_d

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type RegisterClaims struct {
	AuthorizationId string `json:"authorization_id"`
	jwt.RegisteredClaims
}

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTService interface {
	SignUserAccessToken(userID, jti uuid.UUID, serverURL string) (token string, expiresAt time.Time, err error)
	SignRegisterToken(authorizationID, jti uuid.UUID, serverURL string) (token string, expiresAt time.Time, err error)
	TokenDuration() time.Duration
	VerifyUserAccessToken(token string) (userID, jti uuid.UUID, expiresAt time.Time, err error)
	VerifyRegisterToken(token string) (authorizationID uuid.UUID, err error)
}

type jwtService struct {
	publicKey           ed25519.PublicKey
	privateKey          ed25519.PrivateKey
	accessTokenDuration time.Duration
}

func NewJWTService(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, accessTokenDuration time.Duration) JWTService {
	return &jwtService{
		publicKey:           publicKey,
		privateKey:          privateKey,
		accessTokenDuration: accessTokenDuration,
	}
}

func (s *jwtService) TokenDuration() time.Duration {
	return s.accessTokenDuration
}

func (s *jwtService) SignUserAccessToken(userID, jti uuid.UUID, serverURL string) (string, time.Time, error) {
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
	return signed, expiresAt, err
}

func (s *jwtService) SignRegisterToken(authorizationID, jti uuid.UUID, serverURL string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, RegisterClaims{
		AuthorizationId: base64.URLEncoding.EncodeToString(authorizationID[:]),
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
	return signed, expiresAt, err
}

func (s *jwtService) VerifyRegisterToken(token string) (uuid.UUID, error) {
	claims := &RegisterClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "authorization_token" {
		return uuid.Nil, ErrInvalidToken
	}

	decoded, err := base64.URLEncoding.DecodeString(claims.AuthorizationId)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, ErrInvalidToken
	}
	authorizationID := uuid.UUID(decoded)

	return authorizationID, nil
}

func (s *jwtService) VerifyUserAccessToken(token string) (uuid.UUID, uuid.UUID, time.Time, error) {
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
