package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigration(db *sql.DB) error {
	fmt.Println("Running migration...")
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Redo(db, "."); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Printf("Error: %s, Detail: %s, Hint: %s\n", pgErr.Message, pgErr.Detail, pgErr.Hint)
		}
		return err
	}

	fmt.Println("migration ok")
	return nil
}

func ConnectDB() (*sql.DB, error) {
	data, err := os.ReadFile("/run/secrets/db-password")
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:%s@db:5432/%s?sslmode=disable", strings.TrimSpace(string(data)), "example"))
	if err != nil {
		return nil, err
	}

	return db, nil
}
