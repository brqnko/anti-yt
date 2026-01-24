package v1

import (
	"context"
	"database/sql"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"golang.org/x/oauth2"
)

var _ StrictServerInterface = (*Handler)(nil)

const (
	internalErrorTitle  = "internal Server Error"
	internalErrorDetail = "Something went wrong!"
)

type Handler struct {
	db *sql.DB

	authService *auth.Service
}

func NewHandler(db *sql.DB, oauth2Config *oauth2.Config) (*Handler, error) {
	authService, err := auth.NewService(db, oauth2Config)
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:          db,
		authService: authService,
	}, nil
}

func newContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
