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
