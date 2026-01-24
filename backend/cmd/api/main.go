package main

import (
	"context"
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
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/labstack/echo/v4"
)

func main() {
	os.Exit(run(context.Background()))
}

type config struct {
	env                    string
	oidcGoogleClientId     string
	oidcGoogleRedirectURL  string
	oidcGoogleClientSecret string

	dbPassword string
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
	cfg := config{
		env:                    os.Getenv("ENV"),
		oidcGoogleRedirectURL:  os.Getenv("OIDC_GOOGLE_REDIRECT_URL"),
		oidcGoogleClientId:     os.Getenv("OIDC_GOOGLE_CLIENT_ID"),
		oidcGoogleClientSecret: strings.TrimSpace(string(oidcGoogleClientSecret)),
		dbPassword:             strings.TrimSpace(string(dbPassword)),
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

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	oauth2Config, _, _, err := oidc.NewGoogleProvider(initCtx, cfg.oidcGoogleClientId, cfg.oidcGoogleClientSecret, cfg.oidcGoogleRedirectURL)
	if err != nil {
		slog.Error("failed to create oidc provider", "error", err)
		return 1
	}

	db, err := database.ConnectDB(cfg.dbPassword, "example")
	if err != nil {
		slog.Error("failed to connect db", "error", err)
		return 1
	}
	cfg.dbPassword = ""
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close db connection", "error", err)
		}
	}()

	if err = database.RunMigration(db); err != nil {
		slog.Error("failed to run migration", "error", err)
		return 1
	}

	h, err := v1.NewHandler(db, oauth2Config)
	if err != nil {
		slog.Error("failed to create handler", "error", err)
		return 1
	}

	e := echo.New()

	v1.RegisterHandlers(e, v1.NewStrictHandler(h, nil))

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
