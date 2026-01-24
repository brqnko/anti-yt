package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"

	"golang.org/x/oauth2"
)

type Service struct {
	db *sql.DB

	oauth2Config *oauth2.Config
}

func NewService(db *sql.DB, oauth2Config *oauth2.Config) (*Service, error) {
	return &Service{
		db:           db,
		oauth2Config: oauth2Config,
	}, nil
}

func (s Service) CreateAuthCode(ctx context.Context) (redirectURL string, state string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	state = base64.URLEncoding.EncodeToString(b)

	return s.oauth2Config.AuthCodeURL(state), state, nil
}
