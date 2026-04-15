#!/bin/bash
set -euo pipefail

DOC_PATH="docs/schema"

tbls doc "$DATABASE_URL" "$DOC_PATH" --er-format svg --rm-dist --force
