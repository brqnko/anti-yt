package database_d

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigration(ctx context.Context, dbUser, dbPassword, dbHost string, dbPort int, dbName, dbSSLMode string) (err error) {
	defer util.Wrap(&err, "database_d.RunMigration")

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode))
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to close db in migration", slog.Any("error", err))
		}
	}()

	util.LoggerFromContext(ctx).InfoContext(ctx, "running migration")
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "migration error", slog.String("message", pgErr.Message), slog.String("detail", pgErr.Detail), slog.String("hint", pgErr.Hint))
		}
		return err
	}

	util.LoggerFromContext(ctx).InfoContext(ctx, "migration completed")
	return nil
}

func ConnectPostgres(ctx context.Context, dbUser, dbPassword, dbHost string, dbPort int, dbName, dbSSLMode string) (_ *pgxpool.Pool, err error) {
	defer util.Wrap(&err, "database_d.ConnectPostgres")

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

func ConnectRedis(ctx context.Context, url string) (_ *redis.Client, err error) {
	defer util.Wrap(&err, "database_d.ConnectRedis")

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opt), nil
}
