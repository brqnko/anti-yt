package video

import (
	"context"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db      *pgxpool.Pool
	videoQS VideoQueryService
}

func NewService(db *pgxpool.Pool) *Service {
	return new(Service{
		db:      db,
		videoQS: NewVideoQueryService(db),
	})
}

func (s *Service) GetVideoDetail(ctx context.Context, userID uuid.UUID, videoID uuid.UUID) (_ GetVideoDetailView, err error) {
	defer util.Wrap(&err, "video.(*Service).GetVideoDetail")

	detail, err := s.videoQS.GetVideoDetail(ctx, userID, videoID)
	if err != nil {
		return GetVideoDetailView{}, err
	}

	logger := util.LoggerFromContext(ctx)
	channelID := detail.ChannelId
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sqlc.New(s.db).MarkChannelSeen(bgCtx, channelID); err != nil {
			logger.WarnContext(bgCtx, "failed to mark channel as seen",
				slog.String("channel_id", channelID.String()), slog.Any("error", err))
		}
	}()

	return detail, nil
}
