package video

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidGetUploadLimit = util.NewDomainError("video.invalid_get_upload_limit", "invalid get upload limit: out of range (should be [1..50])")
)

// ChannelUploadRefresher refreshes channel uploads if the RSS cache is stale.
type ChannelUploadRefresher interface {
	RefreshChannelIfStale(ctx context.Context, channelID uuid.UUID) error
}

type Service struct {
	db        *pgxpool.Pool
	ytService youtube_d.YouTubeAPIService

	videoQS            VideoQueryService
	channelRefresher   ChannelUploadRefresher
}

func NewService(db *pgxpool.Pool, ytService youtube_d.YouTubeAPIService, channelRefresher ChannelUploadRefresher) (*Service, error) {
	return &Service{
		db:               db,
		ytService:        ytService,
		videoQS:          NewVideoQueryService(db),
		channelRefresher: channelRefresher,
	}, nil
}

func (s *Service) GetVideoDetail(ctx context.Context, videoID uuid.UUID) (GetVideoDetailView, error) {
	view, err := s.videoQS.Find(ctx, videoID)
	if err != nil {
		return GetVideoDetailView{}, err
	}

	return view, nil
}

func (s *Service) GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32) (videos []GetChannelUploadsView, hasNext bool, err error) {
	if limit < 1 || 50 < limit {
		return nil, false, ErrInvalidGetUploadLimit
	}

	if err := s.channelRefresher.RefreshChannelIfStale(ctx, channelID); err != nil {
		return nil, false, err
	}

	videos, err = s.videoQS.GetChannelUploads(ctx, userID, channelID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(videos) > int(limit) {
		return videos[:limit], true, nil
	}
	return videos, false, nil
}

func (s *Service) GetFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (videos []GetVideoFeedView, hasNext bool, err error) {
	videos, err = s.videoQS.GetVideoFeed(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(videos) > int(limit) {
		return videos[:limit], true, nil
	}
	return videos, false, nil
}
