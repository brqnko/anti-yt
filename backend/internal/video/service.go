package video

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	videoQS VideoQueryService
}

func NewService(db *pgxpool.Pool) *Service {
	return new(Service{
		videoQS: NewVideoQueryService(db),
	})
}

func (s *Service) GetVideoDetail(ctx context.Context, userID uuid.UUID, videoID uuid.UUID) (_ GetVideoDetailView, err error) {
	defer util.Wrap(&err, "video.(*Service).GetVideoDetail")

	return s.videoQS.GetVideoDetail(ctx, userID, videoID)
}
