# Filvault Web

Vue 3 + TypeScript + Vite + Pinia.

## Setup

Cần **Node.js ^22** (hoặc >=24.12).

```bash
cp .env.example .env
npm install --legacy-peer-deps
npm run dev
```

Mặc định gọi API tại `http://localhost:8080/api/v1`. Chạy API trước (`make dev-api` từ repo root).

## Routes

- `/` — overview (logo home)
- `/login`, `/register`, `/verify-email`
- `/files` — browser, upload (presigned PUT), search, download
- `/photos` — timeline (placeholder, không load original), albums
- `/trash` — restore / permanent delete
- `/settings` — displayName, password, trash settings

Upload dùng presigned URL thẳng tới MinIO/S3 — không qua body Go API.

## Docker (cùng stack với API)

Từ repo root:

```bash
make up    # postgres + minio + api + web
```

Web: `http://localhost:5173` · API: `http://localhost:8080`

## E2E (Playwright)

Chạy smoke critical-path trên stack local thật (Postgres + MinIO + API + Vite), không mock backend:

```bash
# Từ repo root — một lệnh: compose, migrate, seed, API, Playwright
make e2e
```

Hoặc nếu API đã chạy (`make dev-api`) và chỉ cần web + test:

```bash
cd apps/web
E2E_EMAIL=dev@filvault.com E2E_PASSWORD=<FILVAULT_SEED_PASSWORD từ .env> npm run test:e2e
```

Biến môi trường (không commit secret):

| Biến | Mục đích |
|------|----------|
| `E2E_EMAIL` / `E2E_PASSWORD` | Tài khoản seed (`FILVAULT_SEED_*`) |
| `E2E_BASE_URL` | Web (mặc định `http://127.0.0.1:5173`) |
| `E2E_API_URL` | API health check (mặc định `http://127.0.0.1:8080`) |

Smoke chạy ở 3 viewport: mobile (≤767), tablet (768–1023), desktop (≥1024). Artifact (trace/screenshot/video) chỉ lưu khi fail.
