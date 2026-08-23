# 11 — Versioning spec (Phase 6)

Trạng thái: **chốt slice đầu; đã implement + smoke thật (2026-08-23)** — API, CLI (`upload --replace`, `versions`), UI. Nguồn API: [04-api.md](04-api.md). Schema: [03-database.md](03-database.md).

Versioning giữ lịch sử các phiên bản cũ của một file khi file bị **thay thế** (replace). File giữ nguyên `id` và `name`; mỗi lần replace, bản cũ được lưu vào `file_versions`, bản mới trở thành bản hiện tại.

## 1. Quyết định thiết kế

- **Replace-on-upload**: thay thế file bằng cách upload file mới cùng tên, không phải snapshot thủ công.
- **File giữ nguyên `id`**: bản hiện tại vẫn là row `files` cũ; chỉ `object_key`, `size_bytes`, `mime_type` đổi.
- **Quota**: mọi phiên bản đều chiếm quota. `storage_used` đã tính bản cũ; khi replace chỉ cộng thêm size bản mới (bản cũ vẫn giữ).
- **Không** xóa phiên bản riêng lẻ ở slice này. Restore phiên bản = slice sau.

## 2. Data model

Migration `00004_file_versions`:

```text
file_versions
    id          CHAR(26) PK
    file_id     CHAR(26) NOT NULL  FK files(id) ON DELETE CASCADE
    object_key  TEXT NOT NULL
    size_bytes  BIGINT NOT NULL CHECK (size_bytes >= 0)
    mime_type   TEXT NOT NULL
    created_at  TIMESTAMPTZ NOT NULL

files.replaces_file_id  CHAR(26) NULL   -- chỉ có nghĩa cho file PENDING
```

- `replaces_file_id` trỏ tới file `READY` sẽ bị thay thế. Chỉ set khi tạo upload session có `replaceFileId`.
- Khi complete thành công: archive bản cũ vào `file_versions` (file_id = replaces_file_id), cập nhật row file cũ, xóa row PENDING.

## 3. Endpoint

### 3.1 Tạo session (mở rộng)

`POST /files/upload-sessions` thêm field tùy chọn `replaceFileId`:

```json
{ "name": "a.pdf", "size": 100, "contentType": "application/pdf", "folderId": null, "replaceFileId": "..." }
```

- `replaceFileId` có → file đích phải `READY`, thuộc owner, chưa trash → ngược lại `NOT_FOUND`.
- Bỏ qua kiểm tra trùng tên (đang thay thế chính file đó).
- `name` của session phải khớp `name` file đích → ngược lại `VALIDATION_ERROR`.

### 3.2 Complete (mở rộng)

Khi file PENDING có `replaces_file_id`:

1. Archive bản cũ: `INSERT file_versions (file_id, object_key, size_bytes, mime_type, created_at)` từ file đích.
2. Cập nhật file đích: `object_key`, `size_bytes`, `mime_type` = bản mới, `status = READY`, `updated_at`.
3. Xóa row PENDING.
4. `storage_used += size bản mới` (bản cũ đã tính, giữ nguyên).

### 3.3 Danh sách phiên bản

`GET /files/{id}/versions` → `{ "versions": [ { "id", "sizeBytes", "mimeType", "createdAt" } ] }`.

- File phải thuộc owner, `READY`, chưa trash → ngược lại `NOT_FOUND`.
- Sắp xếp `created_at DESC` (mới nhất trước).

### 3.4 Tải phiên bản

`GET /files/{id}/versions/{versionId}/download` → `{ "downloadUrl", "expiresAt" }`.

- File thuộc owner, version thuộc file đó → ngược lại `NOT_FOUND`.

## 4. Lỗi & exit code (CLI)

- Parse body lỗi `{ error: { code, message } }` → in `code: message` ra stderr.
- Exit code: `0` thành công, `1` lỗi nghiệp vụ/HTTP, `2` sai usage.

## 5. CLI

```text
filvault upload --replace <name> <file>   # thay thế file theo tên (folder hiện tại); flag đứng trước positional (Go flag)
filvault versions <name>                  # liệt kê phiên bản file theo tên
filvault version-download <name> <versionId>  # tải một phiên bản
```

## 6. Test

| Tầng | Gì |
|---|---|
| Service (file) | replace: archive bản cũ, cập nhật file, quota; list versions; download version |
| Store | insert/query file_versions, cập nhật replaces_file_id |
| Handler | endpoint đúng + parse response (fake server) |
| CLI | `upload --replace`, `versions`, `version-download` gọi đúng endpoint |

> Ghi chú triển khai: versioning được tích hợp vào package `internal/file` (không tạo package riêng) vì `file_versions` và `replaces_file_id` gắn trực tiếp với vòng đời file; tách package sẽ tạo circular dependency.

## 7. Không làm ở slice này

```text
restore phiên bản (promote bản cũ thành hiện tại)
xóa phiên bản riêng lẻ
giới hạn số phiên bản / retention
versioning cho folder
```
