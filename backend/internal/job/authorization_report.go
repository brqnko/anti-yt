package job

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authorizationReportJob struct {
	authQS  auth.AuthorizationQueryService
	discord discord_d.Service
}

func (j *authorizationReportJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*authorizationReportJob).run")

	count, err := j.authQS.CountAuthorizations(ctx)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("**[Daily Report]** `m_user_authorization`\nTotal: **%d**", count)
	return j.discord.SendWebhookMessage(ctx, message)
}

func (j *authorizationReportJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run authorization report job", slog.Any("error", err))
	}
}

func NewAuthorizationReportJob(db *pgxpool.Pool, discord discord_d.Service) scheduler.Job {
	return &authorizationReportJob{
		authQS:  auth.NewAuthorizationQueryService(db),
		discord: discord,
	}
}
