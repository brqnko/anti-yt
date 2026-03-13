package util

import (
	"context"
	"log/slog"
)

func LogError(ctx context.Context, err error) {
	attrs := []slog.Attr{
		slog.String("error", err.Error()),
	}

	if requestID, ok := RequestIDFromContext(ctx); ok {
		attrs = append(attrs, slog.String("request_id", requestID.String()))
	}
	if userID, ok := UserIDFromContext(ctx); ok {
		attrs = append(attrs, slog.String("user_id", userID.String()))
	}

	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}

	slog.ErrorContext(ctx, "internal server error", args...)
}
