package video

import (
	"context"

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

func (s *Service) GetVideoDetail(ctx context.Context, videoID uuid.UUID) (GetVideoDetailView, error) {
	view, err := s.videoQS.Find(ctx, videoID)
	if err != nil {
		return GetVideoDetailView{}, err
	}

	return view, nil
}

