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

→ **Việc tiếp theo không phải deploy/AWS.** Smoke API xuyên suốt đã xong (2026-08-23, cuối trang); còn **smoke tay phần UI**. Feature Phase 2 web đã Done — xem [10-phase-2-status.md](10-phase-2-status.md); không nhầm với cửa deploy/RDS.

### Bắt đầu từ đây

1. `make up` → mở `http://localhost:5173`.
2. Đăng nhập bằng `FILVAULT_SEED_EMAIL` / `FILVAULT_SEED_PASSWORD` từ `.env` (copy từ `.env.example`).
3. Smoke tay phần UI (checklist dưới; phần API đã smoke đủ 2026-08-23). Ghi lệch spec / bug nếu thấy.
4. Feature Phase 2 web: [10-phase-2-status.md](10-phase-2-status.md) — đã Done cả 6 slice. Deploy/AWS: bàn riêng sau smoke Phase 1.

> Lưu ý (2026-08-23): container `api`/`web` cũ build từ image 17/08 **đã được rebuild** bằng
> `docker compose up -d --build api web` — hiện đang chạy code Phase 2 mới nhất; API smoke lại
> xanh trên bản này (activity, share-links, shares, favorites). CLI `bin/filvault.exe` cũng đã
> rebuild sau fix piped-login. Nếu pull code mới, luôn chạy lại lệnh compose trên trước khi smoke UI.

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

| Trường | Biến env (xem `.env.example`) |
|---|---|
| Email | `FILVAULT_SEED_EMAIL` |
| Mật khẩu | `FILVAULT_SEED_PASSWORD` |
| Mã mời (register tay) | `FILVAULT_INVITE_CODE` |

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
- [x] Login, register, verify, overview, files, photos, album, trash, settings
- [x] Shell mobile bottom nav + desktop side nav
- [x] Bottom sheet / confirm / prompt / action sheet (không `window.prompt`)
- [x] Upload presign + progress; Photos placeholder (không `<img>` original)
- [x] Loading skeleton riêng Files / Photos / Trash / Album

### Web — UI/UX theo DESIGN.md

| Màn | Functional | Chuẩn UI/UX (review) |
|---|---|---|
| Login / Register / Verify (`AuthCard`) | xong | **xong** (2026-08-17) |
| Overview (`/`, logo home) | xong | **xong** (2026-08-18) |
| Files | xong | **xong** (2026-08-17) |
| Photos | xong | **xong** (2026-08-17) |
| Album detail | xong | **xong** (2026-08-17) |
| Trash | xong | **xong** (2026-08-17) |
| Settings | xong | **xong** (2026-08-17) |

Auth đã sửa khi review: tách loading Verify/Resend, invite không `autocapitalize`, lỗi OTP không bị nuốt bởi `VALIDATION_ERROR` generic, chặn open redirect `?redirect=`, scroll màn ngắn, hint CSS.

Các màn còn lại đã siết UI/UX (2026-08-17): icon nav (side + bottom), icon loại file, icon empty state, tìm kiếm live (debounce), meta file rút gọn, checkbox trash settings, storage bar rõ hơn, badge `itemCount` cho album (kèm field mới trong API `GET /photos/albums`).

---

## Còn lại trước khi sang Phase 2

- [ ] Smoke **phần UI thuần** (click nav, FAB, progress bar — xem ghi chú dưới; API đã smoke đủ 2026-08-23)
- [ ] Chứng minh một lần với **AWS S3 thật** (không chặn DoD local; làm ở cửa Phase 2 cũng được)

### Smoke API xuyên suốt — đã chạy (2026-08-23)

Chạy trên stack thật (`make up`: Postgres + MinIO + API + web), bằng user tạm
`smoke-p1-*@filvault.local` (đăng ký → đọc mã verify từ log mailer console → verify),
không đụng dữ liệu user dev. Kết quả:

- Auth: register trả session `emailVerified=false`; vault chặn 403 khi chưa verify
  (`GET /storage`, `PATCH /users/me`); mã 6 số từ mailer console verify được 204;
  login sau verify `emailVerified=true`.
- Files/folders: tạo folder, upload allowlist vào folder, browser hiện đúng,
  rename, move về root (`folderId=null`), presign download có `expiresAt`,
  search theo tên.
- Từ chối hợp lệ: `.exe` ngoài allowlist → 400; >100 MiB → 413; mime lệch extension
  → 400; trùng tên sibling trong cùng folder → 409; folder không rỗng xóa → 409.
- Photos: jpg + video mp4 vào timeline; PDF không vào timeline và bị từ chối khi thêm
  album (400). Thumbnail theo prefs: mặc định bật → item có presigned
  `thumbnailUrl`; tắt `imageThumbnailsEnabled` → item không còn URL (placeholder grid);
  bật lại → có lại. Không bao giờ expose original qua grid.
- Album: tạo, thêm jpg, từ chối PDF, `itemCount` đúng khi thêm/gỡ, xóa album
  không ảnh hưởng file (file vẫn READY).
- Trash/quota: soft delete → có trong `/trash`, quota vẫn tính; restore
  (`POST /files/:id/restore`) → READY + rời trash; purge
  (`DELETE /trash/files/:id`) → file 404 + quota giảm đúng số byte object.
- Settings: đổi displayName, trash prefs (auto-delete + retention), đổi mật khẩu
  (sai current → 401; đổi xong token mới, mật khẩu cũ login 401, refresh token cũ
  của phiên trước password change hoạt động), logout thu hồi refresh token
  (refresh sau logout → 401).
- Ownership: user B đọc/rename/delete file của A đều 404 uniform (không dò ID).
- Search filters (spec 09 §1.2): `type=image|document` tách đúng jpg/pdf;
  `sort=size&order=asc` đúng thứ tự; `type` lạ → 400.

Lệch spec phát hiện khi smoke: không có. Route restore là `POST /files/:id/restore`
(không nằm dưới `/trash/*`) — đã khớp spec 04, chỉ note để khỏi nhầm.

### Smoke test tay phần UI (còn lại)

Phần chỉ kiểm tra được bằng mắt/tương tác (API tương ứng bên dưới mỗi dòng đã PASS ở trên):

- [ ] Progress bar upload chạy rồi ẩn (API upload-complete PASS)
- [ ] Ô **F** header mobile / wordmark sidebar desktop về Overview `/`
- [ ] Mobile shell: bottom nav + FAB; desktop: side nav
- [ ] Mở ảnh/video từ Photos xem được nội dung qua presign (URL presign PASS)
- [ ] Đăng nhập lại bằng UI với credential từ `.env` (`FILVAULT_SEED_EMAIL` / `FILVAULT_SEED_PASSWORD`) sau toàn bộ flow trên

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

1. `make up` → login bằng credential từ `.env`.
2. Smoke API xuyên suốt đã xong (2026-08-23, xem trên). Chạy nốt **Smoke test tay phần UI**; ghi bug nếu lệch spec / `DESIGN.md`.
3. Lệch spec → sửa spec (và ADR nếu đổi kiến trúc), không code xong rồi viết ngược.
4. Phase 2 feature (S1–S6) đã Done — xem [10-phase-2-status.md](10-phase-2-status.md); smoke Phase 2 đã chạy kèm ở đó. Deploy/AWS vẫn tách khỏi checklist feature.
