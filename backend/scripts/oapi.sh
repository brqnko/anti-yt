#!/bin/bash
set -euo pipefail

BACKEND_DIR="${BACKEND_DIR:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "$BACKEND_DIR"

go generate internal/core/handler/*/generate.go