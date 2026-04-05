package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"
	"github.com/stretchr/testify/require"
)

func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	password, err := os.ReadFile("/run/secrets/db-password")
	require.NoError(t, err, "failed to read /run/secrets/db-password")

	conf := pgtestdb.Config{
		DriverName: "pgx",
		User:       os.Getenv("DATABASE_USER"),
		Password:   strings.TrimSpace(string(password)),
		Host:       os.Getenv("DATABASE_HOST"),
		Port:       os.Getenv("DATABASE_PORT"),
		Options:    "sslmode=" + os.Getenv("DATABASE_SSLMODE"),
	}
	migrator := goosemigrator.New(".", goosemigrator.WithFS(migrations.EmbedMigrations))
	db := pgtestdb.New(t, conf, migrator)

	var dbName string
	require.NoError(t, db.QueryRow("SELECT current_database()").Scan(&dbName))

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		conf.User, conf.Password, conf.Host, conf.Port, dbName,
		os.Getenv("DATABASE_SSLMODE"),
	))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}
