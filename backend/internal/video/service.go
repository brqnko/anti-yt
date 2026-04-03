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
	return &Service{
		videoQS: NewVideoQueryService(db),
	}
}

func (s *Service) GetVideoDetail(ctx context.Context, userID, videoID uuid.UUID) (_ GetVideoDetailView, err error) {
	defer util.Wrap(&err, "video.(*Service).GetVideoDetail")

	view, err := s.videoQS.Find(ctx, userID, videoID)
	if err != nil {
		return GetVideoDetailView{}, err
	}

	return view, nil
}

