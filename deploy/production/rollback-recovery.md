# Rollback and recovery runbook (production-like reference)

Operator guide for **application rollback**, **schema migration policy**, and **explicit recovery verification** after a failed or rolled-back release. Complements [#41](../../docs/ops/health-readiness-migration.md) (health/readiness/migrate) and [#42](README.md) (runtime baseline).

## What can be rolled back safely

| Layer | Rollback | Data impact | Notes |
|-------|----------|-------------|-------|
| **edge / web / api / worker** images | Yes — redeploy previous image tag or `compose up` previous build | None when Postgres + object-store volumes unchanged | Stateless processes; preferred rollback surface |
| **migrate job** (already applied) | **No automatic undo in production** | Depends on migration | Use forward-fix migration instead (see below) |
| **postgres named volume** | Destructive — only for disaster reset | **All metadata lost** | Not an application rollback; requires backup restore |
| **minio named volume** | Destructive | **All objects lost** | Not an application rollback; requires backup restore |
| **mailpit** | Ephemeral | Verification emails only | Safe to recreate |

**Release smoke persistence** (`make prod-like-release-smoke`) tests the **application-service restart** boundary: `api`, `web`, and `worker` restart while `filvault_prod_postgres_data` and `filvault_prod_minio_data` stay mounted. It does **not** wipe DB/object volumes.

## Application rollback procedure

Use when a new release fails `/readyz`, breaks core flows, or release smoke fails **after** migrate already succeeded.

### 1. Stop routing traffic

- Remove or mark unhealthy the new replicas at the load balancer / edge.
- Confirm `GET /readyz` is not used for liveness-only restarts during rollback.

### 2. Roll back stateless services

From the repository root with `deploy/production/.env` present:

```bash
# Option A — previous Compose build (if images still tagged locally)
git checkout <previous-release-tag-or-commit>
make prod-like-up   # rebuilds api/web/worker from that commit

# Option B — force-recreate without rebuilding (previous image already tagged)
docker compose -f docker-compose.prod.yml --env-file deploy/production/.env \
  up -d --force-recreate --no-deps api web worker edge
```

**Do not** run `docker compose down -v` — that destroys named volumes.

### 3. Verify rollback health

```bash
make prod-like-rollback-verify
```

This runs explicit checks (readyz, edge reachability, optional release-smoke subset). Exit non-zero on failure.

### 4. Re-run release smoke (full gate)

```bash
make prod-like-release-smoke
```

Confirms authenticate → browse → upload → download → app restart → persistence.

## Schema migration policy

| Scenario | Policy |
|----------|--------|
| **Pending migration before deploy** | Run `migrate` once; require exit `0` before starting API/worker (Compose enforces this). |
| **Migration fails mid-release** | Do **not** start new API traffic. Fix migration or roll forward; investigate DB state. |
| **Rollback after successful migrate** | **Forward-fix preferred**: ship a new migration that reverses or compensates schema/data changes. |
| **`migrate -down` in production** | **Maintenance only** — may drop columns/tables; coordinate backup + downtime; never part of automated rollback. |

### Forward-fix example (operator maintenance)

```bash
# 1. Fix schema in a new migration under apps/api/migrations/
# 2. Apply explicitly:
docker compose -f docker-compose.prod.yml --env-file deploy/production/.env \
  run --rm migrate
# 3. Verify:
docker compose -f docker-compose.prod.yml --env-file deploy/production/.env \
  exec api wget -qO- http://127.0.0.1:8080/readyz
```

### Migration rollback (non-production / break-glass)

```bash
FILVAULT_DATABASE_URL='postgres://...' migrate -down
```

Only when a down migration exists and operators accept data loss risk. Document every use in the incident record.

## Recovery verification (explicit)

Recovery is **proven**, not assumed, when all commands below exit `0`:

| # | Verification | Command |
|---|--------------|---------|
| 1 | Schema gate | `migrate` container exit `0` on last deploy |
| 2 | Readiness | `curl -sk https://<host>:<port>/api/../readyz` or `compose exec api wget -qO- http://127.0.0.1:8080/readyz` → `"status":"ready"` |
| 3 | Liveness | `/healthz` → `200` |
| 4 | Core release path | `make prod-like-release-smoke` |
| 5 | Infra contract | `make prod-like-smoke-verify` (TLS, volumes, proxy headers) |

Automated bundle:

```bash
./deploy/production/rollback-verify.sh
```

## Distinguishing restart types

| Action | Services | Volumes | Use case |
|--------|----------|---------|----------|
| **App restart** | `api`, `web`, `worker` | Postgres + MinIO **kept** | Rolling deploy, crash recovery, release-smoke persistence |
| **Stack down** | All containers stopped | Volumes **kept** (default `compose down`) | Maintenance window |
| **Destructive reset** | `compose down -v` | Volumes **removed** | Lab reset only — **not** rollback |

## Incident checklist

1. Capture `compose ps`, `compose logs --tail=100 api migrate`, and `/readyz` body (no secrets).
2. Roll back **api/web/worker** to last known-good image.
3. Run `make prod-like-rollback-verify`.
4. If schema-related, choose forward-fix vs break-glass `migrate -down` with backup.
5. Run `make prod-like-release-smoke` before re-enabling traffic.
6. Record root cause and whether persistence boundary was involved.

Refs: #4 (epic), #41, #42, #43.
