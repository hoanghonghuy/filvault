# 10 — Sharing spec (Phase 5)

Trạng thái: **chốt slice đầu**. Nguồn API: [04-api.md](04-api.md). Schema: [03-database.md](03-database.md). Nghiệp vụ: [07-phase-1-business.md](07-phase-1-business.md).

Sharing cho phép user A chia sẻ **file** hoặc **folder** cho user B (theo email). B chỉ **đọc** (xem + tải), không sửa/xóa. Đây là slice đầu, chưa có link công khai, chưa có quyền ghi.

## 1. Phạm vi

```text
POST   /api/v1/shares                    # chia sẻ file/folder cho email
GET    /api/v1/shares                    # danh sách share mình tạo (outgoing)
GET    /api/v1/shares/with-me            # danh sách share nhận được (incoming)
DELETE /api/v1/shares/{id}               # thu hồi share (chỉ owner)
GET    /api/v1/shared/folders/{id}       # browse folder được share (recipient)
GET    /api/v1/shared/files/{id}/download # tải file được share (recipient)
```

Chưa làm: link công khai, quyền ghi, share cho nhóm, thông báo email, share folder đệ quy sâu (chỉ browse 1 cấp + tải file trực tiếp).

## 2. Data model

Bảng `shares` (migration mới `00003_shares`):

```text
id             CHAR(26) PK
owner_id       CHAR(26) NOT NULL  FK users(id)
resource_type  TEXT NOT NULL CHECK (resource_type IN ('file','folder'))
resource_id    CHAR(26) NOT NULL
recipient_id   CHAR(26) NOT NULL  FK users(id)
created_at     TIMESTAMPTZ NOT NULL
UNIQUE (owner_id, resource_type, resource_id, recipient_id)
```

- `resource_id` trỏ tới `files.id` hoặc `folders.id` tùy `resource_type` (không FK cứng vì hai bảng khác nhau).
- Share trùng (cùng owner, cùng resource, cùng recipient) → `CONFLICT`.
- Không share cho chính mình → `VALIDATION_ERROR`.
- Không share resource đã xóa (trash) → `NOT_FOUND`.

## 3. Quy tắc truy cập (recipient)

- File được share **trực tiếp** → recipient tải được.
- Folder được share → recipient browse được (liệt kê con trực tiếp) và tải file nằm **trực tiếp** trong folder đó.
- File nằm trong folder được share (con trực tiếp) → tải được.
- Recipient **không** thấy resource trong `browser`/`search` của mình; chỉ qua endpoint `/shared/*`.

## 4. Endpoint chi tiết

### 4.1 Tạo share

`POST /shares` `{ resourceType, resourceId, email }` → `201` `{ id, resourceType, resourceId, recipient: { id, email, displayName }, createdAt }`.

- `resourceType` ∈ `file` | `folder`.
- `email` lowercase, resolve sang user. Không tồn tại → `NOT_FOUND`.
- `email` = chính owner → `VALIDATION_ERROR`.
- Resource không thuộc owner / đã xóa → `NOT_FOUND`.
- Trùng share → `CONFLICT`.

### 4.2 Danh sách outgoing

`GET /shares` → `{ shares: [ { id, resourceType, resourceId, resourceName, recipient: {...}, createdAt } ] }`.

`resourceName` = tên file/folder hiện tại (để hiển thị).

### 4.3 Danh sách incoming

`GET /shares/with-me` → `{ shares: [ { id, resourceType, resourceId, resourceName, owner: {...}, createdAt } ] }`.

### 4.4 Thu hồi

`DELETE /shares/{id}` → `204`. Chỉ owner của share. Không phải owner → `NOT_FOUND`.

### 4.5 Browse folder được share

`GET /shared/folders/{id}` → giống `browser` nhưng cho recipient:

```json
{
  "folder": { "id", "name" },
  "folders": [ { "id", "name" } ],
  "files": [ { "id", "name", "mimeType", "sizeBytes" } ]
}
```

- Folder phải được share trực tiếp cho recipient → ngược lại `NOT_FOUND`.
- Chỉ liệt kê con trực tiếp (không đệ quy).

### 4.6 Tải file được share

`GET /shared/files/{id}/download` → `{ downloadUrl, expiresAt }`.

- File được share trực tiếp **hoặc** nằm trực tiếp trong folder được share → tải được.
- Ngược lại `NOT_FOUND`.

## 5. Lỗi & exit code (CLI)

- Parse body lỗi `{ error: { code, message } }` → in `code: message` ra stderr.
- Exit code: `0` thành công, `1` lỗi nghiệp vụ/HTTP, `2` sai usage.

## 6. CLI

```text
filvault share <name> <email>          # share file/folder theo tên (folder hiện tại)
filvault shares                        # danh sách share mình tạo
filvault shared-with-me                # danh sách share nhận được
filvault unshare <id>                  # thu hồi share theo id
filvault shared-ls <folderId>          # browse folder được share
filvault shared-download <fileId>      # tải file được share
```

## 7. Test

TDD theo slice. Ưu tiên:

| Tầng | Gì |
|---|---|
| Service | tạo share (resolve email, chặn self-share, trùng), revoke, list |
| Store | insert/query shares, unique violation |
| Handler | endpoint đúng + parse response (fake server) |
| CLI | `share`/`shares`/`shared-with-me`/`unshare`/`shared-ls`/`shared-download` gọi đúng endpoint |

## 8. Không làm ở slice này

```text
public link (share link)
quyền ghi / edit
share cho nhóm
thông báo email khi được share
share folder đệ quy sâu (browse nhiều cấp)
```
