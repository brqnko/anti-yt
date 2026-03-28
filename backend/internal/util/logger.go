package util

import (
	"context"
	"log/slog"
)

type slogKey struct{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, slogKey{}, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(slogKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
