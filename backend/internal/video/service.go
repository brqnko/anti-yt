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

func (s *Service) GetVideoDetail(ctx context.Context, userID uuid.UUID, rawID string) (_ GetVideoDetailView, err error) {
	defer util.Wrap(&err, "video.(*Service).GetVideoDetail")

	var videoID *uuid.UUID
	var externalVideoID *string
	var b util.Base64UUID
	if parseErr := b.UnmarshalText([]byte(rawID)); parseErr == nil {
		id := b.UUID()
		videoID = &id
	} else {
		externalVideoID = &rawID
	}

	return s.videoQS.FindByIDOrExternalID(ctx, userID, videoID, externalVideoID)
}
