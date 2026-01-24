package oidc

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func NewGoogleProvider(ctx context.Context, clientID string, clientSecret string, redirectURL string) (*oauth2.Config, *oidc.Provider, *oidc.IDTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, nil, nil, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID},
		Endpoint:     provider.Endpoint(),
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	return oauth2Config, provider, verifier, nil
}
