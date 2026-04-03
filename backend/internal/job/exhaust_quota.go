package job

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/report"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exhaustQuotaJob struct {
	db            *pgxpool.Pool
	ytService     youtube_d.Service
	reportService *report.Service
	mx            *sync.Mutex
}

func (j *exhaustQuotaJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*exhaustQuotaJob).run")

	tx, err := j.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback in exhaustQuotaJob.run", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// ad lock
	if err := database_d.TryAdLock(ctx, q, []byte("exhaustQuotaJob")); err != nil {
		return err
	}

	// TODO

	if err := tx.Commit(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return err
	}

	return nil
}

func (j *exhaustQuotaJob) Run() {
	// クオータリセットはPT midnight。cronは夏冬両方で登録されるので、
	// リセットまで10分以内でなければスキップする。
	loc, _ := time.LoadLocation("America/Los_Angeles")
	now := time.Now().In(loc)
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	if nextMidnight.Sub(now) > 15*time.Minute {
		slog.Info("skipping exhaust quota job: not close enough to quota reset")
		return
	}

	j.mx.Lock()
	defer j.mx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run exhaust quota job", slog.Any("error", err))
	}
}

func NewExhaustQuotaJob(db *pgxpool.Pool, ytService youtube_d.Service, reportService *report.Service) scheduler.Job {
	return &exhaustQuotaJob{
		db:            db,
		ytService:     ytService,
		reportService: reportService,
		mx:            &sync.Mutex{},
	}
}
