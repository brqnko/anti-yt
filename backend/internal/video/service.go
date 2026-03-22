package video

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db        *pgxpool.Pool
	ytService youtube_d.YouTubeAPIService
}

func NewService(db *pgxpool.Pool, ytService youtube_d.YouTubeAPIService) (*Service, error) {
	return &Service{
		db:        db,
		ytService: ytService,
	}, nil
}

func (s *Service) GetVideoDetail(ctx context.Context, videoID uuid.UUID) (VideoDetail, error) {
	videoDetail, err := sqlc.New(s.db).GetVideoDetail(ctx, videoID)
	if err != nil {
		return VideoDetail{}, fmt.Errorf("getVideoDetail: %w", err)
	}

	video, err := NewVideoDetail(
		videoDetail.ID,
		videoDetail.ExternalID,
		videoDetail.ExternalTitle,
		videoDetail.ExternalDescription,
		videoDetail.ExternalThumbnailUrl,
		videoDetail.ChannelID,
		videoDetail.ChannelExternalID,
		videoDetail.ExternalDisplayName,
		videoDetail.ChannelCustomID,
		videoDetail.ExternalIconUrl,
		int(videoDetail.ExternalSubscribersCount),
	)
	if err != nil {
		return VideoDetail{}, fmt.Errorf("newVideoDetail: %w", err)
	}

	return video, nil
}

func (s *Service) Heartbeat(ctx context.Context, videoID uuid.UUID, positionSeconds int) (int, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()
	q := sqlc.New(tx)

	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, util.Sha256Int64(userID[:]))
	if err != nil {
		return 0, fmt.Errorf("tryAcquireAdvisoryXactLock: %w", err)
	}
	if !acquired {
		return 0, fmt.Errorf("tryAcquireAdvisoryXactLock: lock not acquired")
	}

	if err := q.Heartbeat(ctx, sqlc.HeartbeatParams{
		WatchPositionSeconds: positionSeconds,
		UserPublicID:         userID,
		VideoPublicID:        videoID,
	}); err != nil {
		return 0, err
	}

	watchStats, err := q.GetTotalWatchSeconds(ctx, userID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return user.CalcRemainingSeconds(watchStats.DailyLimitSeconds, watchStats.TodayWatchTotal), nil
}
