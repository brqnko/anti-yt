package video

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetVideoDetailView struct {
	VideoId                         uuid.UUID
	ExternalVideoId                 string
	ExternalVideoTitle              string
	ExternalVideoDescription        string
	ExternalVideoThumbnailUrl       string
	ExternalVideoCreatedAt          time.Time
	ChannelId                       uuid.UUID
	ChannelCustomId                 string
	ExternalChannelDisplayName      string
	ExternalChannelIconUrl          string
	ExternalChannelSubscribersCount uint64
	IsWatched                       bool
}

type VideoQueryService interface {
	Find(ctx context.Context, userID, videoID uuid.UUID) (GetVideoDetailView, error)
}

type videoQueryServiceImpl struct {
	q sqlc.Querier
}

func NewVideoQueryService(db *pgxpool.Pool) VideoQueryService {
	return &videoQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (v *videoQueryServiceImpl) Find(ctx context.Context, userID, videoID uuid.UUID) (_ GetVideoDetailView, err error) {
	defer util.Wrap(&err, "videoQueryService.Find(videoID=%s)", videoID)

	row, err := v.q.GetVideoDetail(ctx, sqlc.GetVideoDetailParams{
		UserID:  userID,
		VideoID: videoID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetVideoDetailView{}, err
		}
		return GetVideoDetailView{}, err
	}

	return GetVideoDetailView{
		VideoId:                         row.ID,
		ExternalVideoId:                 row.ExternalID,
		ExternalVideoTitle:              row.ExternalTitle,
		ExternalVideoDescription:        row.ExternalDescription,
		ExternalVideoThumbnailUrl:       row.ExternalThumbnailUrl,
		ExternalVideoCreatedAt:          row.ExternalCreatedAt,
		ChannelId:                       row.ChannelID,
		ChannelCustomId:                 row.ChannelCustomID,
		ExternalChannelDisplayName:      row.ExternalDisplayName,
		ExternalChannelIconUrl:          row.ExternalIconUrl,
		ExternalChannelSubscribersCount: uint64(row.ExternalSubscribersCount),
		IsWatched:                       row.IsWatched,
	}, nil
}

var _ VideoQueryService = (*videoQueryServiceImpl)(nil)
