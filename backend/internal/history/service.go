package history

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool

	historyQS HistoryQueryService
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:        db,
		historyQS: NewHistoryQueryService(db),
	}
}

func (s *Service) Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int, playlistID *uuid.UUID, loc *time.Location) (_ *int, err error) {
	defer util.Wrap(&err, "history.(*Service).Heartbeat")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	if err := database_d.TryAdLock(ctx, q, userID[:]); err != nil {
		return nil, err
	}

	// TODO: publish recent playlist
	lastHeartbeat, lastVideoLength, lastHeartbeatID, lastUpdatedAt, err := NewHistoryRepository(q).GetLastHeartbeatForUpdate(ctx, userID)
	if err == nil { // 最後のheartbeatの取得に成功
		heartbeat, err := lastHeartbeat.Rotate(videoID, positionSeconds, lastVideoLength, lastUpdatedAt)
		if err != nil {
			return nil, err
		}

		// 動画を最後まで視聴した場合は既読にして、再生位置を0にリセットする
		if lastHeartbeat.WatchPositionSeconds.IsFinished(lastVideoLength) {
			_ = NewHistoryRepository(q).MarkVideoWatched(ctx, userID, lastHeartbeat.VideoID)
			lastHeartbeat.WatchPositionSeconds = 0
		}

		if err := NewHistoryRepository(q).UpdateHeartbeat(ctx, lastHeartbeatID, lastHeartbeat); err != nil {
			return nil, err
		}

		if heartbeat != nil {
			if err := NewHistoryRepository(q).CreateHeartbeat(ctx, heartbeat); err != nil {
				return nil, err
			}
		}
	} else if errors.Is(err, core.ErrNotFound) { // 初めてのHeartbeatの場合
		// 普通にheartbeatを作成して挿入する
		heartbeat, err := NewHeartbeat(videoID, userID, positionSeconds)
		if err != nil {
			return nil, err
		}
		if err := NewHistoryRepository(q).CreateHeartbeat(ctx, heartbeat); err != nil {
			return nil, err
		}
	} else { // DBエラー
		return nil, err
	}

	// 最近再生したプレイリストに追加
	if playlistID != nil {
		if err := playlist.NewPlaylistRepository(q).PushRecentPlaylistID(ctx, userID, *playlistID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	remaining, err := s.historyQS.FindTotalWatchSeconds(ctx, userID, loc)
	if err != nil {
		return nil, err
	}
	if remaining == math.MaxInt {
		return nil, nil
	}
	return &remaining, nil
}

func (s *Service) MarkVideoWatched(ctx context.Context, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "history.(*Service).MarkVideoWatched")

	return NewHistoryRepository(sqlc.New(s.db)).MarkVideoWatched(ctx, userID, videoID)
}

func (s *Service) UnmarkVideoWatched(ctx context.Context, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "history.(*Service).UnmarkVideoWatched")

	return NewHistoryRepository(sqlc.New(s.db)).UnmarkVideoWatched(ctx, userID, videoID)
}

func (s *Service) GetHistory(ctx context.Context, userID uuid.UUID, limit int, cursor *uuid.UUID, loc *time.Location) (_ []GetHistoryView, _ bool, err error) {
	defer util.Wrap(&err, "history.(*Service).GetHistory")

	views, err := s.historyQS.FindHistory(ctx, userID, cursor, int32(limit+1))
	if err != nil {
		return nil, false, err
	}

	for i := range views {
		views[i].WatchedAt = views[i].WatchedAt.In(loc)
	}

	if len(views) > limit {
		return views[:limit], true, nil
	}
	return views, false, nil
}

func (s *Service) GetStatisticsByWeek(ctx context.Context, userID uuid.UUID, targetWeek time.Time, loc *time.Location) (_ *string, _ []GetStatisticsWeeklyView, err error) {
	defer util.Wrap(&err, "history.(*Service).GetStatisticsByWeek")

	aiSummary, views, err := s.historyQS.FindStatisticsByWeek(ctx, userID, targetWeek)
	if err != nil {
		return nil, nil, err
	}

	for i := range views {
		views[i].WatchDate = views[i].WatchDate.In(loc)
	}

	return aiSummary, views, nil
}
