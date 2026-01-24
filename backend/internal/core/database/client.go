package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigration(db *sql.DB) error {
	slog.Info("running migration")
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			slog.Error("migration error", "message", pgErr.Message, "detail", pgErr.Detail, "hint", pgErr.Hint)
		}
		return err
	}

	slog.Info("migration completed")
	return nil
}

func ConnectDB(dbPassword string, dbName string) (*sql.DB, error) {
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:%s@db:5432/%s?sslmode=disable", dbPassword, dbName))
	if err != nil {
		return nil, err
	}

	return db, nil
}
