# Health, readiness, and migration release gates

Filvault exposes two HTTP probes on the API process. Operators and deployment automation should use them as follows.

## Endpoints

| Endpoint | Purpose | Depends on external services? |
|----------|---------|-------------------------------|
| `GET /healthz` | **Liveness** — process is running and the HTTP server can respond | No |
| `GET /readyz` | **Readiness** — safe to route authenticated core traffic | Yes (Postgres, schema, object store) |

### Liveness (`/healthz`)

- Returns `200` with `{"status":"ok"}` when the API process is alive.
- Does **not** ping Postgres, object storage, or any other dependency.
- Use for container/process liveness checks and crash-loop detection.
- A failing liveness probe means restart the process; it does not mean “wait for dependencies”.

### Readiness (`/readyz`)

- Returns `200` with `{"status":"ready","checks":{...}}` only when all mandatory dependencies are usable.
- Returns `503` with `{"status":"not_ready","checks":{...}}` when any check fails.
- Each check is bounded (2s timeout) and reports only `ok` or `failed` — no secrets or connection strings in the response.
- Checks performed:
  - `database` — Postgres connectivity (`Ping`)
  - `schema` — every embedded migration version is recorded in `schema_migrations`
  - `object_store` — configured store is reachable (in-memory always passes; S3-compatible stores use `HeadBucket`)

Use `/readyz` for:

- Load balancer / reverse-proxy upstream health (see also #42)
- Kubernetes `readinessProbe` / Compose “ready for traffic” gates
- Post-deploy smoke: rollout succeeds only when `/readyz` is `200`

Do **not** use `/healthz` to decide whether traffic should be routed.

## Migration release gate

Schema changes are applied by an **explicit, gated step** before API replicas accept traffic. Migrations are **not** run automatically on every API pod startup.

### Command

```bash
FILVAULT_DATABASE_URL='postgres://...' migrate
```

- Exit `0` — all pending migrations applied and `SchemaReady` verified.
- Exit `1` — connection, migration, or schema-readiness failure (stderr message, no secrets).

Rollback (non-production / operator maintenance):

```bash
FILVAULT_DATABASE_URL='postgres://...' migrate -down
```

### Recommended rollout order

1. Run `migrate` once (job, init container, or CI pre-deploy step) and wait for exit `0`.
2. Start or roll API replicas.
3. Wait until `GET /readyz` returns `200` on each instance before marking the release healthy or shifting traffic.
4. Use `GET /healthz` only for liveness restarts after traffic is already routed.

`docker-compose.yml` models this: the `migrate` service completes successfully before `api` and `worker` start.

### Concurrency safety

`migrate` acquires a Postgres advisory lock (`filvault:migrations`) so concurrent invocations serialize. Still run migrations as a single gated release step rather than from every replica.

## Container health checks

The API image defines a **liveness** `HEALTHCHECK` against `/healthz` (no credentials in the probe).

For traffic routing, configure your orchestrator or reverse proxy to probe `/readyz` separately.

## Local development

```bash
make migrate    # explicit schema gate
make run-api    # migrate + seed + API (dev only)
curl -s localhost:8080/healthz
curl -s localhost:8080/readyz
```

Refs: #4 (deployment epic), #41 (this contract), #42 (runtime / reverse-proxy integration).
