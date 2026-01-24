package v1

import (
	"context"
	"database/sql"

	"github.com/brqnko/anti-yt/backend/internal/core/database"
)

var _ ServerInterface = (*Handler)(nil)

type Handler struct {
	db *sql.DB
}

func NewHandler() (*Handler, error) {
	db, err := database.ConnectDB()
	if err != nil {
		return nil, err
	}
	if err = database.RunMigration(db); err != nil {
		return nil, err
	}

	return &Handler{db: db}, nil
}

func (h *Handler) Close(ctx context.Context) error {
	if err := h.db.Close(); err != nil {
		return err
	}

	return nil
}
