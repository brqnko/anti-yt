package job

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/feed"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

type refillFeedJob struct {
	db       *pgxpool.Pool
	feedRepo database_d.FeedRepository
	feedQS   feed.FeedQueryService
	mx       *sync.Mutex
}

func (j *refillFeedJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*refillFeedJob).run")

	q := sqlc.New(j.db)

	if err := database_d.TryAdLockSession(ctx, q, []byte("refillFeedJob")); err != nil {
		return err
	}
	defer func() {
		if err := database_d.ReleaseAdLock(ctx, q, []byte("refillFeedJob")); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to release ad lock(refill feed job)", slog.Any("error", err))
		}
	}()

	userIDs, err := j.feedQS.ListAllActiveUserIDs(ctx)
	if err != nil {
		return err
	}

	var refilled int
	for _, userID := range userIDs {
		videoIDs, err := j.feedQS.ListSubscriptionVideoIDs(ctx, userID, 1000)
		if err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to list subscription video ids(refill feed job)", slog.String("user_id", userID.String()), slog.Any("error", err))
			continue
		}
		if len(videoIDs) == 0 {
			continue
		}

		if err := j.feedRepo.DeleteAll(ctx, userID); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to clear feed(refill feed job)", slog.String("user_id", userID.String()), slog.Any("error", err))
			continue
		}
		if err := j.feedRepo.PushMany(ctx, userID, videoIDs); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to push feed(refill feed job)", slog.String("user_id", userID.String()), slog.Any("error", err))
			continue
		}
		refilled++
	}

	return nil
}

func (j *refillFeedJob) Run() {
	j.mx.Lock()
	defer j.mx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run refill feed job", slog.Any("error", err))
	}
}

func NewRefillFeedJob(db *pgxpool.Pool, feedRepo database_d.FeedRepository) scheduler.Job {
	return &refillFeedJob{
		db:       db,
		feedRepo: feedRepo,
		feedQS:   feed.NewFeedQueryService(db),
		mx:       new(sync.Mutex{}),
	}
}
