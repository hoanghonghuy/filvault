# Phase 1 — Checklist bàn giao

Cập nhật: 2026-08-14.

Nguồn sự thật cho thứ tự làm: [06-phase-1-plan.md](06-phase-1-plan.md).  
Nhánh gần nhất khi ghi: `develop`.

Cách chạy hiện tại:

```bash
make test      # Docker Postgres + test API
make migrate
make run-api   # invite dev: dev-invite
```

Go: `~/.local/go/bin/go` (Makefile đã trỏ). Core **không** import AWS SDK. SQL chỉ trong `apps/api/internal/platform/postgres`.

---

## Đã xong

### Spec (chốt; chưa implement hết)

- [x] Concept v0.1 — `docs/concept-product-scope_v0.1.md`
- [x] Spec 01–07 trong `docs/spec/`
- [x] ADR 0001 — AWS-first, provider-agnostic core
- [x] ADR 0002 — nghiệp vụ Phase 1 (invite, verify, Photos, allowlist, trash)
- [x] Tích hợp app khác = **full API** (không chỉ Photos), làm **sau khi** Filnest hoàn chỉnh; Phase 1 không `tenant_id` / API key

### Slice 0 — repo / DB

- [x] `apps/api`, Makefile, `.env.example`, `.gitignore`, README
- [x] Docker Compose: Postgres 16
- [x] Migration `00001_phase1` up/down: `users`, `refresh_tokens`, `folders`, `files`, `albums`, `album_items`
- [x] Migrator pgx (không ORM, không goose/`database/sql`)
- [x] `make migrate` / `make migrate-down` / `make test`
- [x] `FILNEST_METADATA_STORE=dynamodb` fail fast trong `config.Load`

### Slice 1 — Auth API

- [x] `POST /api/v1/auth/register` + mã mời env (`REGISTER_DISABLED` nếu env trống)
- [x] `POST /api/v1/auth/login` (không lộ email có/không)
- [x] `POST /api/v1/auth/refresh`
- [x] `POST /api/v1/auth/logout` — chỉ revoke **một** session
- [x] `GET /api/v1/users/me`
- [x] JWT access ~15 phút; refresh ~7 ngày (lưu hash SHA-256)
- [x] Password argon2id; không trả `passwordHash`
- [x] Email lowercase; user mới `emailVerified: false`
- [x] Error JSON `{ "error": { "code", "message" } }`
- [x] Test: thiếu/sai invite, trùng email, login sai, refresh, logout một máy
- [x] `make run-api`

---

## Chưa làm — còn trong Phase 1

Làm tiếp **theo thứ tự slice**. TDD: test fail rồi mới code.

| | Slice | Việc | Xong khi |
|---|---|---|---|
| [ ] | **1b** | Verify email | SMTP hoặc console; mã 6 số; chưa verify thì không dùng kho file |
| [ ] | **2** | Folders | create, browser root, rename, move, xóa folder trống |
| [ ] | **3** | Vertical slice file | allowlist + 100 MiB; presign → PUT MinIO/S3 → complete (`Head`) → download |
| [ ] | **4** | Rename / move file | chỉ metadata; `object_key` không đổi |
| [ ] | **5** | Trash + setting | delete / restore / permanent; setting theo user; ticker tự xóa |
| [ ] | **6** | Photos | timeline + album; không thumbnail; grid không `<img>` original |
| [ ] | **7** | Profile | đổi displayName + password (revoke mọi refresh) |
| [ ] | **8** | Quota + search + storage | chặn vượt quota; search tên; thanh dung lượng |
| [ ] | **9** | Siết DoD | ownership A/B, bucket private, migrate idempotent đã có một phần |

### Còn thiếu dù đã khai trong spec / compose

- [ ] Vue web — `apps/web` chưa có
- [ ] Mailer port (console / SMTP); SES chỉ dành chỗ
- [ ] `ObjectStore` + adapter S3 (MinIO và AWS cùng implementation)
- [ ] MinIO trong compose **chưa** gắn vào luồng upload
- [ ] Package `file` / `folder` / `photo` / `search`
- [ ] CORS API; slog request đầy đủ
- [ ] `403 EMAIL_NOT_VERIFIED` trên Files/Photos (phụ thuộc 1b + slice 2/3)
- [ ] Chứng minh một lần với AWS S3 thật (không chặn DoD local)

### Test tối thiểu Phase 1 còn thiếu

- [ ] Verify: chưa verify → 403 kho; console in mã; verify xong dùng được
- [ ] Ownership: user B 404 trên id user A
- [ ] Quota / 100 MiB / ngoài allowlist: tạo session bị từ chối
- [ ] Complete: size từ `Head` thắng size client khai
- [ ] Unique name sibling → 409
- [ ] Folder không rỗng: DELETE → 409
- [ ] Album: không thêm PDF; xóa album không xóa file

---

## Cố ý không làm ở Phase 1

Không phải nợ. Không làm lẫn vào slice hiện tại.

- CLI, desktop sync, sharing, versioning
- RDS (Phase 2 = đổi host, cùng schema), DynamoDB **code**, SQS, Lambda, CloudFront, SES, Cognito
- Terraform, GitHub OIDC / CI deploy
- Thumbnail server, FTS, multipart / resumable, xóa folder đệ quy
- `tenant_id` / API key — tích hợp app khác (ví dụ trường mầm non) **sau khi sản phẩm Filnest xong**, dùng **cả API**

---

## Việc tiếp theo cho người nhận

1. Đọc [README.md](README.md) → [07-phase-1-business.md](07-phase-1-business.md) → file này.
2. `make test` phải xanh.
3. Bắt đầu slice **1b — Verify email**.
4. Lệch spec → sửa spec (và ADR nếu đổi kiến trúc), không code xong rồi viết ngược.
