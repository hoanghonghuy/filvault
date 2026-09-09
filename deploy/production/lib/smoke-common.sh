#!/usr/bin/env bash
# Shared helpers for production-like smoke scripts (#43).
# Secret-safe: never log tokens, passwords, invite codes, or presigned URLs with query params.
set -euo pipefail

SMOKE_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SMOKE_ROOT="$(cd "$SMOKE_SCRIPT_DIR/../.." && pwd)"

COMPOSE="${COMPOSE:-docker compose -f docker-compose.prod.yml --env-file deploy/production/.env}"
MAILPIT_API="${MAILPIT_API:-http://127.0.0.1:8025/api/v1}"

# Deterministic smoke email for login-or-register within a run (not a deployment secret).
SMOKE_USER_EMAIL="${SMOKE_USER_EMAIL:-release-smoke@filvault.local}"
SMOKE_USER_DISPLAY="${SMOKE_USER_DISPLAY:-Release Smoke}"
# SMOKE_USER_PASSWORD: inject via env or generate once per process in ensure_smoke_password() — never commit/log.
SMOKE_USER_PASSWORD="${SMOKE_USER_PASSWORD:-}"
SMOKE_PASSWORD_RESOLVED=false

SMOKE_STEP=""
SMOKE_FAIL_AT=""
API_LAST_CODE=""
API_LAST_BODY=""

smoke_log() {
  printf '[release-smoke] %s\n' "$*"
}

smoke_step() {
  SMOKE_STEP="$1"
  smoke_log "=== $SMOKE_STEP ==="
  if [[ -n "$SMOKE_FAIL_AT" && "$SMOKE_FAIL_AT" == "$SMOKE_STEP" ]]; then
    smoke_fail "forced failure at step '$SMOKE_STEP' (--fail-at)"
  fi
}

smoke_fail() {
  local msg="$1"
  printf '[release-smoke] FAIL step=%s: %s\n' "${SMOKE_STEP:-unknown}" "$msg" >&2
  if [[ -n "${SMOKE_STEP:-}" ]]; then
    printf '[release-smoke] hint: inspect service logs with: %s logs --tail=50 api web worker\n' "$COMPOSE" >&2
    printf '[release-smoke] hint: check readiness: %s exec api wget -qO- http://127.0.0.1:8080/readyz\n' "$COMPOSE" >&2
  fi
  exit 1
}

redact_json() {
  sed -E \
    -e 's/"accessToken"\s*:\s*"[^"]*"/"accessToken":"[redacted]"/g' \
    -e 's/"refreshToken"\s*:\s*"[^"]*"/"refreshToken":"[redacted]"/g' \
    -e 's/(uploadUrl|downloadUrl)"\s*:\s*"[^"]*\?[^"]*"/\1":"[redacted-presigned]"/g' \
    -e 's/(uploadUrl|downloadUrl)"\s*:\s*"[^"]{80,}"/\1":"[redacted-url]"/g'
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    smoke_fail "required command not found: $cmd"
  fi
}

ensure_smoke_password() {
  if $SMOKE_PASSWORD_RESOLVED; then
    return 0
  fi
  if [[ -n "$SMOKE_USER_PASSWORD" ]]; then
    SMOKE_PASSWORD_RESOLVED=true
    smoke_log "smoke password: operator-provided via SMOKE_USER_PASSWORD (value not logged)"
    return 0
  fi
  require_cmd openssl
  local generated=""
  if ! generated="$(openssl rand -hex 16 2>/dev/null)" || [[ -z "$generated" ]]; then
    smoke_fail "could not generate ephemeral smoke password (openssl rand failed)"
  fi
  SMOKE_USER_PASSWORD="$generated"
  unset generated
  SMOKE_PASSWORD_RESOLVED=true
  smoke_log "smoke password: ephemeral (generated in-memory for this run; not logged)"
}

load_prod_env() {
  local env_file="$SMOKE_ROOT/deploy/production/.env"
  if [[ ! -f "$env_file" ]]; then
    smoke_fail "missing $env_file — run ./deploy/production/generate-env.sh first"
  fi
  set -a
  # shellcheck disable=SC1090
  source "$env_file"
  set +a

  ORIGIN="https://${FILVAULT_PUBLIC_HOST:-filvault.local}:${FILVAULT_PUBLIC_PORT:-8443}"
  export ORIGIN
  export FILVAULT_INVITE_CODE
}

api_call() {
  local method="$1" path="$2" body="${3:-}" token="${4:-}"
  local tmp hdr
  tmp="$(mktemp)"
  hdr="$(mktemp)"
  local args=(-sk -X "$method" "$ORIGIN$path" -D "$hdr" -o "$tmp" -w '%{http_code}')
  args+=(-H 'Content-Type: application/json')
  if [[ -n "$token" ]]; then
    args+=(-H "Authorization: Bearer $token")
  fi
  if [[ -n "$body" ]]; then
    args+=(-d "$body")
  fi
  API_LAST_CODE="$(curl "${args[@]}" 2>/dev/null || echo "000")"
  API_LAST_BODY="$(cat "$tmp")"
  rm -f "$tmp" "$hdr"
}

assert_http() {
  local got="$1" want="$2" context="$3"
  if [[ "$got" != "$want" ]]; then
    local safe
    safe="$(printf '%s' "$API_LAST_BODY" | redact_json | head -c 500)"
    smoke_fail "$context: HTTP $got (expected $want) body=${safe}"
  fi
}

wait_readyz() {
  local label="${1:-readyz}"
  local i
  for i in $(seq 1 60); do
    if $COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/readyz 2>/dev/null | grep -q '"status":"ready"'; then
      smoke_log "$label: API ready (${i} attempts)"
      return 0
    fi
    sleep 2
  done
  smoke_fail "$label: API /readyz did not return ready within 120s"
}

mailpit_verification_code() {
  local email="$1"
  local i msg_id body digits
  for i in $(seq 1 30); do
    msg_id="$(
      curl -sf "$MAILPIT_API/messages" 2>/dev/null \
        | jq -r --arg e "$email" '
            .messages[]?
            | select((.To[]?.Address // "") == $e)
            | .ID
          ' 2>/dev/null | head -1
    )"
    if [[ -n "$msg_id" && "$msg_id" != "null" ]]; then
      body="$(curl -sf "$MAILPIT_API/message/$msg_id" | jq -r '.Text // .HTML' 2>/dev/null || true)"
      digits="$(printf '%s' "$body" | grep -oE '[0-9]{6}' | head -1 || true)"
      if [[ -n "$digits" ]]; then
        printf '%s' "$digits"
        return 0
      fi
    fi
    sleep 1
  done
  smoke_fail "verification code not found in Mailpit for $email (UI: http://127.0.0.1:8025)"
}

restart_app_services() {
  smoke_log "restarting stateless app services (api, web, worker) — postgres/minio named volumes untouched"
  $COMPOSE restart api worker web
  wait_readyz "post-restart readyz"
}

ensure_smoke_user() {
  ensure_smoke_password
  local login_body
  login_body="$(printf '{"email":"%s","password":"%s"}' "$SMOKE_USER_EMAIL" "$SMOKE_USER_PASSWORD")"
  api_call POST /api/v1/auth/login "$login_body"
  if [[ "$API_LAST_CODE" == "200" ]]; then
    smoke_log "existing smoke user authenticated"
    return 0
  fi

  local reg_body
  reg_body="$(jq -nc \
    --arg email "$SMOKE_USER_EMAIL" \
    --arg password "$SMOKE_USER_PASSWORD" \
    --arg name "$SMOKE_USER_DISPLAY" \
    --arg invite "$FILVAULT_INVITE_CODE" \
    '{email:$email,password:$password,displayName:$name,inviteCode:$invite}')"
  api_call POST /api/v1/auth/register "$reg_body"
  if [[ "$API_LAST_CODE" == "409" ]]; then
    smoke_fail "smoke user $SMOKE_USER_EMAIL exists but login failed — export SMOKE_USER_PASSWORD for that account, or remove the user / reset postgres volume before re-running with an ephemeral password"
  fi
  assert_http "$API_LAST_CODE" "201" "register smoke user"

  api_call POST /api/v1/auth/resend-verification "$(jq -nc --arg e "$SMOKE_USER_EMAIL" '{email:$e}')"
  assert_http "$API_LAST_CODE" "204" "resend-verification"

  local verify_code verify_body
  verify_code="$(mailpit_verification_code "$SMOKE_USER_EMAIL")"
  verify_body="$(jq -nc --arg e "$SMOKE_USER_EMAIL" --arg c "$verify_code" '{email:$e,code:$c}')"
  api_call POST /api/v1/auth/verify-email "$verify_body"
  assert_http "$API_LAST_CODE" "204" "verify-email"
  smoke_log "smoke user registered and verified"
}

smoke_login() {
  ensure_smoke_password
  local body
  body="$(printf '{"email":"%s","password":"%s"}' "$SMOKE_USER_EMAIL" "$SMOKE_USER_PASSWORD")"
  api_call POST /api/v1/auth/login "$body"
  assert_http "$API_LAST_CODE" "200" "login"
}

smoke_access_token() {
  jq -r '.accessToken' <<<"$API_LAST_BODY"
}

smoke_refresh_token() {
  jq -r '.refreshToken' <<<"$API_LAST_BODY"
}

smoke_logout() {
  local token="$1" refresh="$2"
  api_call POST /api/v1/auth/logout "$(jq -nc --arg r "$refresh" '{refreshToken:$r}')" "$token"
}
