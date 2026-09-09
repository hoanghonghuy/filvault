# Production-like runtime baseline

Vendor-neutral reference deployment for Filvault API + Web + required services (#42). This stack is suitable for **smoke testing** and documents the contract real operators must satisfy; it is not AWS-specific and does not embed cloud business logic.

Health, readiness, and migration semantics come from [#41](../ops/health-readiness-migration.md) — use `/healthz` (liveness), `/readyz` (traffic), and an explicit `migrate` step before API traffic.

## Architecture

```text
Browser ──HTTPS──► edge (Caddy, TLS termination)
                        │
                        ▼
                   web (nginx: SPA + /api + /filvault proxies)
                        │
            ┌───────────┼───────────┐
            ▼           ▼           ▼
          api         minio      (livekit, optional profile)
            │
            ├── postgres (TLS, named volume)
            └── worker
```

| Component | Role | Persistence |
|-----------|------|-------------|
| **edge** | TLS termination, public origin | Stateless |
| **web** | Static UI, reverse-proxy `/api/` and `/filvault/` | Stateless |
| **api** | `FILVAULT_ENV=production`, no dev seed | Stateless |
| **worker** | Background jobs | Stateless |
| **postgres** | Metadata (`sslmode=require`) | Named volume `filvault_prod_postgres_data` |
| **minio** | S3-compatible object store | Named volume `filvault_prod_minio_data` |
| **mailpit** | SMTP sink for smoke (replace with real SMTP in production) | Ephemeral |
| **migrate** | One-shot schema gate (exit 0 required) | N/A |

Container filesystems are **not** durable. Postgres and MinIO data survive only via the named volumes above (or external managed services in real deployments).

## Prerequisites

- Docker Engine with Compose v2
- `openssl` (generate Postgres TLS + smoke secrets)
- Map the public host to the machine running Compose:

  ```bash
  echo '127.0.0.1 filvault.local' | sudo tee -a /etc/hosts
  ```

## Bring-up (clean operator)

From the repository root:

```bash
chmod +x deploy/production/postgres/generate-certs.sh deploy/production/generate-env.sh

# 1. Postgres TLS (required for FILVAULT_ENV=production sslmode=require)
#    Certs are owned as uid/gid 70 — the postgres user in postgres:16-alpine (Alpine standard; not 999).
./deploy/production/postgres/generate-certs.sh

# 2. Generate gitignored deploy/production/.env with random secrets
./deploy/production/generate-env.sh

# 3. Build and start the stack (migrate runs before api/worker)
make prod-like-up
```

Open **https://filvault.local:8443** (default). Mailpit UI for verification emails: **http://127.0.0.1:8025**.

Verify probes:

```bash
curl -sk https://filvault.local:8443/api/v1/../healthz  # via edge → web → api path; use direct:
docker compose -f docker-compose.prod.yml --env-file deploy/production/.env exec api wget -qO- http://127.0.0.1:8080/healthz
docker compose -f docker-compose.prod.yml --env-file deploy/production/.env exec api wget -qO- http://127.0.0.1:8080/readyz
```

Stop:

```bash
make prod-like-down
```

## Release smoke (#43)

One documented command verifies the **core release path** against this stack:

```bash
make prod-like-release-smoke
```

Coverage: edge reachable → register/login (deterministic smoke user) → browse Files → upload fixture → download checksum → **restart `api`/`web`/`worker` only** (Postgres/MinIO volumes intact) → confirm metadata + object persist → logout/cleanup.

| Command | Purpose |
|---------|---------|
| `make prod-like-smoke-verify` | Infra/probes smoke (#42): TLS, migrate gate, volumes, `/healthz`/`/readyz` |
| `make prod-like-release-smoke` | Full release gate (#43) |
| `make prod-like-rollback-verify` | Post-rollback recovery checks |

Forced failure (prove non-zero exit + diagnostics):

```bash
./deploy/production/release-smoke.sh --fail-at=download
```

Static self-check (no Docker):

```bash
./deploy/production/release-smoke.sh --self-check
```

Rollback/recovery policy and explicit verification steps: [`rollback-recovery.md`](rollback-recovery.md).

## Configuration contract

All secrets and URLs are injected via `deploy/production/.env` (or your secret manager). **`FILVAULT_ENV=production` rejects documented placeholders** (see `apps/api/internal/platform/config/validate.go`).

| Variable | Required | Notes |
|----------|----------|-------|
| `FILVAULT_ENV` | yes | Must be `production` |
| `FILVAULT_DATABASE_URL` | yes | `sslmode=require\|verify-ca\|verify-full`; host ≠ localhost |
| `FILVAULT_JWT_SECRET` | yes | ≥ 32 chars, unique |
| `FILVAULT_INVITE_CODE` | yes | ≥ 8 chars, unique |
| `FILVAULT_S3_*` | yes | Internal endpoint + **public** HTTPS path for presigned URLs |
| `FILVAULT_CORS_ALLOWED_ORIGINS` | yes | Public HTTPS origin(s) only |
| `FILVAULT_MAILER` | yes | Must be `smtp` (not `console`) |
| `FILVAULT_SMTP_*` | yes | Real SMTP in production; Mailpit OK for this reference stack |
| `FILVAULT_TRUSTED_PROXIES` | recommended | CIDRs of reverse-proxy hops (Docker private ranges in Compose) |
| `FILVAULT_SEED_DEV_USER` | must be false | Dev seed disabled |
| `FILVAULT_LIVEKIT_*` | optional | Omit both key and secret to disable calls safely |

Public URLs (defaults):

- Web/API origin: `https://${FILVAULT_PUBLIC_HOST}:${FILVAULT_PUBLIC_PORT}` (default `https://filvault.local:8443`)
- Object uploads via same origin: `${PUBLIC_ORIGIN}/filvault/...` (web nginx → MinIO)

Template without working secrets: [`deploy/production/.env.example`](.env.example).

## TLS and reverse proxy

- **TLS terminates at `edge` (Caddy)** using `tls internal` (self-signed) for smoke. Production operators replace this with ACME or managed certificates on their edge.
- **Upstream headers**: Caddy and web nginx set `X-Forwarded-Proto` and `X-Forwarded-For`. The API honors these only from CIDRs in `FILVAULT_TRUSTED_PROXIES` (Gin trusted proxies).
- **Do not expose API or MinIO ports** on the host in this reference stack; only `edge:443` (mapped to host `8443`) and Mailpit `8025` are published.

## LiveKit (optional)

Voice/video is **disabled by default** (empty LiveKit keys). To enable in the reference stack:

1. Add keys to `deploy/production/.env`.
2. Start with profile:  
   `docker compose -f docker-compose.prod.yml --env-file deploy/production/.env --profile livekit up -d`

Set `FILVAULT_LIVEKIT_PUBLIC_URL` to a `wss://` URL reachable by browsers (not localhost in production validation).

## Release gates (coordinate with #41)

1. Run `migrate` once; require exit `0`.
2. Roll `api` / `worker` / `web` / `edge`.
3. Wait for `GET /readyz` → `200` before declaring healthy.
4. Use `GET /healthz` only for liveness restarts.

Compose enforces step 1 via `depends_on: migrate: condition: service_completed_successfully`.

## External production

Replace MinIO/Postgres/Mailpit with managed services by pointing the same env vars at external endpoints. Keep:

- TLS on Postgres
- HTTPS public origin and CORS alignment
- Real SMTP
- Explicit migrate job before traffic
- `/readyz` upstream check on the load balancer

AWS (or other clouds) may be documented later as deployment targets — not required for this baseline.

Refs: #4 (epic), #41 (health/readiness), #42 (this baseline), #43 (release smoke + rollback runbook).
