package v1

import (
	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ StrictServerInterface = (*APIHandler)(nil)

const (
	internalErrorTitle  = "internal Server Error"
	internalErrorDetail = "Something went wrong!"
)

type APIHandler struct {
	db *pgxpool.Pool

	authService *auth.Service
	serverURL   string
	frontendURL string
}

func NewAPIHandler(db *pgxpool.Pool, oidcService oidc.GoogleOIDCService, serverURL, frontendURL string, jwtService jwt_d.JWTService) (*APIHandler, error) {
	authService, err := auth.NewService(db, oidcService, serverURL, jwtService)
	if err != nil {
		return nil, err
	}

	return &APIHandler{
		db:          db,
		authService: authService,
		serverURL:   serverURL,
		frontendURL: frontendURL,
	}, nil
}
