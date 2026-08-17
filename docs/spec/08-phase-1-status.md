# Phase 1 — Checklist bàn giao

Cập nhật: **2026-08-17 tối**.  
Nhánh: `develop`.  
Dừng tại đây để mai làm tiếp — **chưa sang Phase 2 (deploy / AWS)**.

Nguồn sự thật cho thứ tự làm: [06-phase-1-plan.md](06-phase-1-plan.md).  
Nghiệp vụ: [07-phase-1-business.md](07-phase-1-business.md).  
UI: [`DESIGN.md`](../../DESIGN.md).

Go: `~/.local/go/bin/go` (Makefile đã trỏ). Core **không** import AWS SDK. SQL chỉ trong `apps/api/internal/platform/postgres`.

---

## Dừng ở đâu (đọc phần này trước khi làm tiếp)

**API Phase 1 (slice 0–9) đã xong.** Test API xanh. CI lint/typecheck/test (BE + FE) đã có.

**Web đã dùng được đủ API** (login → files/photos/trash/settings), mobile-first shell, skeleton, sheet thay prompt/confirm.

**UI/UX chuẩn `DESIGN.md` mới hoàn thiện xong màn Auth** (Login / Register / Verify) — đã tự review và sửa. **Các màn còn lại vẫn là bản functional**, chưa pass cùng checklist UI/UX như auth card.

→ **Mai không bắt đầu deploy/AWS.** Tiếp tục siết UI/UX các màn trong app, rồi smoke test, rồi mới kế hoạch Phase 2.

### Mai bắt đầu từ đây

1. Đăng nhập `http://localhost:5173` bằng tài khoản dev (bảng dưới).
2. Siết UI/UX theo `DESIGN.md`, **từng màn** (cùng cách đã làm với auth card: chuẩn → implement → tự review):
   1. **Files** — browser, search, upload FAB, rename/move/delete, empty, skeleton
   2. **Photos** + **Album detail** — timeline, album, placeholder (không `<img>` original)
   3. **Trash** — restore / xóa vĩnh viễn
   4. **Settings** — profile, mật khẩu, trash prefs, logout
3. Smoke test xuyên suốt với user dev (upload → Photos → trash → restore → xóa vĩnh viễn → đổi tên/mật khẩu).
4. Chỉ khi 2–3 xong mới bàn Phase 2 (Terraform, RDS, S3 AWS thật).

Skill UI: `/ui-ux-standards`. Object tiếp theo gợi ý: **Files list row + toolbar + FAB**.

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

### Infra local + CI

- [x] Docker Compose: postgres, minio, api, web (`make up`)
- [x] CI GitHub Actions: lint / typecheck / test API + lint / typecheck web
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
| Files | xong | **chưa** — mai làm |
| Photos | xong | **chưa** |
| Album detail | xong | **chưa** |
| Trash | xong | **chưa** |
| Settings | xong | **chưa** |

Auth đã sửa khi review: tách loading Verify/Resend, invite không `autocapitalize`, lỗi OTP không bị nuốt bởi `VALIDATION_ERROR` generic, chặn open redirect `?redirect=`, scroll màn ngắn, hint CSS.

---

## Còn lại trước khi sang Phase 2

- [ ] UI/UX Files, Photos, Album, Trash, Settings — cùng bar với auth card
- [ ] Smoke test tay xuyên suốt với `dev@filvault.com`
- [ ] Chứng minh một lần với **AWS S3 thật** (không chặn DoD local; làm ở cửa Phase 2 cũng được)

### Test API tối thiểu — đã có

Verify 403, ownership 404, quota / 100 MiB / allowlist, complete `Head`, unique sibling 409, folder không rỗng 409, album không PDF.

---

## Cố ý không làm ở Phase 1

Không phải nợ. Không nhét vào công việc mai.

- CLI, desktop sync, sharing, versioning
- RDS, DynamoDB **code**, SQS, Lambda, CloudFront, SES, Cognito
- Terraform, GitHub OIDC / CI **deploy**
- Thumbnail server, FTS, multipart / resumable, xóa folder đệ quy
- `tenant_id` / API key

---

## Việc tiếp theo cho người nhận (mai)

1. Đọc [`DESIGN.md`](../../DESIGN.md) §4 Auth card (mẫu đã xong) + §9 screen map.
2. `make up` → login `dev@filvault.com` / `Dev1234@`.
3. Làm Files trước (list row, toolbar, FAB, empty, lỗi).
4. Lệch spec → sửa spec (và ADR nếu đổi kiến trúc), không code xong rồi viết ngược.
