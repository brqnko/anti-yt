package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database"
	"github.com/brqnko/anti-yt/backend/internal/core/database/sqlc"
	middleware2 "github.com/brqnko/anti-yt/backend/internal/core/handler/middleware"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	os.Exit(run(context.Background()))
}

type config struct {
	env                    string
	oidcGoogleClientId     string
	oidcGoogleRedirectURL  string
	oidcGoogleClientSecret string
	serverURL              string
	frontendURL            string

	dbPassword string
	dbName     string
}

func run(ctx context.Context) int {
	dbPassword, err := os.ReadFile("/run/secrets/db-password")
	if err != nil {
		fmt.Printf("failed to read db-password: %v\n", err)
		return 1
	}
	oidcGoogleClientSecret, err := os.ReadFile("/run/secrets/oidc-google-client-secret")
	if err != nil {
		fmt.Printf("failed to read oidc-google-client-secret: %v\n", err)
		return 1
	}
	jwtPrivate, err := loadPrivateKey("/run/secrets/jwt-private")
	if err != nil {
		fmt.Printf("failed to load jwt-private: %v\n", err)
		return 1
	}
	jwtPublic, err := loadPublicKey("/run/secrets/jwt-public")
	if err != nil {
		fmt.Printf("failed to load jwt-public: %v\n", err)
		return 1
	}
	cfg := config{
		env:                    os.Getenv("ENV"),
		oidcGoogleRedirectURL:  os.Getenv("OIDC_GOOGLE_REDIRECT_URL"),
		oidcGoogleClientId:     os.Getenv("OIDC_GOOGLE_CLIENT_ID"),
		oidcGoogleClientSecret: strings.TrimSpace(string(oidcGoogleClientSecret)),
		dbPassword:             strings.TrimSpace(string(dbPassword)),
		dbName:                 os.Getenv("DATABASE_NAME"),
		serverURL:              os.Getenv("SERVER_URL"),
		frontendURL:            os.Getenv("FRONTEND_URL"),
	}
	clear(dbPassword)
	clear(oidcGoogleClientSecret)

	var handler slog.Handler
	if cfg.env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	if err = database.InitDB(initCtx, cfg.dbPassword, cfg.dbName); err != nil {
		slog.Error("failed to run migration", "error", err)
		return 1
	}

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	db, err := database.ConnectDB(initCtx, cfg.dbPassword, cfg.dbName)
	if err != nil {
		slog.Error("failed to connect db", "error", err)
		return 1
	}
	defer db.Close()

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := sqlc.New(db).CleanupExpiredJTIBlacklist(initCtx, time.Now().UTC()); err != nil {
		slog.Error("failed to clean up jti blacklist", "error", err)
	}

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	oauth2Config, _, verifier, err := oidc.NewGoogleProvider(initCtx, cfg.oidcGoogleClientId, cfg.oidcGoogleClientSecret, cfg.oidcGoogleRedirectURL)
	if err != nil {
		slog.Error("failed to create oidc provider", "error", err)
		return 1
	}

	h, err := v1.NewHandler(db, oauth2Config, verifier, cfg.serverURL, cfg.frontendURL, jwtPrivate, jwtPublic)
	if err != nil {
		slog.Error("failed to create handler", "error", err)
		return 1
	}

	e := echo.New()
	// NOTE: RateLimit, gzip, loggerはLB(pingora, nginx等)で行う。各リクエストへのレートリミットはHandlerが行う
	e.Use(middleware.Recover(), middleware.Secure())
	v1.RegisterHandlers(e, v1.NewStrictHandler(h, []v1.StrictMiddlewareFunc{
		middleware2.AccessTokenMiddleware(jwtPublic, db),
		middleware2.CsrfMiddleware,
		middleware2.AuthTokensMiddleware,
		middleware2.RequestIDMiddleware,
	}))

	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("shutting down the server")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
		return 1
	}

	return 0
}

func loadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to find block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("key is not ed25519 private key")
	}
	return privateKey, nil
}

func loadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to find block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("key is not ed25519 public key")
	}
	return privateKey, nil
}
