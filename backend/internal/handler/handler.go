package handler

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
)

var _ ServerInterface = (*Handler)(nil)

type Handler struct {
	db *sql.DB
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigration(db *sql.DB) error {
	fmt.Println("running migration")
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// "migrations" は go:embed で指定したディレクトリ名
	if err := goose.Up(db, "migrations"); err != nil {
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
	var val int
	err := h.db.QueryRow("SELECT 1 + 1").Scan(&val)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, fmt.Sprintf("OK: %d", val))
}
