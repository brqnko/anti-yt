#!/bin/bash
set -euo pipefail

if [ $# -lt 3 ]; then
  echo "Usage: $0 <channelID> <reason> <description...>"
  echo "  reason: unknown, education, technology, economy, politics, music"
  echo "  例: $0 @func_hs education このチャンネルは教育系で..."
  exit 1
fi

CHANNEL_ID="$1"
REASON="$2"
shift 2
DESCRIPTION="$*"

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_API_KEY="${ADMIN_API_KEY:?ADMIN_API_KEY is required}"

RES=$(curl -s -X POST "${BASE_URL}/api/admin/valuable" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(jq -n --arg c "$CHANNEL_ID" --arg r "$REASON" --arg d "$DESCRIPTION" \
    '{external_channel_id: $c, reason: $r, description: $d}')")
echo "$RES" | jq . 2>/dev/null || echo "$RES"
