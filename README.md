# Filvault

Personal cloud storage. **Current product/release handoff:** [`docs/CURRENT_PRODUCT_STATE.md`](docs/CURRENT_PRODUCT_STATE.md).  
Specs: [`docs/spec/README.md`](docs/spec/README.md). Historical phase checklists are retained for traceability but their old “next work” sections are not the current source of truth.

## Chạy local

```bash
cp .env.example .env
make test          # Postgres + MinIO + migration + API tests
make migrate       # áp schema (chỉ infra)
make up            # Docker: postgres + minio + api + web (+ tài khoản dev)
make down          # dừng toàn bộ stack
make dev-api       # API trên host :8080 (+ tài khoản dev)
make dev-web       # Vue trên host :5173 (cần Node ^22)
make seed          # tạo tài khoản dev (khi chạy API trên host, không qua Docker)
```

Sau `make up`: web `http://localhost:5173`, API `http://localhost:8080`, MinIO `http://localhost:9002`.

## Production-like runtime (#42)

Reference stack với `FILVAULT_ENV=production`, TLS ở edge proxy, migrate trước traffic, và contract `/healthz` + `/readyz` (#41):

```bash
./deploy/production/postgres/generate-certs.sh
./deploy/production/generate-env.sh
echo '127.0.0.1 filvault.local' | sudo tee -a /etc/hosts
make prod-like-up
```

Chi tiết operator: [`deploy/production/README.md`](deploy/production/README.md). Health/readiness: [`docs/ops/health-readiness-migration.md`](docs/ops/health-readiness-migration.md).

## Tài khoản dev (test xuyên suốt)

`make up` (Docker) và `make dev-api` (API trên host) **tự tạo** tài khoản dev sau migrate — không cần chạy `make seed` riêng.

Thông tin đăng nhập dev (chỉ local) lấy từ `.env` sau khi copy `.env.example`:

| Trường | Biến env |
|---|---|
| Email | `FILVAULT_SEED_EMAIL` |
| Mật khẩu | `FILVAULT_SEED_PASSWORD` |
| Mã mời (register tay) | `FILVAULT_INVITE_CODE` |

Tài khoản đã **verify email** — đăng nhập và dùng Files/Photos ngay. Seed idempotent: chạy lại không tạo trùng.

Tuỳ chỉnh qua env: `FILVAULT_SEED_DEV_USER` (Docker: bật mặc định), `FILVAULT_SEED_EMAIL`, `FILVAULT_SEED_PASSWORD`, `FILVAULT_SEED_DISPLAY_NAME`. **Production:** không set `FILVAULT_SEED_DEV_USER=true`.
