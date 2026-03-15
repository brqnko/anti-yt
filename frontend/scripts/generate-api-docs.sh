#!/bin/sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OPENAPI_FILE="$PROJECT_ROOT/shared/api/v1/openapi.yaml"
OUTPUT_FILE="$PROJECT_ROOT/shared/api/v1/index.html"

if [ ! -f "$OPENAPI_FILE" ]; then
  echo "Error: OpenAPI spec not found at $OPENAPI_FILE" >&2
  exit 1
fi

npx --yes @redocly/cli build-docs "$OPENAPI_FILE" --output "$OUTPUT_FILE"

echo "Generated: $OUTPUT_FILE"
