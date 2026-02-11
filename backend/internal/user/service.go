package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) (*Service, error) {
	return &Service{db: db}, nil
}

func (s *Service) CreateNewUser(ctx context.Context, dailyScreenLimit *int, screenLimits []struct{ start, end int }, displayName *string, languageCode string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	var dailyScreenLimitDuration *time.Duration
	if dailyScreenLimit != nil {
		duration := time.Second * time.Duration(*dailyScreenLimit)
		dailyScreenLimitDuration = &duration
	}
	_, err = NewCreateUserInput(displayName, languageCode, dailyScreenLimitDuration, screenLimits)
	if err != nil {
		return err
	}

	q.CreateUserScreenTimeRanges(ctx, []sqlc.CreateUserScreenTimeRangesParams{})

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
