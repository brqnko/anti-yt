package middleware_d

import (
	"context"
	"net/http"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

// stubHandler は呼ばれたらフラグを立てるだけのハンドラ
func stubHandler(called *bool) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		*called = true
		return nil, nil
	}
}

// stubHandlerWithError は指定エラーを返すハンドラ
func stubHandlerWithError(called *bool, err error) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		*called = true
		return nil, err
	}
}
