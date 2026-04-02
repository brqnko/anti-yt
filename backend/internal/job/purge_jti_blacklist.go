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

type purgeJTIBlacklistJob struct {
	db *pgxpool.Pool
}

func (j *purgeJTIBlacklistJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "purgeJTIBlacklistJob.run")

	return sqlc.New(j.db).PurgeExpiredJTIBlacklist(ctx, time.Now().UTC())
}

func (j *purgeJTIBlacklistJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run purge jti blacklist job", slog.Any("error", err))
	}
}

func NewPurgeJTIBlacklistJob(db *pgxpool.Pool) scheduler.Job {
	return &purgeJTIBlacklistJob{db: db}
}
