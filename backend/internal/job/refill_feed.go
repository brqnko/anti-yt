package job

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/feed"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

type refillFeedJob struct {
	db             *pgxpool.Pool
	feedRepo       database_d.FeedRepository
	feedQS         feed.FeedQueryService
	discordService discord_d.Service
	mx             *sync.Mutex
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

	userIDs, err := j.feedRepo.ListUserIDsWithFeedCountLessThan(ctx, 100)
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

		if err := j.feedRepo.PushMany(ctx, userID, videoIDs); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to push feed(refill feed job)", slog.String("user_id", userID.String()), slog.Any("error", err))
			continue
		}
		refilled++
	}

	if err := j.discordService.SendWebhookMessage(ctx, fmt.Sprintf("refill feed jobが完了しました (対象: %d, 補充: %d)", len(userIDs), refilled)); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to send discord webhook", slog.Any("error", err))
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
		if wErr := j.discordService.SendWebhookMessage(ctx, fmt.Sprintf("[Error] refill feed job: %v", err)); wErr != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to send discord webhook", slog.Any("error", wErr))
		}
	}
}

func NewRefillFeedJob(db *pgxpool.Pool, feedRepo database_d.FeedRepository, discordService discord_d.Service) scheduler.Job {
	return &refillFeedJob{
		db:             db,
		feedRepo:       feedRepo,
		feedQS:         feed.NewFeedQueryService(db),
		discordService: discordService,
		mx:             &sync.Mutex{},
	}
}
