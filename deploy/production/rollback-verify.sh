#!/usr/bin/env bash
# Explicit recovery verification after rollback or forward-fix (#43).
# Exit 0 only when health gates pass; non-zero with actionable diagnostics otherwise.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

# shellcheck source=deploy/production/lib/smoke-common.sh
source "$ROOT/deploy/production/lib/smoke-common.sh"

SELF_CHECK=false
RUN_RELEASE_SMOKE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --self-check) SELF_CHECK=true ;;
    --with-release-smoke) RUN_RELEASE_SMOKE=true ;;
    -h|--help)
      echo "Usage: $0 [--self-check] [--with-release-smoke]"
      exit 0
      ;;
    *) smoke_fail "unknown argument: $1" ;;
  esac
  shift
done

if $SELF_CHECK; then
  bash -n "$0"
  smoke_log "rollback-verify self-check OK"
  exit 0
fi

require_cmd curl
require_cmd jq
require_cmd docker

load_prod_env

smoke_step "migrate-gate"
MIGRATE_STATE="$($COMPOSE ps -a migrate --format '{{.State}} {{.ExitCode}}' 2>/dev/null || true)"
if [[ ! "$MIGRATE_STATE" =~ exited.*0 ]]; then
  smoke_fail "migrate container not exited 0 (state: ${MIGRATE_STATE:-unknown})"
fi
smoke_log "migrate gate OK ($MIGRATE_STATE)"

smoke_step "readyz"
wait_readyz "rollback-verify readyz"
READY_BODY="$($COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/readyz 2>/dev/null || true)"
if ! jq -e '.checks.database == "ok" and .checks.schema == "ok" and .checks.object_store == "ok"' <<<"$READY_BODY" >/dev/null 2>&1; then
  smoke_fail "readyz checks not all ok: $(printf '%s' "$READY_BODY" | head -c 300)"
fi
smoke_log "readyz checks: database, schema, object_store OK"

smoke_step "healthz"
HEALTH="$($COMPOSE exec -T api wget -qO- http://127.0.0.1:8080/healthz 2>/dev/null || true)"
if ! grep -q '"status":"ok"' <<<"$HEALTH"; then
  smoke_fail "healthz not ok: $HEALTH"
fi
smoke_log "healthz OK"

smoke_step "edge-reachable"
code="$(http_code "$ORIGIN/")"
if [[ "$code" != "200" ]]; then
  smoke_fail "edge $ORIGIN not reachable (HTTP $code)"
fi
smoke_log "edge reachable (HTTP $code)"

smoke_step "volume-boundary"
for vol in filvault_prod_postgres_data filvault_prod_minio_data; do
  if ! docker volume inspect "$vol" >/dev/null 2>&1; then
    smoke_fail "named volume missing: $vol"
  fi
  smoke_log "volume present: $vol"
done

if $RUN_RELEASE_SMOKE; then
  smoke_step "release-smoke"
  "$ROOT/deploy/production/release-smoke.sh" --skip-cleanup
fi

smoke_log "ROLLBACK VERIFY PASSED"
exit 0
