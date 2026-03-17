package util

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrUserIDNotFoundInContext = errors.New("user id not found in context")

type accessTokenKey struct{}
type refreshTokenKey struct{}
type userIDKey struct{}
type requestIDKey struct{}
type requestPathKey struct{}

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

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrUserIDNotFoundInContext
	}
	return userID, nil
}

func WithRequestID(ctx context.Context, requestID uuid.UUID) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(uuid.UUID)
	return requestID, ok
}

func WithRequestPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, requestPathKey{}, path)
}

func RequestPathFromContext(ctx context.Context) (string, bool) {
	path, ok := ctx.Value(requestPathKey{}).(string)
	return path, ok
}

type responseCookiesKey struct{}

func WithResponseCookies(ctx context.Context) context.Context {
	return context.WithValue(ctx, responseCookiesKey{}, &[]string{})
}

func AddResponseCookie(ctx context.Context, cookie string) {
	cookies, ok := ctx.Value(responseCookiesKey{}).(*[]string)
	if !ok {
		return
	}
	*cookies = append(*cookies, cookie)
}

func ResponseCookiesFromContext(ctx context.Context) []string {
	cookies, ok := ctx.Value(responseCookiesKey{}).(*[]string)
	if !ok {
		return nil
	}
	return *cookies
}
