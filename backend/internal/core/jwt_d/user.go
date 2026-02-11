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

func VerifyUserAccessToken(public ed25519.PublicKey, token string) (userID uuid.UUID, err error) {
	claims := &UserClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return public, nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	if sub, _ := claims.GetSubject(); sub != "user_access_token" {
		return uuid.Nil, ErrInvalidToken
	}

	decoded, err := base64.RawURLEncoding.DecodeString(claims.UserID)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, ErrInvalidToken
	}
	userID, err = uuid.FromBytes(decoded)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func VerifyRegisterToken(public ed25519.PublicKey, token string) (authorizationID uuid.UUID, err error) {
	claims := &RegisterClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return public, nil
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

	decoded, err := base64.RawURLEncoding.DecodeString(claims.AuthorizationId)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, ErrInvalidToken
	}
	authorizationID = uuid.UUID(decoded)

	return authorizationID, nil
}

func VerifyUserAccessTokenWithExpiry(public ed25519.PublicKey, token string) (userID, jti uuid.UUID, expiresAt time.Time, err error) {
	claims := &UserClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, ErrInvalidToken
		}
		return public, nil
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

	decoded, err := base64.RawURLEncoding.DecodeString(claims.UserID)
	if err != nil || len(decoded) != 16 {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}
	userID, err = uuid.FromBytes(decoded)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	jtiDecoded, err := base64.RawURLEncoding.DecodeString(claims.ID)
	if err != nil || len(jtiDecoded) != 16 {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}
	jti, err = uuid.FromBytes(jtiDecoded)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, ErrInvalidToken
	}

	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}

	return userID, jti, expiresAt, nil
}
