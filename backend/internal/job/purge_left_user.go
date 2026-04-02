package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

type purgeLeftUserJob struct {
	db *pgxpool.Pool
}

func (j *purgeLeftUserJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "purgeLeftUserJob.run")

	q := sqlc.New(j.db)

	users, err := q.ListLeftUsers(ctx)
	if err != nil {
		return err
	}

	for _, u := range users {
		if err := q.PurgeLeftUser(ctx, sqlc.PurgeLeftUserParams{
			HUserID:         u.HUserID,
			AuthorizationID: u.MUserAuthorizationID,
			UserPublicID:    u.PublicID,
		}); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to purge left user", slog.Int64("h_user_id", u.HUserID), slog.Any("error", err))
			continue
		}
	}

	return nil
}

func (j *purgeLeftUserJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run purge left user job", slog.Any("error", err))
	}
}

func NewPurgeLeftUserJob(db *pgxpool.Pool) scheduler.Job {
	return &purgeLeftUserJob{db: db}
}
