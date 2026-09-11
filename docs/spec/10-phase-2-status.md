> [!IMPORTANT]
> **Tài liệu lịch sử — không dùng làm current handoff hoặc backlog.** File này ghi lại trạng thái/next-step của Phase 2 web tại thời điểm 2026-08. Cả 6/6 slice bên dưới đã hoàn tất; hướng triển khai hiện tại, workstream đang active và release blockers được quản lý tại [`../CURRENT_PRODUCT_STATE.md`](../CURRENT_PRODUCT_STATE.md) và các GitHub issue/PR đang mở. Mọi chỉ dẫn “bắt đầu S2/S3/…” ở cuối file chỉ còn giá trị lịch sử.

# Phase 2 — Checklist bàn giao (web features)

Cập nhật: **2026-08-22**.  
Spec: [09-phase-2-features.md](09-phase-2-features.md).  
ADR phạm vi: [decisions/0003-phase-2-features-scope.md](decisions/0003-phase-2-features-scope.md).  
UI: [`DESIGN.md`](../../DESIGN.md) §9 / §9a.  
Phase 1 local: [08-phase-1-status.md](08-phase-1-status.md).

Phase 2 đợt này = **tính năng web** (search / cover / favorites / share / activity). **Không** gồm Terraform, RDS, deploy AWS — vẫn theo cửa Phase 1 / ops riêng.

---

## Dừng ở đâu (đọc trước)

| Slice | Trạng thái |
|---|---|
| S1 Search filter + sort | **Done** |
| S2 Album cover | **Done** |
| S3 Favorites | **Done** |
| S4 Public share link | **Done** (ADR 0004) |
| S6 Activity log | **Done** (2026-08-23) |
| S5 Share user nội bộ | **Done** (2026-08-23) — spec §5.1–5.6 |

→ **Phase 2 đủ 6/6 slice.** Việc còn lại: smoke tay tích lũy (mục cuối file này) + các mục "không làm" đã ghi rõ.

---

## Chạy local (giống Phase 1)

```bash
make up            # postgres + minio + api + web; seed user dev
make test          # API tests
make dev-api       # API host :8080
make dev-web       # Vue :5173
```

| | |
|---|---|
| Web | `http://localhost:5173` |
| API | `http://localhost:8080` |
| Dev login | `FILVAULT_SEED_EMAIL` / `FILVAULT_SEED_PASSWORD` từ `.env` |

---

## Đã xong

### Spec / quyết định

- [x] Spec 09 — Phase 2 features (Accepted)
- [x] ADR 0003 — phạm vi Phase 2 (Accepted)
- [x] `DESIGN.md` §9a — FilterSheet, album cover, share, favorites, activity, public page

### S1 — Search nâng cao

- [x] API: `GET /search` nhận `type`, `folderId`, `from`, `to`, `sort`, `order`, `limit` (1–50); validate + `folderId` lạ → 404
- [x] Store: ILIKE + mime groups + subtree CTE + sort
- [x] UI: `SearchFilterSheet`, chips, caption “N results · filtered”, empty “No results match your filters”
- [x] Test: `apps/api/internal/search` + `apps/web/src/views/search-filter.test.ts`

**File chính:** `apps/api/internal/search/*`, `apps/api/internal/platform/postgres/search_store.go`, `apps/web/src/components/SearchFilterSheet.vue`, `apps/web/src/views/FilesView.vue`.

---

## Còn lại theo slice

Tick khi DoD 09 §0.1 xanh **và** smoke tay slice xong.

### S2 — Album cover

- [x] Migration `albums.cover_file_id`
- [x] API: list albums có `coverUrl?`; `POST/DELETE .../cover`; detail có `coverFileId`
- [x] UI: card cover trên Photos; Set/Remove cover trong AlbumView action sheet
- [x] Test API + smoke: ghim → đổi bìa; remove item bìa → fallback auto

**File chính:** `apps/api/migrations/00006_album_cover.*.sql`, `apps/api/internal/photo/*`, `apps/api/internal/platform/postgres/photo_store.go`, `apps/web/src/views/PhotosView.vue`, `apps/web/src/views/AlbumView.vue`, `apps/web/src/api/types.ts`.

### S3 — Favorites

- [x] Migration `favorites`
- [x] API: PUT/DELETE favorite; GET favorites
- [x] UI: action sheet star (FilesView, PhotoMediaSheet cho Photos/Album); segment Favorites | All ở Files; section Favorites ở Overview
- [x] Test API: trash ẩn khỏi list; permanent delete xóa row (`favorite_handler_test.go`, 4 case)

**File chính:** `apps/api/migrations/00007_favorites.*.sql`, `apps/api/internal/file/{model,port,service,handler,favorite_handler_test}.go`, `apps/api/internal/platform/postgres/file_store.go`, `apps/web/src/api/types.ts`, `apps/web/src/components/{AppIcon,PhotoMediaSheet}.vue`, `apps/web/src/views/{FilesView,PhotosView,AlbumView,OverviewView}.vue`.

### S4 — Public share link

- [x] **ADR bảo mật link** ([ADR 0004](decisions/0004-public-share-link-security.md): token 32B hash SHA-256, TTL NULL/1h/24h/7d, 404 chung, rate limit 30/phút/IP, cap 20 link/user, 1 link/file)
- [x] Migration `share_links` (kèm `recipient_user_id NULL`)
- [x] API owner: `POST/DELETE /files/{id}/share`, `GET /share-links`; public `GET /public/shares/{token}[/download]` (rate limit 429)
- [x] UI: ShareSheet (TTL segmented, copy, revoke), Settings "Shared links", trang public `/s/:token` ngoài shell
- [x] Test: `apps/api/internal/sharelink` (5 case: create/revoke, recreate, list, trash/purge, 404 chống dò)

**File chính:** `apps/api/migrations/00008_share_links.*.sql`, `apps/api/internal/sharelink/*`, `apps/api/internal/platform/postgres/share_link_store.go`, `apps/api/internal/platform/httpx/rate_limit.go`, `apps/web/src/components/ShareSheet.vue`, `apps/web/src/views/{FilesView,SettingsView,PublicShareView}.vue`.

### S6 — Activity log

- [x] Migration `activity_events` (index `owner_id, created_at DESC`)
- [x] Ghi fail-open tại upload / trash / restore / share create+revoke / password change (tầng service qua `ActivityRecorder`)
- [x] API `GET /activity?before=&limit=`; ticker dọn >90 ngày (nhịp 1 giờ, wire trong `cmd/api/main.go`)
- [x] Ghi đủ enum: bổ sung `folder.trashed`/`folder.restored` (tầng `folder`/`trash`) + `file.purged` với `targetName` — test case 6, go test 95/95 xanh; smoke live xác nhận cả 3 type
- [x] UI: section Activity trong Settings + Load more theo `nextBefore`
- [x] Test: `apps/api/internal/activity` (6 case: lifecycle, share/password, folder+purge, pagination, isolation, validation)

**File chính:** `apps/api/migrations/00009_activity_events.*.sql`, `apps/api/internal/activity/*`, `apps/api/internal/platform/postgres/{activity_store.go,schema_test.go}`, `apps/api/internal/{folder,file,trash,sharelink,share,auth}/service.go`, `apps/api/internal/ids/ids.go`, `apps/web/src/api/types.ts`, `apps/web/src/views/SettingsView.vue`.

### S5 — Share user nội bộ

- [x] Spec chi tiết §5.1–5.6 viết vào 09 trước khi code
- [x] API chống dò email: email chưa đăng ký → `201 {invited:true}` + mail mời (fail-open), không lộ account tồn tại; email tồn tại → 201 kèm `recipient`
- [x] Activity: `share.created` / `share.revoked` ghi cho share nội bộ (`targetName` = tên resource)
- [x] UI owner: "Share with user" trong action sheet file (`ShareUserSheet`); toast phân biệt share / invite
- [x] UI recipient: `/shared` (`SharedWithMeView`) — list incoming, download file, browse folder 1 cấp; entry từ Overview card thứ 3 (giữ 4 tab nav)
- [x] Test: `internal/share/internal_share_test.go` (3 case mới) + cập nhật case email lạ; go test 94/94 xanh; web lint/type-check/test 71/71 + build xanh

**File chính:** `apps/api/internal/share/{service.go,port.go,handler.go,internal_share_test.go}`, `apps/api/internal/share/handler_test.go`, `apps/api/internal/platform/postgres/share_store.go`, `apps/api/internal/app/app.go`, `apps/web/src/components/ShareUserSheet.vue`, `apps/web/src/views/{FilesView.vue,SharedWithMeView.vue}`, `apps/web/src/lib/shellNav.ts`, `apps/web/src/router/index.ts`, `apps/web/src/api/types.ts`, `apps/web/src/components/AppIcon.vue`.

---

## Smoke tay Phase 2 (tích lũy)

Sau mỗi slice, tick thêm (dùng `dev@filvault.com`):

- [x] S1: search + filter type/date/sort; chip × và Clear all; empty có filter
- [x] S2: album có/không cover; Set cover / Remove cover (smoke API 2026-08-23: auto = item mới nhất, pin/unpin, coverUrl xuất hiện khi bật image thumbnails)
- [x] S3: favorite/unfavorite; list Favorites; trash không hiện trong favorites (smoke API 2026-08-23: PUT 204, list có tên, DELETE 204)
- [x] S4: tạo link → mở ẩn danh tải được; revoke → 404 (smoke API 2026-08-23: public meta + download anonymous, sau revoke 404; hết hạn theo TTL unit test)
- [x] S6: Settings Activity hiện event sau upload/trash (smoke API 2026-08-23: feed chứa file.uploaded, share.created, share.revoked)
- [x] S5: share theo email user nội bộ; trang /shared thấy item nhận được; download file được share; email lạ → "Invitation sent" (smoke API 2026-08-23: invited=true + mail log, recipient payload, with-me, shared download, revoke / non-owner blocked / 409 / 400, activity share.created + share.revoked)

---

## Cố ý không làm ở Phase 2 đợt này

Không phải nợ của checklist này:

- share folder, quyền write, password-protect link
- `pg_trgm` / FTS, EXIF date, thumbnail server
- sync, versioning, CLI
- Terraform / RDS / deploy AWS

---

## Việc tiếp theo cho người nhận

1. Đọc [09](09-phase-2-features.md) §0 + § slice sắp làm; `DESIGN.md` §9a.
2. `make up` → login dev → xác nhận S1 vẫn ổn (filter search).
3. Bắt đầu **S2** theo TDD; tick DoD trong file này trước khi S3.
4. Trước S4: viết ADR bảo mật, merge/duyệt ADR, rồi mới migration/API.
5. Lệch spec → sửa 09 (và ADR nếu đổi kiến trúc), không code xong rồi viết ngược.
