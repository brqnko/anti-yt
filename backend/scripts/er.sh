#!/bin/bash
set -euo pipefail

DB_HOST="${DB_HOST:-db}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DATABASE_NAME:-example}"
DB_PASSWORD="$(cat /run/secrets/db-password)"

DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
DOC_PATH="docs/schema"

tbls doc "$DSN" "$DOC_PATH" --er-format svg --rm-dist --force
