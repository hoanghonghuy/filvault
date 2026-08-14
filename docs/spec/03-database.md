# 03 — Database spec (Phase 1)

Trạng thái: **chốt cho Phase 1**. Bảng phase sau **không** tạo sẵn.

Adapter mặc định: `postgres` (Docker local; RDS cùng schema). DynamoDB: xem [02-provider-strategy.md](02-provider-strategy.md) §4 — không có bảng DynamoDB ở phase này.

## 1. Quy ước

| Mục | Quyết định |
|---|---|
| Engine | PostgreSQL 16+ |
| ID | ULID (26 ký tự, sortable theo thời gian), `CHAR(26)` |
| Thời gian | `TIMESTAMPTZ`, UTC |
| Tên cột | `snake_case` |
| Xóa Phase 1 | Soft delete qua `deleted_at` |
| Migration | SQL tuần tự trong `apps/api/migrations/` (up/down) |
| Email | Lưu lowercase; `UNIQUE` |

Không dùng UUID v4 làm PK (không sort theo thời gian). Không dùng tên file làm khóa.

## 2. File status

Phase 1 chỉ dùng:

```text
PENDING  →  đã tạo metadata, chưa confirm object
READY    →  object có, hiện trong My Files
FAILED   →  confirm thất bại / hết hạn
```

`PROCESSING` / `UPLOADED` **dành chỗ** (cột `status` cho phép sau này), không dùng trong Phase 1. File `PENDING`/`FAILED` không hiện ở My Files. Trash chỉ chứa bản từng `READY` rồi soft-delete.

## 3. Bảng Phase 1

### 3.1 `users`

```text
id               CHAR(26) PK
email            TEXT NOT NULL UNIQUE
display_name     TEXT NOT NULL
password_hash    TEXT NOT NULL
email_verified_at          TIMESTAMPTZ NULL
verification_code_hash     TEXT NULL
verification_expires_at    TIMESTAMPTZ NULL
storage_used     BIGINT NOT NULL DEFAULT 0  CHECK (storage_used >= 0)
storage_quota    BIGINT NOT NULL            CHECK (storage_quota > 0)
trash_auto_delete_enabled  BOOLEAN NOT NULL
trash_retention_days       INT NOT NULL CHECK (trash_retention_days >= 1)
created_at       TIMESTAMPTZ NOT NULL
updated_at       TIMESTAMPTZ NOT NULL
```

Quota mặc định: **10 GiB** (`10737418240` bytes) — hằng số `DefaultStorageQuotaBytes`.

Trash default lúc register: copy từ config hệ thống (`DefaultTrashAutoDeleteEnabled` gợi ý `false`, `DefaultTrashRetentionDays` gợi ý `30`).

`email_verified_at IS NULL` = chưa verify. Mã verify: hash, TTL 15 phút; resend ghi đè.

Nhiều row `refresh_tokens` / user = nhiều máy. Logout một máy chỉ `revoked_at` token đó.

### 3.2 `refresh_tokens`

Logout / rotation cần lưu server-side. Không nhét refresh token thuần vào JWT stateless.

```text
id            CHAR(26) PK
user_id       CHAR(26) NOT NULL  FK users(id) ON DELETE CASCADE
token_hash    TEXT NOT NULL UNIQUE
expires_at    TIMESTAMPTZ NOT NULL
revoked_at    TIMESTAMPTZ NULL
created_at    TIMESTAMPTZ NOT NULL
```

Lưu **hash** của refresh token, không lưu plaintext.

### 3.3 `folders`

```text
id            CHAR(26) PK
owner_id      CHAR(26) NOT NULL  FK users(id)
parent_id     CHAR(26) NULL      FK folders(id)
name          TEXT NOT NULL
created_at    TIMESTAMPTZ NOT NULL
updated_at    TIMESTAMPTZ NOT NULL
deleted_at    TIMESTAMPTZ NULL
```

`parent_id IS NULL` = nằm ở root (“My Files”). **Không** tạo row folder root giả.

Ràng buộc:

- `owner_id` của folder con phải trùng folder cha (enforce trong service + có thể trigger sau).
- Không cho `parent_id = id`.
- Tên **không trùng** trong cùng parent, cùng owner, đang không xóa:

```text
UNIQUE (owner_id, parent_id, name)  WHERE deleted_at IS NULL AND parent_id IS NOT NULL
UNIQUE (owner_id, name)             WHERE deleted_at IS NULL AND parent_id IS NULL
```

Index list:

```text
(owner_id, parent_id) WHERE deleted_at IS NULL
(owner_id)            WHERE deleted_at IS NOT NULL   -- Trash
```

### 3.4 `files`

Không tách bảng `upload_sessions` ở Phase 1. File `status = PENDING` **là** upload session.

```text
id              CHAR(26) PK
owner_id        CHAR(26) NOT NULL  FK users(id)
folder_id       CHAR(26) NULL      FK folders(id)
name            TEXT NOT NULL
original_name   TEXT NOT NULL
object_key      TEXT NOT NULL UNIQUE
mime_type       TEXT NOT NULL
size_bytes      BIGINT NOT NULL CHECK (size_bytes >= 0)
status          TEXT NOT NULL
created_at      TIMESTAMPTZ NOT NULL
updated_at      TIMESTAMPTZ NOT NULL
deleted_at      TIMESTAMPTZ NULL
upload_expires_at  TIMESTAMPTZ NULL   -- hạn presign; NULL khi READY
```

`folder_id IS NULL` = file ở root. `original_name` = tên lúc upload; `name` đổi khi rename.

`object_key` format (không chứa tên file):

```text
users/{userId}/files/{fileId}
```

Tên không trùng cùng folder, cùng owner, chưa xóa — cùng kiểu unique index như folders.

Index:

```text
(owner_id, folder_id) WHERE deleted_at IS NULL AND status = 'READY'
(owner_id)            WHERE deleted_at IS NOT NULL
(status, upload_expires_at) WHERE status = 'PENDING'
```

`size_bytes` lúc `PENDING` = size client khai (để chặn quota sớm). Lúc confirm: **ghi đè** bằng size thật từ `ObjectStore.Head`. Không tin client.

Allowlist + trần 100 MiB: [07-phase-1-business.md](07-phase-1-business.md) §3. Cột `mime_type` chỉ nhận giá trị trong bảng đó.

Index thêm cho Photos:

```text
(owner_id, created_at DESC) WHERE deleted_at IS NULL AND status = 'READY'
```

### 3.5 `albums`

```text
id            CHAR(26) PK
owner_id      CHAR(26) NOT NULL  FK users(id)
name          TEXT NOT NULL
created_at    TIMESTAMPTZ NOT NULL
updated_at    TIMESTAMPTZ NOT NULL
```

Tên không trùng trong cùng owner:

```text
UNIQUE (owner_id, name)
```

Xóa album = xóa row (file không đụng). Không soft-delete album Phase 1.

### 3.6 `album_items`

```text
album_id      CHAR(26) NOT NULL  FK albums(id) ON DELETE CASCADE
file_id       CHAR(26) NOT NULL  FK files(id) ON DELETE CASCADE
position      INT NOT NULL DEFAULT 0
added_at      TIMESTAMPTZ NOT NULL
PRIMARY KEY (album_id, file_id)
```

Chỉ file ảnh/video, `READY`, cùng `owner_id`. Enforce trong service.

## 4. Hành vi xóa / quota

**Xóa file:** set `deleted_at`. Không xóa object. `storage_used` **giữ nguyên** khi vào Trash (file vẫn chiếm quota). Permanent delete mới: xóa object + xóa row (hoặc archive) + giảm `storage_used`.

**Xóa folder Phase 1:** chỉ khi **không còn con chưa xóa** (file `READY` hoặc folder). Không cascade. Cascade trash là phase sau.

**Restore:** `deleted_at = NULL`. Nếu tên trùng sibling đang sống → `CONFLICT`, client phải rename.

**Quota:** `storage_used + claimed_size <= storage_quota` trước khi tạo `PENDING`. Confirm: nếu size thật lớn hơn claimed và vượt quota → `FAILED`, không tăng `storage_used`. Thành công: `storage_used += size thật` (không cộng claimed).

File `PENDING` hết `upload_expires_at` mà chưa confirm: coi `FAILED`; không cộng quota. Confirm phải reject session hết hạn.

**Tự xóa trash:** nếu user `trash_auto_delete_enabled`, ticker API xóa vĩnh viễn khi `now - deleted_at >= trash_retention_days`. File trước, folder trống sau. Chi tiết: spec 07 §5.

## 5. Search Phase 1

Không FTS, không trigram bắt buộc.

```text
files.name   ILIKE '%q%'  AND owner_id = ? AND deleted_at IS NULL AND status = 'READY'
folders.name ILIKE '%q%'  AND owner_id = ? AND deleted_at IS NULL
```

`q` rỗng → không search, trả validation error. Giới hạn kết quả (ví dụ 50). `pg_trgm` / OpenSearch: sau này.

## 6. Bảng **không** tạo ở Phase 1

```text
file_versions
shares
share_links
devices
sync_states
favorites
audit_logs
upload_sessions   ← dùng files PENDING
tenant / app / api_keys   ← tích hợp app khác, sau này
```

Khi cần: migration mới. Không tạo bảng trống “cho tương lai”.

## 7. Ownership

Mọi query file/folder **luôn** kèm `owner_id = current_user`. Không `SELECT ... WHERE id = ?` rồi tin client. ID lộ trên URL không phải giấy phép.

## 8. Mapping DynamoDB (nhắc lại)

Schema này là của Postgres. DynamoDB **không** clone 1:1. Khi làm adapter đó, dịch **thao tác repository**, không dịch bảng. Chi tiết: spec 02 §4.
