package admin

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool

	ytService youtube_d.Service
}

func NewService(db *pgxpool.Pool, ytService youtube_d.Service) *Service {
	return &Service{
		db:        db,
		ytService: ytService,
	}
}

func (s *Service) CreateNewValuableChannel(ctx context.Context, externalChannelID string, reason, description string) (_ *channel.ValuableChannel, err error)

func (s *Service) UpdateValuableChannel(ctx context.Context, externalChannelID string, reaason, description *string) (_ *channel.ValuableChannel, err error)

func (s *Service) RemoveValuableChannel(ctx context.Context, externalChannelID string) (err error)
