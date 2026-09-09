# Credential rotation evidence (issue #3)

This record documents rotation/revocation decisions after the secret-history audit. **No secret values appear below.**

## Summary

| Credential class | Was it real/shared? | Rotation action | Evidence |
|---|---|---|---|
| Postgres (local Docker) | No — local container only | N/A | Category B audit; never a managed RDS instance in repo history |
| MinIO (local Docker) | No — local object store | N/A | Category B audit; no AWS S3 keys committed |
| JWT signing secret (`FILVAULT_JWT_SECRET`) | No — local dev default only | Invalidate by using new local values in `.env`; production must inject unique secret via secret manager | Removed from Compose in `dc4441f`; blocked in production by config validation |
| Invite code (`FILVAULT_INVITE_CODE`) | No — local dev default only | Replace in local `.env`; production requires unique invite via env/secret manager | Same as above |
| Dev seed password (`FILVAULT_SEED_PASSWORD`) | No — local dev account only | Replace in local `.env`; disable `FILVAULT_SEED_DEV_USER` in production | Seed fallbacks removed from code; production rejects dev seed |
| LiveKit API key/secret | No — LiveKit demo pair in local config | Regenerate for any non-local deployment; do not reuse historical demo pair | Removed from `deploy/livekit.yaml`; supplied via env |
| AWS / SMTP / third-party SaaS | Not found in history | N/A | Category C audit |

## Operator checklist (production)

1. Set `FILVAULT_ENV=production`.
2. Inject unique values for JWT, database URL (`sslmode=require` or stricter), S3 credentials, invite code, SMTP, and LiveKit (if enabled) via your secret manager — not from `.env.example`.
3. Confirm `FILVAULT_SEED_DEV_USER` is unset or `false`.
4. Run API/worker startup once; verify fail-closed errors are absent only after real secrets are configured.
5. If any historical Category B literal was ever reused outside local Docker, rotate that component immediately and note the date here.

## Verification performed in this PR

- Full-history scan documented in [`secret-lifecycle-audit.md`](secret-lifecycle-audit.md).
- Production config validation tests in `apps/api/internal/platform/config/load_test.go`.
- Gitleaks current-tree scan + synthetic self-test via `scripts/verify-secret-scan-gate.sh`.

## Sign-off

Audit performed by automated/human review on 2026-09-09 against public repository `hoanghonghuy/filvault`. No new secret values were written to Git as part of this remediation.
