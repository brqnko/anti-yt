package oidc

import (
	"context"
	"errors"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var ErrIDTokenNotFound = core.NewDomainError("auth.id_token_not_found", "id token not found", core.StatusBadRequest)

type GoogleOIDCService interface {
	AuthCodeURL(state string) string
	ExchangeAndVerify(ctx context.Context, code string) (_ string, err error)
}

type googleOIDCService struct {
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func NewGoogleOIDCService(ctx context.Context, clientID, clientSecret, redirectURL string) (_ GoogleOIDCService, err error) {
	defer util.Wrap(&err, "oidc.NewGoogleOIDCService")

	if clientSecret == "" {
		return nil, errors.New("google client secret is empty")
	}

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID},
		Endpoint:     provider.Endpoint(),
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return &googleOIDCService{
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
}

func (s *googleOIDCService) AuthCodeURL(state string) string {
	return s.oauth2Config.AuthCodeURL(state)
}

func (s *googleOIDCService) ExchangeAndVerify(ctx context.Context, code string) (_ string, err error) {
	defer util.Wrap(&err, "oidc.(*googleOIDCService).ExchangeAndVerify")

	oauth2Token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", ErrIDTokenNotFound
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", err
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", err
	}

	return claims.Sub, nil
}
