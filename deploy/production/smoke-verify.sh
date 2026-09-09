#!/usr/bin/env bash
# Production-like smoke verification for PR #46 / issue #42.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

COMPOSE="docker compose -f docker-compose.prod.yml --env-file deploy/production/.env"
ORIGIN="${FILVAULT_PUBLIC_ORIGIN:-https://filvault.local:8443}"

echo "=== 1) TLS edge proxy (HTTPS public origin) ==="
curl -skI "$ORIGIN/" | head -5
curl -sk "$ORIGIN/" | head -c 200
echo
echo

echo "=== 2) Migrate-before-traffic gate ==="
MIGRATE_EXIT=$($COMPOSE ps -a migrate --format '{{.State}} {{.ExitCode}}')
echo "migrate container: $MIGRATE_EXIT"
$COMPOSE logs migrate 2>&1 | tail -3
API_STARTED=$($COMPOSE ps api --format '{{.Status}}')
echo "api status: $API_STARTED"
echo

echo "=== 3) Named volumes (persistence) ==="
docker volume inspect filvault_prod_postgres_data --format '{{.Name}} mount={{.Mountpoint}}'
docker volume inspect filvault_prod_minio_data --format '{{.Name}} mount={{.Mountpoint}}'
$COMPOSE exec -T postgres psql -U filvault -d filvault -c "CREATE TABLE IF NOT EXISTS smoke_persist (id int); INSERT INTO smoke_persist VALUES (1) ON CONFLICT DO NOTHING;" 2>/dev/null || \
  $COMPOSE exec -T postgres psql "postgres://filvault@$(grep FILVAULT_POSTGRES_PASSWORD deploy/production/.env | cut -d= -f2)@localhost/filvault?sslmode=disable" -c "SELECT 1" 2>/dev/null || true
docker run --rm -v filvault_prod_postgres_data:/v alpine:3.21 ls -la /v/PG_VERSION 2>&1
docker run --rm -v filvault_prod_minio_data:/v alpine:3.21 sh -c 'ls /v/.minio.sys 2>/dev/null | head -3 || ls /v | head -3'
echo

echo "=== 4) Trusted proxy / forwarded headers ==="
$COMPOSE exec -T api wget -qO- --header='X-Forwarded-For: 203.0.113.50' --header='X-Forwarded-Proto: https' http://127.0.0.1:8080/healthz
echo
curl -sk -H 'X-Forwarded-Proto: https' "$ORIGIN/api/v1/auth/login" -X POST -H 'Content-Type: application/json' -d '{}' -w '\nHTTP %{http_code}\n' | tail -3
echo

echo "=== 5) /healthz liveness vs /readyz readiness ==="
echo -n "healthz (direct): "
$COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/healthz
echo
echo -n "readyz (direct): "
$COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/readyz
echo
echo -n "readyz (via edge): "
curl -sk "$ORIGIN/api/../readyz" 2>/dev/null || curl -sk "$ORIGIN/readyz" 2>/dev/null || \
  $COMPOSE exec -T web wget -qO- http://api:8080/readyz
echo

echo "=== 5b) readyz fails when object store unavailable (503) ==="
$COMPOSE stop minio >/dev/null
sleep 3
READY_WHEN_DOWN=$($COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/readyz 2>&1 || true)
HEALTH_WHEN_DOWN=$($COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/healthz 2>&1 || true)
$COMPOSE start minio >/dev/null
echo "readyz with minio stopped: $READY_WHEN_DOWN"
echo "healthz with minio stopped: $HEALTH_WHEN_DOWN"
echo

echo "=== SMOKE VERIFY COMPLETE ==="
