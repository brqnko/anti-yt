package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
)

var _ ServerInterface = (*Handler)(nil)

type Handler struct {
	db *sql.DB
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
