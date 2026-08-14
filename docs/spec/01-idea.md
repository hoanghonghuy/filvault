# 01 — Spec ý tưởng

Trạng thái: **chốt cho hướng đi**. Chi tiết sản phẩm Phase 1 vẫn lấy từ concept v0.1.

## 1. Filnest là gì

Filnest là nền tảng lưu trữ file cá nhân trên cloud (đám mây): upload, download, tổ chức folder, xóa/khôi phục, quản lý dung lượng.

Khi hoàn thiện, cùng một API phục vụ:

```text
Web (Vue)    CLI (filnest)    Desktop sync
                 │
                 ▼
            Filnest API
```

Mobile là khả năng mở rộng, không phải mục tiêu hiện tại.

Không xây bản clone Google Drive đầy đủ ngay. Phase hiện tại chỉ làm **lõi lưu trữ file tốt**, có AWS thật, kiến trúc đủ sạch để thêm sync, sharing, versioning, worker, CLI sau này mà không đập đi làm lại.

## 2. Hai mục tiêu song song

### Sản phẩm

Trải nghiệm gần Drive / Dropbox / OneDrive, nhưng nhỏ gọn, tập trung chức năng cốt lõi.

### Học tập

Thực hành stack đã chốt trong concept:

```text
Vue 3, TypeScript, Vite
Go, Gin
PostgreSQL
REST, authn/authz, database design
AWS IAM, S3, RDS, DynamoDB, Lambda, SQS, CloudFront, CloudWatch
Docker, Terraform, GitHub Actions, OIDC
```

Không thêm AWS service chỉ để có logo. Mỗi service phải giải quyết một vấn đề thật.

Hướng học AWS được siết thêm bằng nguyên tắc kiến trúc:

```text
AWS-first integration
provider-agnostic core
```

Xem [02-provider-strategy.md](02-provider-strategy.md).

## 3. Quyết định sản phẩm đã chốt (từ concept)

| Mục | Quyết định |
|---|---|
| Tên | Filnest; CLI `filnest` |
| Backend | Go + Gin; **modular monolith** (một ứng dụng, chia module theo domain); không microservices |
| Metadata (siêu dữ liệu) | PostgreSQL là source of truth (nguồn sự thật) mặc định |
| File binary | Object storage (kho object); **không** nhét vào PostgreSQL |
| Object key | ID ổn định, không lấy đường dẫn/tên file làm S3 key |
| Upload | Presigned URL (URL ký sẵn, hết hạn nhanh); client upload thẳng lên object storage |
| Auth Phase 1 | Email + password, JWT, invite env, verify email (SMTP hoặc console); **chưa** Cognito |
| Web Phase 1 | My Files + Photos (cùng kho; album/timeline; không thumbnail server) |
| IaC | Terraform khi infrastructure ổn định |
| CI | GitHub Actions + OIDC tới IAM Role; tránh access key dài hạn |

## 4. Tách hai loại dữ liệu

Đây là lõi sản phẩm, không đổi:

```text
PostgreSQL (metadata)          Object storage (bytes)
id, name, size, mime,          users/{userId}/files/{fileId}
object_key, status, ...
```

Rename file chỉ sửa metadata. Không copy/xóa object vì đổi tên.

File không “tồn tại” ngay khi tạo record. Có vòng đời trạng thái (`PENDING` → `UPLOADED` → … → `READY`) như concept §8.

## 5. Vertical slice đầu tiên

Mục tiêu kỹ thuật đầu tiên không phải Lambda, SQS hay sync. Một flow hoàn chỉnh:

```text
Register / Login
    → Create folder
    → Request upload session
    → Presigned URL
    → Client PUT thẳng lên object storage
    → Confirm upload
    → Metadata lưu
    → File hiện trong My Files
    → Download qua presigned URL
```

Local: PostgreSQL + MinIO. Production: cùng cổng, adapter AWS (S3; Postgres trên RDS khi deploy).

## 6. Ranh giới Phase 1

Làm: auth (invite + verify email), user (đổi tên/mật khẩu), folder, file, upload/download, browser **và Photos** (timeline + album), search tên, trash (kèm setting tự xóa), quota, allowlist + 100 MiB.

Chi tiết nghiệp vụ: [07-phase-1-business.md](07-phase-1-business.md).

Không làm ngay (concept §23 vẫn đúng về **sản phẩm**, có siết thêm):

```text
desktop/CLI sync, version history, public share, team workspace
SQS, Lambda worker, thumbnail server, virus scan, OCR, FTS nặng
CloudFront, multi-region, mobile, WebSocket, microservices
tenant/app API cho dự án khác (mầm non, …) — sau khi Filnest xong
```

Điều **được phép ngay từ Phase 1** (khác với “implement nhiều cloud”):

- Định nghĩa cổng (port) cho storage, metadata, queue, …
- Implement đúng adapter đang cần: Postgres + S3/MinIO
- Ghi mapping cho adapter tương lai (RDS là deploy Postgres; DynamoDB là adapter khác)

Không viết code DynamoDB, Lambda, SQS trong Phase 1.

## 7. Backend shape (không ép Clean Architecture)

Giữ như concept §60:

```text
handler → service → repository / port → adapter
```

Module theo domain (`file`, `folder`, `user`, …). Wiring (lắp adapter) ở composition root (`cmd/api`).

Chi tiết tiếp:

- Cổng / nhà cung cấp: [02-provider-strategy.md](02-provider-strategy.md)
- Schema: [03-database.md](03-database.md)
- HTTP: [04-api.md](04-api.md)
- Module & flow: [05-architecture.md](05-architecture.md)
- Thứ tự làm + DoD: [06-phase-1-plan.md](06-phase-1-plan.md)
- Nghiệp vụ chốt: [07-phase-1-business.md](07-phase-1-business.md)
