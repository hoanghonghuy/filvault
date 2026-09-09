#!/usr/bin/env bash
# Release smoke: production-like core path + app-service restart persistence (#43).
#
# Usage:
#   ./deploy/production/release-smoke.sh              # full gate (stack must be up)
#   ./deploy/production/release-smoke.sh --self-check # static validation only
#   ./deploy/production/release-smoke.sh --fail-at=download  # prove non-zero exit
#
# Prerequisites: make prod-like-up (or equivalent), jq, curl, docker compose.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

# shellcheck source=deploy/production/lib/smoke-common.sh
source "$ROOT/deploy/production/lib/smoke-common.sh"

SELF_CHECK=false
SKIP_CLEANUP=false

usage() {
  cat <<'EOF'
Filvault production-like release smoke (#43).

  ./deploy/production/release-smoke.sh [--self-check] [--fail-at=STEP] [--skip-cleanup]

Steps (--fail-at values): reachable, authenticate, browse, upload, verify, restart, persist, cleanup

Environment:
  SMOKE_USER_EMAIL, SMOKE_USER_PASSWORD  deterministic test identity (defaults in smoke-common.sh)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --self-check) SELF_CHECK=true ;;
    --skip-cleanup) SKIP_CLEANUP=true ;;
    --fail-at=*) SMOKE_FAIL_AT="${1#*=}" ;;
    -h|--help) usage; exit 0 ;;
    *) smoke_fail "unknown argument: $1" ;;
  esac
  shift
done

if $SELF_CHECK; then
  smoke_log "self-check: bash syntax"
  bash -n "$ROOT/deploy/production/lib/smoke-common.sh"
  bash -n "$ROOT/deploy/production/release-smoke.sh"
  bash -n "$ROOT/deploy/production/rollback-verify.sh"
  require_cmd jq
  require_cmd curl
  [[ -f "$ROOT/deploy/production/fixtures/release-smoke.txt" ]] || smoke_fail "missing fixture"
  grep -q 'release-smoke-persistence-v1' "$ROOT/deploy/production/fixtures/release-smoke.txt" || smoke_fail "fixture marker missing"
  smoke_log "self-check OK"
  exit 0
fi

require_cmd jq
require_cmd curl
require_cmd docker
require_cmd sha256sum

load_prod_env

FIXTURE="$ROOT/deploy/production/fixtures/release-smoke.txt"
FILE_NAME="release-smoke-$(date -u +%Y%m%dT%H%M%SZ).txt"
FOLDER_NAME="release-smoke-$(date -u +%Y%m%d)"
FILE_SIZE="$(wc -c <"$FIXTURE" | tr -d ' ')"
FIXTURE_SHA="$(sha256sum "$FIXTURE" | awk '{print $1}')"

TOKEN=""
REFRESH=""
FILE_ID=""
FOLDER_ID=""

cleanup() {
  if $SKIP_CLEANUP || [[ -z "$TOKEN" ]]; then
    return 0
  fi
  smoke_step "cleanup"
  if [[ -n "$FILE_ID" ]]; then
    api_call DELETE "/api/v1/files/$FILE_ID" "" "$TOKEN"
    smoke_log "file $FILE_ID soft-deleted (HTTP $API_LAST_CODE)"
  fi
  if [[ -n "$FOLDER_ID" ]]; then
    api_call DELETE "/api/v1/folders/$FOLDER_ID" "" "$TOKEN"
    smoke_log "folder $FOLDER_ID deleted (HTTP $API_LAST_CODE)"
  fi
  if [[ -n "$REFRESH" ]]; then
    smoke_logout "$TOKEN" "$REFRESH"
    smoke_log "logout complete (HTTP $API_LAST_CODE)"
  fi
}
trap cleanup EXIT

smoke_step "reachable"
edge_code="$(http_code "$ORIGIN/")"
if [[ "$edge_code" != "200" ]]; then
  smoke_fail "edge not reachable at $ORIGIN (HTTP $edge_code) — run make prod-like-up"
fi
wait_readyz "pre-smoke readyz"
smoke_log "edge reachable; API ready"

smoke_step "authenticate"
ensure_smoke_user
smoke_login
TOKEN="$(smoke_access_token)"
REFRESH="$(smoke_refresh_token)"
api_call GET /api/v1/users/me "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "users/me"
smoke_log "authenticated as $(jq -r '.email' <<<"$API_LAST_BODY")"

smoke_step "browse"
api_call GET /api/v1/browser "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "browser root"
smoke_log "browser root OK"

smoke_step "upload"
api_call POST /api/v1/folders "$(jq -nc --arg n "$FOLDER_NAME" '{name:$n}')" "$TOKEN"
assert_http "$API_LAST_CODE" "201" "create folder"
FOLDER_ID="$(jq -r '.id' <<<"$API_LAST_BODY")"

session_body="$(jq -nc \
  --arg name "$FILE_NAME" \
  --arg folder "$FOLDER_ID" \
  --argjson size "$FILE_SIZE" \
  '{name:$name,size:$size,contentType:"text/plain",folderId:$folder}')"
api_call POST /api/v1/files/upload-sessions "$session_body" "$TOKEN"
assert_http "$API_LAST_CODE" "201" "upload session"
FILE_ID="$(jq -r '.fileId' <<<"$API_LAST_BODY")"
UPLOAD_URL="$(jq -r '.uploadUrl' <<<"$API_LAST_BODY")"

put_code="$(curl -sk -o /dev/null -w '%{http_code}' -X PUT "$UPLOAD_URL" \
  -H 'Content-Type: text/plain' --data-binary @"$FIXTURE")"
if [[ "$put_code" != "200" ]]; then
  smoke_fail "S3 PUT failed HTTP $put_code (presigned upload)"
fi

api_call POST "/api/v1/files/$FILE_ID/complete" "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "upload complete"
smoke_log "uploaded file id=$FILE_ID size=$FILE_SIZE sha256=$FIXTURE_SHA"

smoke_step "verify"
api_call GET "/api/v1/browser?folderId=$FOLDER_ID" "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "browser after upload"
if ! jq -e --arg n "$FILE_NAME" '.files[]? | select(.name == $n)' <<<"$API_LAST_BODY" >/dev/null; then
  smoke_fail "uploaded file not listed in browser"
fi

api_call GET "/api/v1/files/$FILE_ID/download" "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "download URL"
DOWNLOAD_URL="$(jq -r '.downloadUrl' <<<"$API_LAST_BODY")"
TMP_DL="$(mktemp)"
dl_code="$(curl -sk -o "$TMP_DL" -w '%{http_code}' "$DOWNLOAD_URL")"
if [[ "$dl_code" != "200" ]]; then
  smoke_fail "download GET failed HTTP $dl_code"
fi
DL_SHA="$(sha256sum "$TMP_DL" | awk '{print $1}')"
rm -f "$TMP_DL"
if [[ "$DL_SHA" != "$FIXTURE_SHA" ]]; then
  smoke_fail "download checksum mismatch (got $DL_SHA want $FIXTURE_SHA)"
fi
smoke_log "download verified (sha256 match)"

smoke_step "restart"
restart_app_services

smoke_step "persist"
smoke_login
TOKEN="$(smoke_access_token)"
api_call GET "/api/v1/files/$FILE_ID" "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "file metadata after restart"
if [[ "$(jq -r '.name' <<<"$API_LAST_BODY")" != "$FILE_NAME" ]]; then
  smoke_fail "file name mismatch after restart"
fi

api_call GET "/api/v1/browser?folderId=$FOLDER_ID" "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "browser after restart"
if ! jq -e --arg id "$FILE_ID" '.files[]? | select(.id == $id)' <<<"$API_LAST_BODY" >/dev/null; then
  smoke_fail "file missing from browser after app-service restart"
fi

api_call GET "/api/v1/files/$FILE_ID/download" "" "$TOKEN"
assert_http "$API_LAST_CODE" "200" "download after restart"
DOWNLOAD_URL="$(jq -r '.downloadUrl' <<<"$API_LAST_BODY")"
TMP_DL="$(mktemp)"
curl -sk -o "$TMP_DL" "$DOWNLOAD_URL"
DL_SHA="$(sha256sum "$TMP_DL" | awk '{print $1}')"
rm -f "$TMP_DL"
if [[ "$DL_SHA" != "$FIXTURE_SHA" ]]; then
  smoke_fail "download checksum mismatch after restart"
fi
smoke_log "persistence OK — metadata and object survived api/web/worker restart (DB/MinIO volumes intact)"

smoke_log "RELEASE SMOKE PASSED"
exit 0
