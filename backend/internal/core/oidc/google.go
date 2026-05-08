package oidc

import (
	"context"
	"errors"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
)

var ErrIDTokenNotFound = core.NewDomainError("auth.id_token_not_found", "id token not found", core.StatusBadRequest)

type GoogleClient interface {
	AuthCodeURL(state string) string
	ExchangeAndVerify(ctx context.Context, code string) (_ string, err error)
}

type googleClient struct {
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func NewGoogleClient(ctx context.Context, clientID, clientSecret, redirectURL string) (_ GoogleClient, err error) {
	defer util.Wrap(&err, "oidc.NewGoogleClient")

	if clientSecret == "" {
		return nil, errors.New("google client secret is empty")
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, new(http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}))
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}

	oauth2Config := new(oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID},
		Endpoint:     provider.Endpoint(),
	})

	verifier := provider.Verifier(new(oidc.Config{
		ClientID: clientID,
	}))

	return &googleClient{
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
}

func (s *googleClient) AuthCodeURL(state string) string {
	return s.oauth2Config.AuthCodeURL(state)
}

func (s *googleClient) ExchangeAndVerify(ctx context.Context, code string) (_ string, err error) {
	defer util.Wrap(&err, "oidc.(*googleClient).ExchangeAndVerify")

	ctx = context.WithValue(ctx, oauth2.HTTPClient, new(http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}))
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
