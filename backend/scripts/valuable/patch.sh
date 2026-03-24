#!/bin/bash
set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: $0 <channelID> [reason] [description...]"
  echo "  reason: unknown, education, technology, economy, politics, music"
  echo "  例: $0 @func_hs education 更新後の説明文"
  exit 1
fi

CHANNEL_ID="$1"
REASON="${2:-}"
shift 2 2>/dev/null || shift $#
DESCRIPTION="$*"

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_API_KEY="${ADMIN_API_KEY:?ADMIN_API_KEY is required}"

ARGS=(--arg c "$CHANNEL_ID")
FIELDS='{external_channel_id: $c'
[ -n "$REASON" ] && ARGS+=(--arg r "$REASON") && FIELDS="$FIELDS, reason: \$r"
[ -n "$DESCRIPTION" ] && ARGS+=(--arg d "$DESCRIPTION") && FIELDS="$FIELDS, description: \$d"
FIELDS="$FIELDS}"

RES=$(curl -s -X PATCH "${BASE_URL}/admin/valuable" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(jq -n "${ARGS[@]}" "$FIELDS")")
echo "$RES" | jq . 2>/dev/null || echo "$RES"
