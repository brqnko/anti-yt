package testutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type gooseMigrator struct{}

func (g *gooseMigrator) Hash() (string, error) {
	h := sha256.New()
	err := fs.WalkDir(migrations.EmbedMigrations, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := migrations.EmbedMigrations.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (g *gooseMigrator) Migrate(ctx context.Context, db *sql.DB, _ pgtestdb.Config) error {
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, ".")
}

// secretsPathMap は環境変数名に対応する secrets ファイルパス
var secretsPathMap = map[string]string{
	"DATABASE_PASSWORD": "/run/secrets/db-password",
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	if v := os.Getenv(key); v != "" {
		return v
	}
	if path, ok := secretsPathMap[key]; ok {
		data, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	t.Fatalf("required environment variable %s is not set (and no secret found)", key)
	return ""
}

// NewTestDB creates an isolated test database using pgtestdb.
func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	conf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       requireEnv(t, "DATABASE_HOST"),
		Port:       requireEnv(t, "DATABASE_PORT"),
		User:       requireEnv(t, "DATABASE_USER"),
		Password:   requireEnv(t, "DATABASE_PASSWORD"),
		Database:   requireEnv(t, "DATABASE_NAME"),
		Options:    "sslmode=disable",
		TestRole: &pgtestdb.Role{
			Username:     pgtestdb.DefaultRoleUsername,
			Password:     pgtestdb.DefaultRolePassword,
			Capabilities: "SUPERUSER",
		},
	}

	dbConf := pgtestdb.Custom(t, conf, &gooseMigrator{})

	pool, err := pgxpool.New(context.Background(), dbConf.URL())
	if err != nil {
		t.Fatalf("failed to create pgxpool: %s", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
