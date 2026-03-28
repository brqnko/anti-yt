#!/bin/bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <channelID>"
  echo "  例: $0 @func_hs"
  exit 1
fi

CHANNEL_ID="$1"

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_API_KEY="${ADMIN_API_KEY:?ADMIN_API_KEY is required}"

RES=$(curl -s -X POST "${BASE_URL}/api/admin/channel/import-playlists" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(jq -n --arg c "$CHANNEL_ID" '{external_channel_id: $c}')")
echo "$RES" | jq . 2>/dev/null || echo "$RES"
