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
	discord discord_d.Service
}

type Option func(*Service)

// WithDiscord は discord_d.Service を直接注入する。
// nilを渡すとDiscord通知が無効化される。
func WithDiscord(svc discord_d.Service) Option {
	return func(s *Service) {
		s.discord = svc
	}
}

func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ReportError はエラーをDiscordに非同期で通知する。
// 呼び出し元のctxがキャンセルされても通知処理は継続するが、
// 10秒でタイムアウトする。
func (s *Service) ReportError(ctx context.Context, operationID string, err error) {
	if s.discord == nil {
		return
	}

	// 元リクエストのctxはレスポンス返却後にキャンセルされるので、
	// 値(ロガー等)は引き継ぎつつキャンセルから切り離す。
	reportCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)

	go func() {
		defer cancel()
		message := fmt.Sprintf("**[ERROR]** `%s`\n```\n%v\n```", operationID, err)
		if sendErr := s.discord.SendWebhookMessage(reportCtx, message); sendErr != nil {
			util.LoggerFromContext(reportCtx).ErrorContext(reportCtx, "failed to send error report", slog.Any("error", sendErr))
		}
	}()
}
