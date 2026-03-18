package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) (*Service, error) {
	return &Service{
		db: db,
	}, nil
}

// TODO: YouTube Playlistからimportできるようにする
func (s *Service) CreatePlaylist(ctx context.Context, title, description, visibilityStr, playlistTypeStr string, basePlaylistUrl *string) (*Playlist, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pl, err := NewPlaylist(
		uuid.Nil,
		title,
		description,
		visibilityStr,
		playlistTypeStr,
		0,
		time.Time{},
		time.Time{},
		nil,
		[]*video.Video{},
	)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(s.db)
	row, err := q.CreatePlaylist(ctx, sqlc.CreatePlaylistParams{
		UserPublicID:        userID,
		PlaylistTitle:       string(*pl.Title),
		PlaylistDescription: string(*pl.Description),
		VisibilityCode:      int(pl.VisibilityCode),
		PlaylistCode:        int(pl.PlaylistCode),
	})
	if err != nil {
		return nil, fmt.Errorf("createPlaylist: %w", err)
	}

	pl.ID = row.PublicID
	pl.VideoCount = row.VideoCount
	pl.CreatedAt = row.CreatedAt
	pl.UpdatedAt = row.UpdatedAt

	return pl, nil
}
