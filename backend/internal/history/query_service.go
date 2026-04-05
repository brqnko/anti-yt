package history

import (
	"context"
	"math"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetHistoryView struct {
	WatchId                    uuid.UUID
	VideoId                    uuid.UUID
	ExternalVideoTitle         string
	ExternalVideoThumbnailUrl  string
	ExternalVideoLengthSeconds int
	WatchPositionSeconds       int
	WatchedAt                  time.Time
	ChannelId                  uuid.UUID
	ExternalChannelDisplayName string
	ExternalChannelIconUrl     string
}

type GetStatisticsWeeklyView struct {
	WatchDate  time.Time
	VideoCount int
	WatchSum   int64
}

type HistoryQueryService interface {
	FindHistory(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetHistoryView, error)
	FindStatisticsByWeek(ctx context.Context, userID uuid.UUID, startDate time.Time) (*string, []GetStatisticsWeeklyView, error)
	// FindTotalWatchSeconds は今日の残り視聴可能秒数を返す。無制限の場合は math.MaxInt を返す。
	FindTotalWatchSeconds(ctx context.Context, userID uuid.UUID, loc *time.Location) (int, error)
}

type historyQueryServiceImpl struct {
	q sqlc.Querier
}

func NewHistoryQueryService(db *pgxpool.Pool) HistoryQueryService {
	return &historyQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (h *historyQueryServiceImpl) FindHistory(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetHistoryView, err error) {
	defer util.Wrap(&err, "history.(*historyQueryServiceImpl).FindHistory(userID=%s)", userID)

	rows, err := h.q.ListWatchHistory(ctx, sqlc.ListWatchHistoryParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	views := make([]GetHistoryView, len(rows))
	for i, row := range rows {
		views[i] = GetHistoryView{
			WatchId:                    row.WatchID,
			VideoId:                    row.VideoID,
			ExternalVideoTitle:         row.ExternalVideoTitle,
			ExternalVideoThumbnailUrl:  row.ExternalVideoThumbnailUrl,
			ExternalVideoLengthSeconds: row.ExternalVideoLengthSeconds,
			WatchPositionSeconds:       row.WatchPositionSeconds,
			WatchedAt:                  row.WatchedAt,
			ChannelId:                  row.ChannelID,
			ExternalChannelDisplayName: row.ExternalChannelDisplayName,
			ExternalChannelIconUrl:     row.ExternalChannelIconUrl,
		}
	}
	return views, nil
}

func (h *historyQueryServiceImpl) FindStatisticsByWeek(ctx context.Context, userID uuid.UUID, startDate time.Time) (_ *string, _ []GetStatisticsWeeklyView, err error) {
	defer util.Wrap(&err, "history.(*historyQueryServiceImpl).FindStatisticsByWeek(userID=%s)", userID)

	_, offsetSec := startDate.Zone()
	rows, err := h.q.ListDailyWatchStatsByRange(ctx, sqlc.ListDailyWatchStatsByRangeParams{
		TzOffset:  offsetSec,
		UserID:    userID,
		StartDate: startDate,
		EndDate:   startDate.Add(7 * 24 * time.Hour), // NOTE: postgresql側で+ '7 days'するとsqlcがパースエラー起こす
	})
	if err != nil {
		return nil, nil, err
	}

	views := make([]GetStatisticsWeeklyView, len(rows))
	for i, row := range rows {
		views[i] = GetStatisticsWeeklyView{
			WatchDate:  row.WatchDate,
			VideoCount: int(row.VideoCount),
			WatchSum:   row.WatchSum,
		}
	}

	var aiSummary *string
	summary, err := h.q.GetLatestMonthlyVideoWatchSummary(ctx, userID)
	if err == nil {
		desc := summary.AiSummaryDescription
		aiSummary = &desc
	}

	return aiSummary, views, nil
}

func (h *historyQueryServiceImpl) FindTotalWatchSeconds(ctx context.Context, userID uuid.UUID, loc *time.Location) (_ int, err error) {
	defer util.Wrap(&err, "history.(*historyQueryServiceImpl).FindTotalWatchSeconds(userID=%s)", userID)

	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	row, err := h.q.GetDailyWatchSummary(ctx, sqlc.GetDailyWatchSummaryParams{
		PublicID:    userID,
		TodayStart: todayStart,
	})
	if err != nil {
		return 0, err
	}

	if row.DailyLimitSeconds >= 24*60*60 {
		return math.MaxInt, nil
	}

	remaining := row.DailyLimitSeconds - row.TodayWatchTotal
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

var _ HistoryQueryService = (*historyQueryServiceImpl)(nil)
