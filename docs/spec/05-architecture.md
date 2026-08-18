# 05 — Architecture spec (Phase 1)

Trạng thái: **chốt khung implement**. Cổng: [02](02-provider-strategy.md). API: [04](04-api.md). DB: [03](03-database.md). Nghiệp vụ: [07](07-phase-1-business.md).

Không ép Clean Architecture nhiều lớp. Giữ:

```text
handler → service → port → adapter
```

## 1. Monorepo

Concept §59 giữ, chỉnh `docs/` cho khớp thư mục spec hiện tại:

```text
filvault/
├── apps/
│   ├── web/                 Vue 3 + TS + Vite
│   └── api/
│       ├── cmd/api/         composition root (wiring)
│       ├── internal/
│       │   ├── auth/
│       │   ├── user/
│       │   ├── file/
│       │   ├── folder/
│       │   ├── search/      có thể mỏng, gọi file+folder repo
│       │   ├── photo/       timeline + album
│       │   └── platform/
│       │       ├── config/
│       │       ├── http/    error JSON, middleware JWT, validate
│       │       ├── postgres/  pgxpool + SQL thuần
│       │       ├── s3/      AWS + MinIO, một package
│       │       └── mailer/  console + smtp
│       └── migrations/
├── workers/                 chưa có code Phase 1
├── infrastructure/          Terraform: Phase 2
├── docs/
│   ├── concept-product-scope_v0.1.md
│   └── spec/
├── docker-compose.yml
├── Makefile
└── README.md
```

Không tạo `platform/dynamodb`, `platform/sqs`, `platform/ses`, `apps/cli`, `workers/file-processor` cho đến khi phase cần.

`internal/storage` trong concept gộp vào `file` (use case) + `platform/s3` (adapter). Tránh package `storage` vừa domain vừa AWS.

## 2. Module Go

Mỗi domain package (ví dụ `internal/file`):

```text
handler.go
service.go
port.go          interface repository + ObjectStore phía use case cần
model.go
```

SQL **không** nằm trong `file/`. `platform/postgres` implement `file.Repository`.

`cmd/api/main.go` (hoặc `wire.go`): đọc config → mở pgxpool → tạo S3 client → Mailer → new services → mount Gin.

## 2.1 Stack Go (bắt buộc)

```text
Gin
pgx v5 + pgxpool, SQL viết tay (không ORM, không database/sql)
slog structured
Validate input ở handler; service không parse JSON
Lỗi domain (code) → platform/http map JSON spec 04
```

Config thêm (cùng spec 02 / 07):

```text
FILVAULT_METADATA_STORE=postgres
FILVAULT_OBJECT_STORE=s3
FILVAULT_S3_ENDPOINT=
FILVAULT_S3_BUCKET=
FILVAULT_S3_REGION=
FILVAULT_DATABASE_URL=
FILVAULT_QUEUE=none
FILVAULT_INVITE_CODE=
FILVAULT_MAILER=auto
FILVAULT_SMTP_HOST=            # trống → console mailer
FILVAULT_DEFAULT_TRASH_AUTO_DELETE=false
FILVAULT_DEFAULT_TRASH_RETENTION_DAYS=30
```

`FILVAULT_METADATA_STORE=dynamodb` → process **exit** với lỗi rõ.

## 3. ObjectStore port (Phase 1)

```go
type ObjectStore interface {
    CreateUploadURL(ctx context.Context, key string, opts UploadOptions) (PresignedURL, error)
    CreateDownloadURL(ctx context.Context, key string, opts DownloadOptions) (PresignedURL, error)
    Head(ctx context.Context, key string) (ObjectStat, error)
    Delete(ctx context.Context, key string) error
}

type PresignedURL struct {
    URL       string
    ExpiresAt time.Time
}

type ObjectStat struct {
    Size        int64
    ContentType string
}
```

Không dùng type AWS SDK ở đây. `UploadOptions` gồm `ContentType`, `Expires`.

Bucket **private**. Không `public-read`. IAM (khi AWS): tối thiểu `GetObject PutObject DeleteObject HeadObject` (+ `s3:PutObject` presign).

Mailer (spec 02 §3.6):

```go
type Mailer interface {
    Send(ctx context.Context, msg Mail) error
}
```

`console` được log mã verify. `smtp` không log mã. Test dùng fake.

## 4. Flow upload (lớp)

```text
Vue                    Go core                         Adapters
 │
 │ POST upload-sessions
 ▼
handler  →  file.Service
              ├─ quota + unique name     → Folder/File/User repo  → postgres
              ├─ tạo PENDING
              └─ CreateUploadURL                                  → s3
 │
 │ PUT presigned                       (không qua Go)
 ▼
MinIO / S3
 │
 │ POST .../complete
 ▼
handler  →  file.Service
              ├─ Head(key)                                        → s3
              ├─ so size / quota
              └─ READY + storage_used                             → postgres
```

Download tương tự: service check owner + `READY` → `CreateDownloadURL` → Vue lấy bytes từ object storage.

## 5. Auth trong request

```text
Gin middleware
  → parse Bearer JWT
  → UserID vào context
  → handler lấy UserID, truyền service
```

Service không đọc header. Refresh không đi middleware access; route riêng.

Hash mật khẩu: argon2id, wrapper trong `internal/auth` (không phải AWS).

## 6. Web (Vue)

Đủ cho DoD Phase 1:

```text
/login  /register  /verify-email
/                         overview (logo home)
/files                    browser + breadcrumb
/photos                   timeline (placeholder, không <img> original)
/photos/albums/:id
/files/trash
/settings                 displayName, password, trash auto-delete
upload (progress, lỗi, hủy)
download
rename / move / delete / restore
storage usage
search by name
```

Pinia store: auth, browser. API client gọi `/api/v1`. Không gọi S3 SDK từ Vue — chỉ `PUT/GET` presigned URL.

CORS: API cho origin Vite; MinIO/S3 bucket CORS cho `PUT` từ origin web.

## 7. Local runtime

`docker-compose.yml`:

```text
postgres
minio
minio-init     (tạo bucket + CORS)
```

API và web chạy host (hoặc compose thêm) để hot-reload. Makefile: `make dev`, `make migrate`, `make test`.

Adminer/pgAdmin: tùy chọn, không bắt buộc.

## 8. Test

TDD theo slice (spec 06). Tầng ưu tiên:

| Tầng | Gì |
|---|---|
| Service | quota, unique name, ownership, allowlist, PENDING→READY, album chỉ media |
| Handler | HTTP status + error code; JWT; EMAIL_NOT_VERIFIED |
| Adapter postgres | repository + migration trên Postgres test (pgx) |
| Adapter s3 | fake `ObjectStore` ở test service; contract test MinIO khi cần |
| Mailer | fake in-memory; test console không bắt SMTP |

Không cần test Lambda/SQS. Fake `ObjectStore` in-memory cho unit service là đủ; không mock SQL trong service test nếu service không biết SQL.

## 9. Việc cố ý để trống

```text
Terraform, RDS, CloudWatch, GitHub OIDC   → Phase 2
SQS, Lambda, thumbnail                    → Phase 3
CLI, sync                                 → phase sau
```

Khung package ở §1 phải cho Phase 2 chỉ đổi `DATABASE_URL` + endpoint S3, không đổi service.
