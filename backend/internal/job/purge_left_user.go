package job

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

type purgeLeftUserJob struct {
	db             *pgxpool.Pool
	feedRepo       database_d.FeedRepository
	discordService discord_d.Service
}

func (j *purgeLeftUserJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*purgeLeftUserJob).run")

	q := sqlc.New(j.db)

	users, err := q.ListLeftUsers(ctx)
	if err != nil {
		return err
	}

	for _, u := range users {
		if err := q.PurgeLeftUser(ctx, sqlc.PurgeLeftUserParams{
			HUserID:         u.HUserID,
			AuthorizationID: u.MUserAuthorizationID,
		}); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to purge left user", slog.Int64("h_user_id", u.HUserID), slog.Any("error", err))
			continue
		}
		if err := j.feedRepo.DeleteAll(ctx, u.PublicID); err != nil {
			util.LoggerFromContext(ctx).WarnContext(ctx, "failed to delete feed for purged user", slog.String("public_id", u.PublicID.String()), slog.Any("error", err))
		}
	}

	return nil
}

func (j *purgeLeftUserJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run purge left user job", slog.Any("error", err))
		if wErr := j.discordService.SendWebhookMessage(ctx, fmt.Sprintf("[Error] purge left user job: %v", err)); wErr != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to send discord webhook", slog.Any("error", wErr))
		}
	}
}

func NewPurgeLeftUserJob(db *pgxpool.Pool, feedRepo database_d.FeedRepository, discordService discord_d.Service) scheduler.Job {
	return &purgeLeftUserJob{db: db, feedRepo: feedRepo, discordService: discordService}
}
