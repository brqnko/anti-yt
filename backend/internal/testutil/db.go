package testutil

import (
	"context"
	"net/url"
	"os"
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

	dbURL := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dbURL, "DATABASE_URL is not set")

	u, err := url.Parse(dbURL)
	require.NoError(t, err, "failed to parse DATABASE_URL")

	password, _ := u.User.Password()
	conf := pgtestdb.Config{
		DriverName: "pgx",
		User:       u.User.Username(),
		Password:   password,
		Host:       u.Hostname(),
		Port:       u.Port(),
		Options:    u.RawQuery,
	}
	migrator := goosemigrator.New(".", goosemigrator.WithFS(migrations.EmbedMigrations))
	db := pgtestdb.New(t, conf, migrator)

	var dbName string
	require.NoError(t, db.QueryRow("SELECT current_database()").Scan(&dbName))

	poolURL := *u
	poolURL.Path = "/" + dbName
	pool, err := pgxpool.New(context.Background(), poolURL.String())
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}
