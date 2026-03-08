package util

import (
	"context"

	"github.com/google/uuid"
)

type accessTokenKey struct{}
type refreshTokenKey struct{}
type userIDKey struct{}
type requestIDKey struct{}

func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, token)
}

func AccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(accessTokenKey{}).(string)
	return token, ok
}

func WithRefreshToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, refreshTokenKey{}, token)
}

func RefreshTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(refreshTokenKey{}).(string)
	return token, ok
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return userID, ok
}

func WithRequestID(ctx context.Context, requestID uuid.UUID) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(uuid.UUID)
	return requestID, ok
}
