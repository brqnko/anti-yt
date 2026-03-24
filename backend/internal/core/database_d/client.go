package database_d

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigration(ctx context.Context, dbUser, dbPassword, dbHost string, dbPort int, dbName, dbSSLMode string) error {
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode))
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close db in migration", "error", err)
		}
	}()

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

func ConnectDB(ctx context.Context, dbUser, dbPassword, dbHost string, dbPort int, dbName, dbSSLMode string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
