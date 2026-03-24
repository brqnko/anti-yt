package video

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

type VideoRepository interface {
	Save(ctx context.Context, video *Video) (int64, error)
}

type videoRepositoryImpl struct {
	q sqlc.Querier
}

func NewVideoRepository(q sqlc.Querier) VideoRepository {
	return &videoRepositoryImpl{
		q: q,
	}
}

func (v *videoRepositoryImpl) Save(ctx context.Context, video *Video) (_ int64, err error) {
	defer util.Wrap(&err, "videoRepository.Save")

	id, err := v.q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
		ChannelID:             video.ChannelID,
		ExternalID:            string(video.Video.ID),
		ExternalTitle:         video.Video.Title,
		ExternalDescription:   video.Video.Description,
		FetchedAt:             video.FetchedAt,
		ExternalCreatedAt:     video.Video.CreatedAt,
		ExternalThumbnailUrl:  video.Video.ThumbnailURL,
		ExternalLengthSeconds: video.Video.LengthSeconds,
		ID:                    video.ID,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

var _ VideoRepository = (*videoRepositoryImpl)(nil)
