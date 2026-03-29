package report

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

type Service struct {
	discord discord_d.Service
}

type Option func(*Service)

func WithDiscordWebhook(webhookURL string) Option {
	return func(s *Service) {
		s.discord = discord_d.NewDiscordClient(webhookURL)
	}
}

func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) ReportError(ctx context.Context, operationID string, err error) {
	if s.discord == nil {
		return
	}

	message := fmt.Sprintf("**[ERROR]** `%s`\n```\n%v\n```", operationID, err)
	if sendErr := s.discord.SendWebhookMessage(ctx, message); sendErr != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to send error report", slog.Any("error", sendErr))
	}
}
