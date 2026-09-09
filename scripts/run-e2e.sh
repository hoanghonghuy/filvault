#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export FILVAULT_POSTGRES_PASSWORD="${FILVAULT_POSTGRES_PASSWORD:-replace-with-local-postgres-password}"
export FILVAULT_S3_ACCESS_KEY="${FILVAULT_S3_ACCESS_KEY:-filvault}"
export FILVAULT_S3_SECRET_KEY="${FILVAULT_S3_SECRET_KEY:-replace-with-local-s3-secret-key}"
export FILVAULT_INVITE_CODE="${FILVAULT_INVITE_CODE:-replace-with-local-invite-code}"
export FILVAULT_JWT_SECRET="${FILVAULT_JWT_SECRET:-replace-with-local-jwt-secret-min-32-chars}"
export FILVAULT_SEED_EMAIL="${FILVAULT_SEED_EMAIL:-dev@filvault.com}"
export FILVAULT_SEED_PASSWORD="${FILVAULT_SEED_PASSWORD:-replace-with-local-seed-password}"
export FILVAULT_SEED_DISPLAY_NAME="${FILVAULT_SEED_DISPLAY_NAME:-Filvault Dev}"
export DSN="${DSN:-postgres://filvault:${FILVAULT_POSTGRES_PASSWORD}@127.0.0.1:5435/filvault?sslmode=disable}"

export E2E_EMAIL="${E2E_EMAIL:-$FILVAULT_SEED_EMAIL}"
export E2E_PASSWORD="${E2E_PASSWORD:-$FILVAULT_SEED_PASSWORD}"
export E2E_BASE_URL="${E2E_BASE_URL:-http://127.0.0.1:5173}"
export E2E_API_URL="${E2E_API_URL:-http://127.0.0.1:8080}"
export E2E_RUN_ID="${E2E_RUN_ID:-$(date +%s)}"

API_PID=""

cleanup() {
  if [[ -n "$API_PID" ]] && kill -0 "$API_PID" 2>/dev/null; then
    kill "$API_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

echo "Starting Postgres and MinIO..."
make compose-up

echo "Running migrations and seeding dev user..."
make migrate seed

echo "Starting API on ${E2E_API_URL}..."
(
  cd apps/api
  FILVAULT_DATABASE_URL="$DSN" \
  FILVAULT_METADATA_STORE=postgres \
  FILVAULT_OBJECT_STORE=s3 \
  FILVAULT_S3_ENDPOINT=http://127.0.0.1:9002 \
  FILVAULT_S3_PUBLIC_ENDPOINT=http://127.0.0.1:9002 \
  FILVAULT_S3_BUCKET=filvault \
  FILVAULT_S3_REGION=us-east-1 \
  FILVAULT_S3_ACCESS_KEY="$FILVAULT_S3_ACCESS_KEY" \
  FILVAULT_S3_SECRET_KEY="$FILVAULT_S3_SECRET_KEY" \
  FILVAULT_JWT_SECRET="$FILVAULT_JWT_SECRET" \
  FILVAULT_INVITE_CODE="$FILVAULT_INVITE_CODE" \
  FILVAULT_HTTP_ADDR=:8080 \
  FILVAULT_MAILER=console \
  FILVAULT_CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173,http://localhost:5173 \
  go run ./cmd/api
) &
API_PID=$!

echo "Waiting for API readiness..."
for _ in $(seq 1 60); do
  if curl -sf "${E2E_API_URL}/healthz" >/dev/null; then
    echo "API ready"
    break
  fi
  sleep 1
done

if ! curl -sf "${E2E_API_URL}/healthz" >/dev/null; then
  echo "API did not become ready at ${E2E_API_URL}" >&2
  exit 1
fi

echo "Running Playwright E2E smoke..."
(
  cd apps/web
  npm run test:e2e
)
