package v1

import (
	"crypto/ed25519"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

var _ StrictServerInterface = (*Handler)(nil)

const (
	internalErrorTitle  = "internal Server Error"
	internalErrorDetail = "Something went wrong!"
)

type Handler struct {
	db *pgxpool.Pool

	authService *auth.Service
	serverURL   string
	frontendURL string
}

func NewHandler(db *pgxpool.Pool, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier, serverURL, frontendURL string, jwtPrivate ed25519.PrivateKey, jwtPublic ed25519.PublicKey) (*Handler, error) {
	authService, err := auth.NewService(db, oauth2Config, verifier, 30*time.Minute, 30*24*time.Hour, serverURL, jwtPrivate, jwtPublic)
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:          db,
		authService: authService,
		serverURL:   serverURL,
		frontendURL: frontendURL,
	}, nil
}
