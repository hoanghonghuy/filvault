# Secret lifecycle audit (issue #3)

Audit date: 2026-09-09  
Scope: current tree and full Git history (`gitleaks` + manual review of `.env*`, Docker/Compose, `deploy/`, Makefile, GitHub Actions, docs, test fixtures).

This document classifies findings **by category only**. Secret values are intentionally omitted because the repository is public and historical commits remain readable.

## Method

1. `gitleaks detect --source . --config .gitleaks.toml --log-opts="--all"` across all commits.
2. Targeted `git log -S` searches for historically tracked credential keys (`FILVAULT_JWT_SECRET`, `POSTGRES_PASSWORD`, `MINIO_ROOT_PASSWORD`, LiveKit `keys:` block, Makefile inline env).
3. Manual review of current tree files listed in issue scope.

## Category A — Documented local placeholders (current tree, safe)

| Location | Notes |
|---|---|
| `.env.example` | All values use `replace-with-local-…` placeholders; copy to gitignored `.env` before local runs. |
| `Makefile` | Fallback `FILVAULT_POSTGRES_PASSWORD ?= replace-with-local-postgres-password` when `.env` is absent; loads `.env` when present. |
| `apps/api/internal/platform/config/validate.go` | Blocklist of placeholder/historical literals used for production fail-closed checks (not runtime secrets). |
| Test fixtures | Handler/integration tests use synthetic literals such as `test-jwt-secret-not-for-prod`; not used outside tests. |

**Action:** Keep placeholders clearly fake; do not copy into production configuration.

## Category B — Historical local-stack defaults (removed from current tree)

Found in Git history before commit `dc4441f` (2026-09-09). These were **local Docker/MinIO/Postgres/LiveKit dev defaults**, not cloud-provider credentials.

| Surface | What was committed | Current state |
|---|---|---|
| `docker-compose.yml` | Hardcoded Postgres password default, MinIO root user/password, JWT/invite/seed/LiveKit env literals | Replaced with `${ENV}` interpolation; requires `.env` or CI/test env injection |
| `deploy/livekit.yaml` | LiveKit `keys:` block with demo key pair | Keys removed; supplied via `LIVEKIT_KEYS` env in Compose |
| `Makefile` | Inline `dev-jwt-secret`, `filvault`/`filvaultsecret`, `dev-invite` for host `dev-api` | Removed; inherits from `.env` |
| `README.md` / status docs | Documented dev login literals | Updated to reference `.env` variables |
| `apps/api/internal/devseed` | Fallback seed password/invite constants | Removed; seed command requires `FILVAULT_SEED_PASSWORD` and `FILVAULT_INVITE_CODE` |

**Action:** Treat as exposed in public history. Do **not** reuse these literals in any environment. Production startup rejects them via `FILVAULT_ENV=production` validation.

## Category C — Real/shared cloud or production credentials

**Finding:** None identified.

History review did not surface AWS access keys, SMTP passwords, production RDS URLs, paid SaaS tokens, or other third-party credentials outside the local dev defaults in Category B.

**Action:** No cloud credential rotation required for this audit. Re-scan before any production deploy; rotate immediately if a future audit reclassifies a finding.

## Category D — Git history rewrite

Not required for this issue. Remediation is credential rotation (N/A for Category C) plus prevention gates (production validation + PR secret scanning).

## Prevention controls added

| Control | Purpose |
|---|---|
| `FILVAULT_ENV=production` validation in `apps/api/internal/platform/config` | Fail closed on placeholders, localhost DB/CORS, console mailer, dev seed |
| `.gitleaks.toml` + `scripts/verify-secret-scan-gate.sh` | PR CI secret scan of current tree + scanner self-test |
| `.github/workflows/ci.yml` `secret-scan` job | Blocks regressions on pull requests |

## Allowlist policy (gitleaks)

No whole-file path allowlists for `Makefile`, `docker-compose.yml`, `.github/workflows/ci.yml`, or config/scanner sources.

Allowlisting is limited to:

- **Rule-scoped regex allowlist** on `filvault-replace-with-placeholder` for the seven exact `replace-with-local-*` placeholder literals
- **Rule-scoped paths** on `filvault-historical-compose-default` for validation/tests/security docs that name blocked literals without deploying them
- **Line-level `gitleaks:allow`** on synthetic S3 env fixture lines in `load_test.go` (test-only; not a file skip)

Default gitleaks detectors remain active on Compose, Makefile, CI, and application config. Production startup still rejects placeholders when `FILVAULT_ENV=production`.
