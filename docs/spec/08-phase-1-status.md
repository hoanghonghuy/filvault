# Phase 1 — Checklist bàn giao

Cập nhật: **2026-08-18**.  
Nhánh: `develop`.  
**Chưa sang Phase 2 (deploy / AWS).**

Nguồn sự thật cho thứ tự làm: [06-phase-1-plan.md](06-phase-1-plan.md).  
Nghiệp vụ: [07-phase-1-business.md](07-phase-1-business.md).  
UI: [`DESIGN.md`](../../DESIGN.md).

Go: `~/.local/go/bin/go` (Makefile đã trỏ). Core **không** import AWS SDK. SQL chỉ trong `apps/api/internal/platform/postgres`.

---

## Dừng ở đâu (đọc phần này trước khi làm tiếp)

**API Phase 1 (slice 0–9) đã xong.** Test API xanh.

**Web đã dùng được đủ API** (login → files/photos/trash/settings), mobile-first shell, skeleton, sheet thay prompt/confirm.

**UI/UX theo `DESIGN.md` đã xong cả 6 màn** (Auth, Files, Photos, Album, Trash, Settings) — 2026-08-17.

**CI:** lint / typecheck / test (BE + FE) + workflow AI review PR (2026-08-18).

→ **Việc tiếp theo không phải deploy/AWS.** Làm **smoke test tay** xuyên suốt. Chỉ khi smoke test xanh mới bàn Phase 2.

### Bắt đầu từ đây

1. `make up` → mở `http://localhost:5173`.
2. Đăng nhập `dev@filvault.com` / `Dev1234@`.
3. Smoke test tay (checklist dưới). Ghi lệch spec / bug nếu thấy.
4. Chỉ khi bước 3 xong mới bàn Phase 2 (Terraform, RDS, S3 AWS thật).

---

## Chạy local

```bash
make up            # Docker: postgres + minio + api + web; tự seed user dev
make down
make test          # API tests (cần Postgres + MinIO)
make dev-api       # API trên host :8080 (cũng seed)
make dev-web       # Vue :5173 (Node ^22)
```

Web `http://localhost:5173` · API `http://localhost:8080` · MinIO `http://localhost:9002`.

### Tài khoản dev (đã verify)

Tạo tự động khi `make up` / `make dev-api` nếu `FILVAULT_SEED_DEV_USER=true` (chỉ bật trong `docker-compose.yml` local).

| Trường | Giá trị |
|---|---|
| Email | `dev@filvault.com` |
| Mật khẩu | `Dev1234@` |
| Mã mời (register tay) | `dev-invite` |

**Production: không set `FILVAULT_SEED_DEV_USER=true`.**

---

## Đã xong

### Spec

- [x] Concept v0.1 — `docs/concept-product-scope_v0.1.md`
- [x] Spec 01–07 trong `docs/spec/`
- [x] ADR 0001 — AWS-first, provider-agnostic core
- [x] ADR 0002 — nghiệp vụ Phase 1 (invite, verify, Photos, allowlist, trash)
- [x] `DESIGN.md` — Cal.com-like + teal, mobile-first; có mini spec **Auth card**

### API slice 0–9

Toàn bộ checklist kỹ thuật API (auth, verify, folder, file, trash, photos, profile, search, quota, ownership, ULID, migrate, MinIO private) **đã tick** — xem git history / test `make test`. Không lặp lại từng dòng ở đây.

Field thêm sau slice: `itemCount` trên `GET /photos/albums`.

### Infra local + CI

- [x] Docker Compose: postgres, minio, api, web (`make up`)
- [x] CI GitHub Actions: lint / typecheck / test API + lint / typecheck web
- [x] CI GitHub Actions: AI pull request review (PR Agent)
- [x] Seed user dev (idempotent); `apps/api/cmd/seed` + `FILVAULT_SEED_DEV_USER`

### Web — functional (đủ gọi API)

- [x] Vue 3 + Pinia + Vue Router; CORS; slog request log
- [x] Login, register, verify, files, photos, album, trash, settings
- [x] Shell mobile bottom nav + desktop side nav
- [x] Bottom sheet / confirm / prompt / action sheet (không `window.prompt`)
- [x] Upload presign + progress; Photos placeholder (không `<img>` original)
- [x] Loading skeleton riêng Files / Photos / Trash / Album

### Web — UI/UX theo DESIGN.md

| Màn | Functional | Chuẩn UI/UX (review) |
|---|---|---|
| Login / Register / Verify (`AuthCard`) | xong | **xong** (2026-08-17) |
| Files | xong | **xong** (2026-08-17) |
| Photos | xong | **xong** (2026-08-17) |
| Album detail | xong | **xong** (2026-08-17) |
| Trash | xong | **xong** (2026-08-17) |
| Settings | xong | **xong** (2026-08-17) |

Auth đã sửa khi review: tách loading Verify/Resend, invite không `autocapitalize`, lỗi OTP không bị nuốt bởi `VALIDATION_ERROR` generic, chặn open redirect `?redirect=`, scroll màn ngắn, hint CSS.

Các màn còn lại đã siết UI/UX (2026-08-17): icon nav (side + bottom), icon loại file, icon empty state, tìm kiếm live (debounce), meta file rút gọn, checkbox trash settings, storage bar rõ hơn, badge `itemCount` cho album (kèm field mới trong API `GET /photos/albums`).

---

## Còn lại trước khi sang Phase 2

- [ ] Smoke test tay xuyên suốt với `dev@filvault.com` (checklist dưới)
- [ ] Chứng minh một lần với **AWS S3 thật** (không chặn DoD local; làm ở cửa Phase 2 cũng được)

### Smoke test tay (chưa chạy)

Đăng nhập `dev@filvault.com` / `Dev1234@`, rồi tick từng dòng khi làm:

- [ ] Login thành công; chưa verify không vào được kho (thử register + verify nếu cần)
- [ ] Files: tạo folder, upload file allowlist (kèm progress), rename, move, download, search tên
- [ ] Upload ngoài allowlist / quá 100 MiB bị từ chối
- [ ] Ảnh/video vừa hiện Files vừa hiện Photos (timeline); PDF không vào Photos
- [ ] Photos: placeholder grid (không `<img>` original); mở/tải qua presign
- [ ] Album: tạo, thêm/gỡ ảnh, `itemCount` đúng, xóa album không xóa file
- [ ] Delete → Trash → Restore; xóa vĩnh viễn mất object + giảm quota
- [ ] Settings: đổi displayName, đổi mật khẩu, trash prefs, logout
- [ ] Mobile shell: bottom nav + FAB; desktop: side nav

### Test API tối thiểu — đã có

Verify 403, ownership 404, quota / 100 MiB / allowlist, complete `Head`, unique sibling 409, folder không rỗng 409, album không PDF.

---

## Cố ý không làm ở Phase 1

Không phải nợ. Không nhét vào công việc tiếp theo.

- CLI, desktop sync, sharing, versioning
- RDS, DynamoDB **code**, SQS, Lambda, CloudFront, SES, Cognito
- Terraform, GitHub OIDC / CI **deploy**
- Thumbnail server, FTS, multipart / resumable, xóa folder đệ quy
- `tenant_id` / API key

---

## Việc tiếp theo cho người nhận

1. `make up` → login `dev@filvault.com` / `Dev1234@`.
2. Chạy hết checklist **Smoke test tay** ở trên; ghi bug nếu lệch spec / `DESIGN.md`.
3. Lệch spec → sửa spec (và ADR nếu đổi kiến trúc), không code xong rồi viết ngược.
4. Smoke test xanh → mới bàn Phase 2.
