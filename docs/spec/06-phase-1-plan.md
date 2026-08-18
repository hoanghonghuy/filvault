# 06 — Phase 1 plan & Definition of Done

Trạng thái: **chốt thứ tự làm**. Tiến độ thực tế: [08-phase-1-status.md](08-phase-1-status.md). Không bắt đầu Phase 2 trước khi DoD dưới đây xanh **và** smoke test tay trong file 08 đã tick.

Mục tiêu phase (concept):

> Người dùng đăng ký bằng mã mời, verify email, quản lý folder/file (allowlist, 100 MiB), xem Photos (timeline + album), upload/download qua object storage.

## 1. Thứ tự slice

Làm **đứng** từng slice (API + DB + UI tối thiểu + test). Không dựng hết skeleton AWS rồi mới có login.

| # | Slice | Xong khi |
|---|---|---|
| 0 | Repo + compose + migrate chạy được | `make migrate` lên Postgres trống |
| 1 | Auth | invite, register, login, refresh, logout, me |
| 1b | Verify email | SMTP hoặc console; chưa verify thì 403 kho file |
| 2 | Folders | create, list qua browser root, rename, move, xóa folder trống |
| 3 | **Vertical slice file** | allowlist + 100 MiB; upload-session → PUT MinIO → complete → browser → download |
| 4 | Rename / move file | đổi metadata, object_key không đổi |
| 5 | Trash + setting | delete/restore/permanent; PATCH setting; ticker (có thể stub interval ngắn khi test) |
| 6 | Photos | timeline ảnh/video; album CRUD + items; grid không `<img>` original |
| 7 | Profile | đổi displayName, đổi password (revoke token) |
| 8 | Quota + search + storage UI | chặn vượt quota; search tên; thanh dung lượng |
| 9 | Siết DoD | ownership, bucket private, argon2id, migrate idempotent |

Slice 3 vẫn là trái tim. Photos (6) **sau** khi file READY đã chạy.

TDD: mỗi slice viết test service/handler **trước**, thấy fail, rồi implement. Test slice 3 dùng `ObjectStore` fake hoặc MinIO trong compose. Mailer fake/console.

## 2. Có / không Phase 1

**Có**

```text
Vue, Go, Gin, pgx SQL thuần, slog
PostgreSQL, MinIO, S3 adapter
Invite env + verify email (SMTP hoặc console)
Nhiều refresh token
Files, folders, Photos (timeline + album)
Allowlist + 100 MiB
Trash + setting tự xóa (ticker in-process)
Search tên, quota
Đổi displayName / password
Presigned upload/download
```

**Không** (dù spec 02 đã dành cổng)

```text
RDS, DynamoDB code, SQS, Lambda, CloudFront, SES
Cognito, Terraform, GitHub OIDC
CLI, sync, sharing, versioning, thumbnail server, FTS
Xóa folder đệ quy, multipart, resumable
tenant / API key cho app khác
```

Chứng minh AWS S3 thật: **nên** làm một lần ở cuối Phase 1 nếu có account (đổi endpoint/credential). **Không** chặn DoD coding nếu chưa có AWS — MinIO + cùng adapter là đủ để gọi Phase 1 **local complete**. Gắn AWS S3 vào checklist “production-shaped”.

## 3. Definition of Done

### 3.1 Flow người dùng

```text
Invite + Register → Verify email (console hoặc SMTP)
→ Login
→ Create folder
→ Upload file trong allowlist (progress trên web)
→ File vào object storage
→ Metadata Postgres, status READY
→ Hiện My Files và (nếu ảnh/video) Photos timeline
→ Album: tạo, thêm ảnh
→ Rename, move
→ Download (presigned)
→ Delete → Trash → Restore
→ Permanent delete (mất object + giảm quota)
→ Đổi displayName / password
```

### 3.2 Bắt buộc kỹ thuật

```text
User A không list / download / complete file User B
Chưa verify không vào Files/Photos
Sai invite không register được
Bucket / MinIO private; không public-read
Upload không đi xuyên body Go API
Allowlist + 100 MiB được enforce
Password argon2id; không trả hash ra API
JWT access + nhiều refresh; logout một máy không đá máy khác
Migration up/down chạy lại được
docker-compose dựng Postgres + MinIO từ máy sạch
Core không import AWS SDK
FILVAULT_METADATA_STORE=dynamodb fail fast
Không SMTP → mã verify in console
Photos grid không load original 100MB
```

### 3.3 Test tối thiểu

```text
Auth: register thiếu invite, trùng email, login sai, refresh, logout một session
Verify: chưa verify → 403 kho; console mailer có mã; verify xong dùng được
Ownership: user B 404 trên id user A
Quota / 100 MiB / ngoài allowlist: tạo session bị từ chối
Complete: Head size thắng claimed size
Unique name: trùng sibling → 409
Folder không rỗng: DELETE → 409
Album: không thêm PDF; xóa album không xóa file
```

## 4. Sau Phase 1 (không làm lẫn)

| Phase | Việc |
|---|---|
| 2 | Terraform, RDS (cùng schema), deploy API, CloudWatch, CI + OIDC |
| 3 | Cổng JobQueue: S3 event → SQS → worker; thumbnail |
| 4+ | Sharing, versioning, CLI, sync |

RDS Phase 2 = đổi host, không đổi spec 03.

## 5. Khi nào được viết code

Khi slice 0 bắt đầu: bám spec 03–05 và **07**. Lệch spec → sửa spec (và ADR nếu đổi kiến trúc), không “code xong rồi viết ngược”.
