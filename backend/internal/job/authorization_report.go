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
	authQS        auth.AuthorizationQueryService
	discordClient discord_d.Client
}

func (j *authorizationReportJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*authorizationReportJob).run")

	count, err := j.authQS.CountAuthorizations(ctx)
	if err != nil {
		return err
	}

	return j.discordClient.SendWebhookMessage(ctx, fmt.Sprintf("**[Daily Report]** `m_user_authorization`\nTotal: **%d**", count))
}

func (j *authorizationReportJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run authorization report job", slog.Any("error", err))
		if wErr := j.discordClient.SendWebhookMessage(ctx, fmt.Sprintf("[Error] authorization report job: %v", err)); wErr != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to send discord webhook", slog.Any("error", wErr))
		}
	}
}

func NewAuthorizationReportJob(db *pgxpool.Pool, discordClient discord_d.Client) scheduler.Job {
	return &authorizationReportJob{
		authQS:        auth.NewAuthorizationQueryService(db),
		discordClient: discordClient,
	}
}
