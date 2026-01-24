package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/labstack/echo/v4"
)

func main() {
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	h, err := v1.NewHandler()
	if err != nil {
		fmt.Printf("failed to create handler: %v\n", err)
		return 1
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := h.Close(ctx); err != nil {
			fmt.Printf("failed to close handler: %v\n", err)
		}
	}()
	e := echo.New()

	v1.RegisterHandlers(e, h)

	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Printf("failed to shutdown: %v\n", err)
		return 1
	}

	return 0
}
