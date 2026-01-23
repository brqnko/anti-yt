package v1

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
)

var _ ServerInterface = (*Handler)(nil)

type Handler struct {
	db *sql.DB
}

func (h *Handler) PostAuthGoogle(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetAuthGoogleCallback(ctx echo.Context, params GetAuthGoogleCallbackParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthRefresh(ctx echo.Context, params PostAuthRefreshParams) error {
	//TODO implement me
	panic("implement me")
}

func RunMigration(db *sql.DB) error {
	fmt.Println("running migration")
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// "migrations" は go:embed で指定したディレクトリ名
	if err := goose.Up(db, "."); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Printf("Error: %s, Detail: %s, Hint: %s\n", pgErr.Message, pgErr.Detail, pgErr.Hint)
		}
		return err
	}

	fmt.Println("migration ok")
	return nil
}

func NewHandler() *Handler {
	data, err := os.ReadFile("/run/secrets/db-password")
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:%s@db:5432/example?sslmode=disable", strings.TrimSpace(string(data))))
	if err != nil {
		log.Fatal(err)
	}
	if err = RunMigration(db); err != nil {
		log.Fatal(err)
	}

	return &Handler{db: db}
}

func (h *Handler) GetHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, fmt.Sprintf("OK"))
}
