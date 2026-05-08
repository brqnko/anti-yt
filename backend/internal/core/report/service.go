package report

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

type Service struct {
	discordClient discord_d.Client
}

type Option func(*Service)

func WithDiscord(discordClient discord_d.Client) Option {
	return func(s *Service) {
		s.discordClient = discordClient
	}
}

func NewService(opts ...Option) *Service {
	s := new(Service{})
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) ReportError(ctx context.Context, operationID string, err error) {
	if s.discordClient == nil {
		return
	}

	reportCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

	go func() {
		if sendErr := s.discordClient.SendWebhookMessage(reportCtx, fmt.Sprintf("[ERROR] `%s`\n```\n%v\n```", operationID, err)); sendErr != nil {
			util.LoggerFromContext(reportCtx).ErrorContext(reportCtx, "failed to send error report", slog.Any("error", sendErr))
		}
	}()
}
