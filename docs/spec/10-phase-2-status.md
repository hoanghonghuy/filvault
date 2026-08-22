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
| S2 Album cover | Chưa |
| S3 Favorites | Chưa |
| S4 Public share link | Chưa (cần ADR bảo mật trước) |
| S6 Activity log | Chưa |
| S5 Share user nội bộ | Chưa (sau S4; cần bổ sung spec chi tiết) |

→ **Việc tiếp theo:** bắt đầu **S2** theo TDD (xem 09 §0.1). Không nhảy S4 trước ADR.

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
| Dev login | `dev@filvault.com` / `Dev1234@` |

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

- [ ] Migration `albums.cover_file_id`
- [ ] API: list albums có `coverUrl?`; `POST/DELETE .../cover`; detail có `coverFileId`
- [ ] UI: card cover trên Photos; Set/Remove cover trong AlbumView action sheet
- [ ] Test API + smoke: ghim → đổi bìa; remove item bìa → fallback auto

### S3 — Favorites

- [ ] Migration `favorites`
- [ ] API: PUT/DELETE favorite; GET favorites
- [ ] UI: action sheet star; segment Favorites | All ở Files root; Overview ưu tiên Favorites
- [ ] Test API + smoke: trash ẩn khỏi list; permanent delete xóa row

### S4 — Public share link

- [ ] **ADR bảo mật link** (TTL, rate limit, 404 chung) — **bắt buộc trước code**
- [ ] Migration `share_links` (kèm `recipient_user_id NULL`)
- [ ] API owner + public `GET /s/:token`
- [ ] UI: ShareSheet, Settings “Shared links”, trang public `/s/:token`
- [ ] Test: token hash, revoke, hết hạn, rate limit, ownership

### S6 — Activity log

- [ ] Migration `activity_events`
- [ ] Ghi fail-open tại upload / trash / restore / share / password / settings
- [ ] API `GET /activity`; ticker dọn >90 ngày
- [ ] UI: section Activity trong Settings + Load more
- [ ] Test ghi/đọc + smoke Settings

### S5 — Share user nội bộ

- [ ] Bổ sung chi tiết API/UI vào 09 (hoặc file con) trước khi code
- [ ] Reuse `share_links.recipient_user_id`; quyền read-only
- [ ] Test + smoke

---

## Smoke tay Phase 2 (tích lũy)

Sau mỗi slice, tick thêm (dùng `dev@filvault.com`):

- [x] S1: search + filter type/date/sort; chip × và Clear all; empty có filter
- [ ] S2: album có/không cover; Set cover / Remove cover
- [ ] S3: favorite/unfavorite; list Favorites; trash không hiện trong favorites
- [ ] S4: tạo link → mở ẩn danh tải được; revoke → 404; hết hạn → 404
- [ ] S6: Settings Activity hiện event sau upload/trash
- [ ] S5: share theo email user nội bộ (khi có)

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
