package video

import (
	"context"
	"fmt"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
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

func (s *Service) GetVideoDetail(ctx context.Context, videoId uuid.UUID) (*VideoDetail, error) {
	videoDetail, err := sqlc.New(s.db).GetVideoDetail(ctx, videoId)
	if err != nil {
		return nil, fmt.Errorf("getVideoDetail: %w", err)
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
		return nil, fmt.Errorf("newVideoDetail: %w", err)
	}

	return video, nil
}

func (s *Service) Heartbeat(ctx context.Context, videoId uuid.UUID) error {
	_, ok := util.UserIDFromContext(ctx)
	if !ok {
		return util.ErrUserIDNotFoundInContext
	}

	return nil
}
