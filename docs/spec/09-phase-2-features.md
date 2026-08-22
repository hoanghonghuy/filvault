# 09 — Spec tính năng Phase 2 (web)

Trạng thái: **đề xuất — chờ duyệt trước khi implement**. Không đụng kiến trúc Phase 1; chỉ thêm bảng/migration/API/UI theo slice dưới đây.

Nguồn nền: spec 03 (schema), spec 04 (API), spec 07 (nghiệp vụ), [`DESIGN.md`](../../DESIGN.md) (UI). Quy tắc cũ vẫn đúng: **thiết kế sẵn ≠ implement sẵn** — mỗi slice implement khi đến lượt, migration chỉ tạo khi slice đó bắt đầu.

## 0. Tổng quan slice

| # | Slice | Phụ thuộc | Độ lớn ước tính |
|---|---|---|---|
| S1 | Search nâng cao (filter + sort, giữ engine ILIKE) | không | nhỏ |
| S2 | Album cover (tự động + ghim tay) | không | nhỏ |
| S3 | Favorites (đánh dấu sao file/ảnh) | không | vừa |
| S4 | Share qua **public link** | không | lớn |
| S5 | Share cho **user nội bộ** (email, quyền read) | S4 | lớn |
| S6 | Activity log (hiển thị trong Settings) | không | nhỏ |

Thứ tự khuyến nghị: S1 → S2 → S3 → S4 → S6 → S5. S4 phải có ADR riêng về bảo mật link (làm khi bắt đầu slice).

---

## 1. S1 — Search nâng cao

### 1.1 Phạm vi

- Vẫn tìm theo **tên** (`ILIKE '%q%'`) — không đổi engine. `pg_trgm` / FTS: chỉ nâng khi dữ liệu thật cho thấy cần (ADR sẽ bổ sung lúc đó).
- Thêm **bộ lọc** và **sắp xếp**. Không search trong trash (giữ rule spec 04 §5).

### 1.2 API

```text
GET /api/v1/search?q=&type=&folderId=&from=&to=&sort=&order=&limit=
```

| Param | Giá trị | Mặc định |
|---|---|---|
| `q` | bắt buộc, 1–200 ký tự | — |
| `type` | `all` \| `image` \| `video` \| `document` \| `archive` \| `folder` | `all` |
| `folderId` | ULID — giới hạn trong cây con của folder này | omit = mọi nơi |
| `from` / `to` | ISO date (UTC) — lọc `created_at` file | omit |
| `sort` | `relevance` \| `name` \| `date` \| `size` | `relevance` |
| `order` | `asc` \| `desc` | `desc` |
| `limit` | 1–50 | 50 |

- `type=document` = nhóm PDF/office/text theo bảng allowlist spec 07 §3; `archive` = zip.
- Kết quả trả cùng shape cũ `{ folders, files }`; `folders` chỉ có khi `type=all|folder`.
- Filter không hợp lệ (`from > to`, `type` lạ, ULID sai) → `400 VALIDATION_ERROR`.
- `folderId` không tồn tại/không của user → `404` (giữ rule không dò ID).

### 1.3 UI

- Toolbar Files: ô search hiện tại + nút **filter chip** mở `FilterSheet` (bottom sheet mới, dùng `BottomSheet` sẵn có).
- Chip đang bật hiển thị ngay dưới search (vd `Video ×`, `This month ×`) — chạm × để bỏ filter đó, "Clear all" bỏ hết.
- Kết quả giữ list hiện tại; thêm caption "12 results · filtered" khi có filter.
- Empty state khi lọc ra 0 kết quả: "No results match your filters" + nút "Clear filters".
- Motion: sheet mở/đóng theo tokens P0; kết quả render qua `TransitionGroup name="row"` như list Files.

---

## 2. S2 — Album cover

### 2.1 Nghiệp vụ

- Mỗi album có **ảnh bìa**: `cover_file_id` (nullable) trỏ tới item ảnh/video trong album.
- Không ghim → API tự chọn bìa = item **mới nhất** trong album (deterministic, theo `created_at DESC, id DESC`).
- Ghim tay: user chọn "Set cover" từ action sheet của item trong album → ghi `cover_file_id`.
- Item bị remove khỏi album / trash / permanent delete mà đang là bìa → server **xóa `cover_file_id`** (fallback về auto).
- Ghim file không thuộc album → `404`/`VALIDATION_ERROR` (theo rule item hiện có).

### 2.2 API

```text
GET  /photos/albums            → thêm "coverUrl?" vào mỗi album
POST /photos/albums/{id}/cover { fileId }   → 204
DELETE /photos/albums/{id}/cover            → 204 (bỏ ghim, về auto)
```

- `coverUrl`: presign TTL ngắn như thumbnail (`ThumbnailPresignTTL`), chỉ trả khi bìa là ảnh và prefs cho phép; video → không có `coverUrl`, UI render placeholder video.
- `GET /photos/albums/{id}` trả thêm `coverFileId` (id của bìa đang dùng, kể cả auto) để UI đánh dấu item nào là bìa.

### 2.3 UI

- Album card (PhotosView): khung 1:1 phía trên tên — có `coverUrl` thì `<img>`, video thì placeholder icon video, album trống thì icon photos nền `surface-card`.
- Action sheet của item trong AlbumView thêm hàng **"Set cover"** (icon `image`) khi item là ảnh/video; nếu item đang là bìa thì label "Remove cover" (icon `restore`).
- Đổi bìa không có animation riêng; card cập nhật qua re-render thường (đã có `.appear` stagger khi vào trang).

---

## 3. S3 — Favorites

### 3.1 Nghiệp vụ

- User đánh dấu **sao** file (mọi loại) và item ảnh/video. Folder **không** favorite (giữ đơn giản).
- Favorite không đổi vị trí file; file vào trash → ẩn khỏi list favorites (row giữ); permanent delete → xóa row favorite (FK cascade).

### 3.2 API

```text
PUT    /api/v1/files/{id}/favorite    → 204
DELETE /api/v1/files/{id}/favorite    → 204
GET    /api/v1/files/favorites?limit= → { files: [...] }  (mới nhất trước, max 100)
```

- `PUT` khi đã favorite → vẫn `204` (idempotent). `DELETE` khi chưa → `204`.
- File không tồn tại/không của user/đang trash → `404`.
- `files[]` cùng shape `BrowserFile` + `favoritedAt`.

### 3.3 UI

- Action sheet file (FilesView) + item ảnh/video (Photos/Album): thêm **"Add to favorites" / "Remove from favorites"** (icon `star` / `star-filled`, không phải danger).
- FilesView: hàng **"Favorites"** thu gọn trên cùng khi đang ở root (chỉ hiện khi có ≥1 favorite), tap mở list favorites trong cùng view (segment "Favorites | All").
- Overview: thay/kèm "Recent files" bằng section "Favorites" khi có dữ liệu (giữ "Recent" làm fallback).
- Icon sao tô `warning` (#d97706) khi đã favorite — điểm nhấn màu duy nhất ngoài accent, đã có token.

---

## 4. S4 — Share qua public link

### 4.1 Nghiệp vụ

- User tạo **link công khai** cho 1 file (Phase này **chưa** share folder).
- Người nhận **không cần tài khoản**: mở link → trang public chỉ có preview/tải file đó, không thấy gì khác.
- Link có **hạn dùng tùy chọn** (không hạn / 1 giờ / 24 giờ / 7 ngày) và **có thể thu hồi** bất cứ lúc nào.
- File vào trash / bị permanent delete → link tự vô hiệu (server từ chối phục vụ).
- Vô hiệu hóa prefs thumbnail không ảnh hưởng link (link phục vụ original qua presign riêng).
- Giới hạn: tối đa **20 link hoạt động** mỗi user; mỗi file tối đa **1 link** (tạo lại = thu hồi link cũ).

### 4.2 API (owner)

```text
POST   /api/v1/files/{id}/share        { expiresIn?: "1h"|"24h"|"7d"|null } → 201
DELETE /api/v1/files/{id}/share        → 204 (thu hồi)
GET    /api/v1/shares                  → { shares: [...] }  (link đang hoạt động)
```

`201` trả:

```json
{
  "id": "...",
  "fileId": "...",
  "url": "https://<host>/s/<token>",
  "expiresAt": null,
  "createdAt": "..."
}
```

- `token`: 32 byte random, URL-safe; DB lưu **hash** (SHA-256) — lộ DB không lộ link (cùng triết lý refresh token spec 03 §3.2).
- File không `READY` / đang trash → `INVALID_STATE`.

### 4.3 API (public, không Bearer)

```text
GET /api/v1/public/shares/{token}           → metadata tối thiểu
GET /api/v1/public/shares/{token}/download  → { downloadUrl } (presign 5 phút)
```

- Metadata chỉ gồm: `name`, `mimeType`, `sizeBytes`, `expiresAt`. **Không** trả id nội bộ, owner, folder.
- Token sai / thu hồi / hết hạn / file đã trash → `404` chung một mã (không phân biệt nguyên nhân, chống dò).
- Public endpoint có **rate limit** đơn giản theo IP (ví dụ 30 req/phút; hằng số config).
- Trang public `/s/:token` (web, không cần login, không shell): thẻ card giữa màn — tên, loại, kích thước, nút **Download**; đúng ngôn ngữ visual DESIGN.md nhưng **không** bottom nav.

### 4.4 UI (owner)

- Action sheet file: thêm **"Share link"** (icon `share`) → `ShareSheet` mới:
  - Chưa có link: chọn hạn dùng (segmented: Forever / 1h / 24h / 7d) + nút "Create link".
  - Đã có link: hiện URL (nút Copy, dùng `navigator.clipboard` + toast "Link copied"), hạn hết hạn, nút **"Revoke link"** (danger).
- Settings: section **"Shared links"** liệt kê link đang hoạt động (tên file, hạn, thu hồi) — chỗ quản lý tập trung.
- Toast/confirm đều dùng sheet sẵn có; revoke phải confirm.

---

## 5. S5 — Share cho user nội bộ

- **Chỉ làm sau S4.** Chia sẻ file cho user Filvault khác bằng email, quyền **read** (xem + tải).
- Người nhận thấy mục "Shared with me" trong web (nav thứ 5 hoặc entry từ Overview — chốt khi implement UI).
- Reuse bảng `shares` của S4, thêm `recipient_user_id` (null = public link). Quyền check ở mọi endpoint đọc: owner **hoặc** recipient.
- Email người nhận không tồn tại → vẫn `201` + gửi mail mời (chống dò email, cùng triết lý resend spec 04 §2).
- Chi tiết UI/endpoint viết bổ sung khi bắt đầu slice — phần này chỉ **giành chỗ** trong schema.

---

## 6. S6 — Activity log

### 6.1 Nghiệp vụ

- Ghi sự kiện **quan trọng** của chính user: upload hoàn tất, xóa vào trash, restore, xóa vĩnh viễn, tạo/thu hồi share link, đổi mật khẩu, đổi settings.
- Không ghi: đọc/list/search/download (nhiễu, không giá trị).
- Retention: giữ **90 ngày**, dọn bằng ticker sẵn có của trash (cùng nhịp 1 giờ).

### 6.2 API

```text
GET /api/v1/activity?before=&limit=   → { events: [...], nextBefore? }
```

```json
{
  "events": [
    { "id": "...", "type": "file.uploaded", "targetName": "a.jpg", "createdAt": "..." }
  ]
}
```

- `type` enum chốt khi implement; mỗi event tối đa 1 `targetName` (đơn giản, đủ đọc).
- Chỉ owner đọc được activity của mình. Ghi log **fail-open**: lỗi ghi log không làm fail request chính.

### 6.3 UI

- Settings: section **"Activity"** — list đơn giản (icon theo type + `targetName` + thời gian tương đối), "Load more" theo `nextBefore` (cùng pattern timeline Photos).
- Không thêm tab mới vào bottom nav (giữ rule 4 tab của DESIGN.md §5).

---

## 7. Database (tổng hợp migration dự kiến)

Chỉ tạo khi slice tương ứng bắt đầu:

```text
S2: ALTER TABLE albums ADD COLUMN cover_file_id CHAR(26) NULL REFERENCES files(id)
S3: CREATE TABLE favorites (id, owner_id, file_id, created_at, UNIQUE(owner_id, file_id))
S4: CREATE TABLE share_links (
      id, owner_id, file_id, token_hash UNIQUE, expires_at NULL,
      revoked_at NULL, created_at,
      recipient_user_id NULL   ← giành chỗ cho S5, Phase này luôn NULL
    )
S6: CREATE TABLE activity_events (id, owner_id, type, target_name, created_at)
```

Mọi bảng đều `owner_id FK users(id)`; index tối thiểu: `favorites(owner_id, created_at DESC)`, `share_links(token_hash)`, `activity_events(owner_id, created_at DESC)`.

---

## 8. Không có ở Phase 2 (đợt này)

```text
share folder / thư mục
share có quyền write / edit
password-protect link
pg_trgm / full-text search
EXIF date, thumbnail server-side
sync, versioning, CLI
```

Mỗi mục trên mở lại bằng ADR riêng khi có nhu cầu thật.
