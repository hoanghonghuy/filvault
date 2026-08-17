# Phase 1 — Checklist bàn giao

Cập nhật: 2026-08-17.

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

### Slice 1b — Verify email

- [x] Port `Mailer` + adapter `console` / `smtp` (`FILNEST_MAILER=auto` → SMTP nếu có host, không thì console)
- [x] `POST /api/v1/auth/resend-verification` — luôn `204`; email lạ không lộ
- [x] `POST /api/v1/auth/verify-email` — mã 6 số, TTL 15 phút, lưu hash; sai/hết hạn → `400 VALIDATION_ERROR`
- [x] Register gửi mã qua Mailer; Memory mailer dùng trong test
- [x] Test: resend → verify sai → verify đúng → `emailVerified: true`

### Slice 2 — Folders

- [x] `POST /api/v1/folders` — create root / nested; trùng tên sibling → `409`
- [x] `GET /api/v1/folders/{id}` — ownership → `404` nếu không phải owner
- [x] `PATCH /api/v1/folders/{id}` — rename, move; chu trình → `400 VALIDATION_ERROR`
- [x] `DELETE /api/v1/folders/{id}` — soft delete; còn con → `409`
- [x] `GET /api/v1/browser` — list root (folders + files READY)
- [x] Middleware `403 EMAIL_NOT_VERIFIED` trên folder/browser
- [x] Package `internal/folder` + postgres adapter
- [x] Test HTTP: unverified forbidden, CRUD flow, ownership

### Slice 3 — Vertical slice file

- [x] Port `ObjectStore` + adapter `memory` (test) / `s3` (MinIO/AWS)
- [x] Allowlist extension + MIME; `FILE_TOO_LARGE` (> 100 MiB)
- [x] `POST /api/v1/files/upload-sessions` — quota check, `PENDING`, presign
- [x] `POST /api/v1/files/{id}/complete` — `Head`, `READY`, cập nhật `storage_used`
- [x] `GET /api/v1/files/{id}`, `GET .../download`, `DELETE` abort `PENDING`
- [x] Browser hiển thị file `READY`
- [x] Test: reject type/size; upload → complete → browser → download

### Slice 4 — Rename / move file

- [x] `PATCH /api/v1/files/{id}` — `{ name?, folderId? }`; chỉ `READY`
- [x] Tên trùng sibling → `409`; `PENDING` patch → `INVALID_STATE`
- [x] `object_key` không đổi (download vẫn hoạt động sau patch)
- [x] Test: rename, move vào folder, browser, conflict, invalid state

### Slice 5 — Trash + setting

- [x] `DELETE /files/{id}` — soft delete `READY`; `PENDING` vẫn abort
- [x] `GET /api/v1/trash` — list file/folder đã xóa
- [x] `POST /files/{id}/restore`, `POST /folders/{id}/restore` — conflict nếu trùng tên
- [x] `DELETE /trash/files/{id}`, `DELETE /trash/folders/{id}` — permanent (file: xóa object + giảm quota)
- [x] `PATCH /users/me` — `trashAutoDeleteEnabled`, `trashRetentionDays`
- [x] Ticker auto-cleanup in-process (`RunAutoCleanup`, interval 1h)
- [x] Package `internal/trash` + test

---

## Chưa làm — còn trong Phase 1

Làm tiếp **theo thứ tự slice**. TDD: test fail rồi mới code.

| | Slice | Việc | Xong khi |
|---|---|---|---|
| [x] | **1b** | Verify email | SMTP/console + mã 6 số (gate kho file khi có API folder/file) |
| [x] | **2** | Folders | create, browser root, rename, move, xóa folder trống |
| [x] | **3** | Vertical slice file | allowlist + 100 MiB; presign → complete (`Head`) → browser → download |
| [x] | **4** | Rename / move file | chỉ metadata; `object_key` không đổi |
| [x] | **5** | Trash + setting | delete/restore/permanent; PATCH setting; ticker |
| [ ] | **6** | Photos | timeline + album; không thumbnail; grid không `<img>` original |
| [ ] | **7** | Profile | đổi displayName + password (revoke mọi refresh) |
| [ ] | **8** | Quota + search + storage | chặn vượt quota; search tên; thanh dung lượng |
| [ ] | **9** | Siết DoD | ownership A/B, bucket private, migrate idempotent đã có một phần |

### Còn thiếu dù đã khai trong spec / compose

- [ ] Vue web — `apps/web` chưa có
- [x] Mailer port (console / SMTP); SES chỉ dành chỗ
- [x] `ObjectStore` + adapter S3 (MinIO/AWS; test dùng memory)
- [ ] MinIO trong compose — chạy `make run-api` với `FILNEST_S3_ENDPOINT` (upload thật qua presign)
- [x] Package `file` (`folder` đã có); còn `photo` / `search`
- [ ] CORS API; slog request đầy đủ
- [x] `403 EMAIL_NOT_VERIFIED` trên folder/file/browser
- [ ] Chứng minh một lần với AWS S3 thật (không chặn DoD local)

### Test tối thiểu Phase 1 còn thiếu

- [x] Verify: chưa verify → 403 folder/browser
- [x] Ownership: user B 404 trên folder user A
- [x] Quota / 100 MiB / ngoài allowlist: tạo session bị từ chối (một phần — chưa test quota DB)
- [x] Complete: size từ `Head` (test 128 bytes)
- [x] Unique name sibling → 409 (file patch)
- [x] Folder không rỗng: DELETE → 409
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
3. Bắt đầu slice **6 — Photos**.
4. Lệch spec → sửa spec (và ADR nếu đổi kiến trúc), không code xong rồi viết ngược.
