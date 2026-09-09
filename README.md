# Filvault

Personal cloud storage. Spec: [`docs/spec/README.md`](docs/spec/README.md).  
Checklist bàn giao Phase 1: [`docs/spec/08-phase-1-status.md`](docs/spec/08-phase-1-status.md).

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
