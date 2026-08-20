# 09 — CLI spec (Phase 4)

Trạng thái: **slice 1 + slice 2**. Nguồn API: [04-api.md](04-api.md). Kiến trúc: [05-architecture.md](05-architecture.md).

CLI `filvault` nói chuyện với API **qua HTTP** (`/api/v1`), không import package `internal/` của API. Đây là binary độc lập, cùng monorepo.

## 1. Phạm vi slice đầu (core)

```text
filvault login                 # email + password → lưu token
filvault logout                # revoke refresh token + xóa token file
filvault whoami                # GET /users/me
filvault ls [folderId]         # GET /browser (root nếu bỏ trống)
filvault upload <file> [--folder <id>]
filvault download <name>       # tìm theo tên trong folder hiện tại
filvault --help                # in usage ra stdout, exit 0
filvault --version             # in version, exit 0
```

## 1b. Phạm vi slice 2 (quản lý file/folder)

```text
filvault mkdir <name> [--parent <id>]   # POST /folders
filvault rm <name>                      # soft delete file/folder theo tên (root)
filvault mv <name> <newName>            # rename file/folder
filvault mv <name> --to <folderId>      # move file/folder sang folder
filvault trash                          # GET /trash
filvault restore <name>                 # restore file/folder từ trash theo tên
filvault search <q>                     # GET /search?q=
filvault storage                        # GET /storage
```

## 1c. Phạm vi slice 3 (điều hướng thư mục)

```text
filvault cd <name>     # vào folder theo tên (trong folder hiện tại)
filvault cd ..         # lên folder cha
filvault cd            # về root (cùng với cd /)
filvault pwd           # in folder hiện tại (path hoặc id)
```

- Config lưu thêm `currentFolderId` (rỗng = root).
- `ls` (không arg), `upload`, `download`, `rm`, `mv`, `mkdir` hoạt động **trong folder hiện tại**.
- `ls <folderId>` vẫn liệt kê folder cụ thể (không đổi `currentFolderId`).

Chưa làm: `sync`, glob `*.pdf`, sharing, versioning, permanent delete (`rm -f`).

## 2. Ngôn ngữ & module

- Go, module riêng `apps/cli` (module path `filvault/cli`), `go 1.25`.
- Không import AWS SDK, không import `apps/api/internal/*`.
- HTTP client dùng stdlib `net/http`. JSON stdlib `encoding/json`.
- CLI framework: **stdlib `flag`** (KISS, không thêm dependency). Subcommand dispatch thủ công.
- Dependency duy nhất: `golang.org/x/term` (ẩn password khi gõ).

## 3. Config & token

| Nguồn | Giá trị |
|---|---|
| `FILVAULT_API_BASE` | mặc định `http://localhost:8080/api/v1` |
| Token file | `~/.config/filvault/config.json` (hoặc `%APPDATA%\filvault\config.json` trên Windows) |

Token file JSON:

```json
{
  "apiBase": "http://localhost:8080/api/v1",
  "accessToken": "...",
  "refreshToken": "..."
}
```

- `login` ghi đè file. `logout` revoke refresh token rồi xóa token trong file.
- Access hết hạn (401) → tự `POST /auth/refresh` một lần rồi retry, giống web client. Refresh fail → xóa token, báo phải `login` lại.

## 4. Lệnh chi tiết

### 4.1 login

```text
filvault login
  email:    (nhập, không echo password)
  password: (nhập ẩn)
```

- `POST /auth/login` `{ email, password }` → `200` `{ user, accessToken, refreshToken }`.
- Lưu token file. In `Logged in as <email>`.
- Password đọc ẩn qua `golang.org/x/term` (không echo); fallback stdin thường khi không có TTY (piped input).
- Sai mật khẩu → `401 UNAUTHORIZED` → in message, exit 1.
- Chưa verify: login vẫn `200` (spec 04 §2) — CLI in cảnh báo `Email not verified` nhưng vẫn lưu token (để `whoami` chạy được).

### 4.2 logout

- `POST /auth/logout` `{ refreshToken }` → `204`.
- Xóa `accessToken`/`refreshToken` trong token file. In `Logged out`.

### 4.3 whoami

- `GET /users/me` → in `displayName <email>` + `storageUsed/storageQuota` (dùng `formatBytes`).

### 4.4 ls

```text
filvault ls [folderId]
```

- `GET /browser?folderId=` (omit nếu trống) → in folders trước (prefix `/`), rồi files (tên + size).
- Không có folderId → root.

### 4.5 upload

```text
filvault upload <file> [--folder <id>]
```

1. Đọc file local, lấy `size` (bytes) + `contentType` (`mime.TypeByExtension`, fallback `application/octet-stream`).
2. `POST /files/upload-sessions` `{ name, size, contentType, folderId }` → `201` `{ fileId, uploadUrl, expiresAt }`.
3. `PUT uploadUrl` thẳng object storage, header `Content-Type` khớp `contentType`.
4. `POST /files/{fileId}/complete` → `200` metadata.
5. In `Uploaded <name>`.

Lỗi: ngoài allowlist / quá 100 MiB / vượt quota → in `code` + `message` từ body lỗi, exit 1.

### 4.6 download

```text
filvault download <name>
```

1. `GET /browser` (folder hiện tại) → tìm file theo `name` (khớp chính xác).
2. Không thấy → `NOT_FOUND`, exit 1.
3. `GET /files/{id}/download` → `{ downloadUrl, expiresAt }`.
4. `GET downloadUrl` → ghi bytes ra file local cùng tên.
5. In `Downloaded <name>`.

## 5. Lỗi & exit code

- Parse body lỗi `{ error: { code, message } }` → in `code: message` ra stderr.
- Exit code: `0` thành công, `1` lỗi nghiệp vụ/HTTP, `2` sai usage (thiếu arg / lệnh không rõ).
- `--help` / `--version` in ra **stdout** và exit `0` (không phải lỗi).

## 5b. Lệnh slice 2 chi tiết

### 5b.1 mkdir

```text
filvault mkdir <name> [--parent <id>]
```

- `POST /folders` `{ name, parentId? }` → `201`. In `Created folder <name>`.
- `parentId` omit = root. Tên trùng sibling → `409 CONFLICT` (in message, exit 1).

### 5b.2 rm

```text
filvault rm <name>
```

- `GET /browser` (root) → tìm theo `name` (khớp chính xác) trong folders rồi files.
- Folder → `DELETE /folders/{id}`; file → `DELETE /files/{id}`. Cả hai soft delete.
- Không thấy → `NOT_FOUND`, exit 1. In `Deleted <name>`.

### 5b.3 mv

```text
filvault mv <name> <newName>          # rename
filvault mv <name> --to <folderId>    # move
```

- `GET /browser` → resolve `name` → `(type, id)`.
- Rename: folder → `PATCH /folders/{id}` `{ name }`; file → `PATCH /files/{id}` `{ name }`.
- Move: folder → `PATCH /folders/{id}` `{ parentId }`; file → `PATCH /files/{id}` `{ folderId }`.
- In `Moved <name>`.

### 5b.4 trash

```text
filvault trash
```

- `GET /trash` → in folders trước (prefix `/`), rồi files (tên + size). Không có gì → in `(empty)`.

### 5b.5 restore

```text
filvault restore <name>
```

- `GET /trash` → tìm theo `name` (khớp chính xác) trong folders rồi files.
- Folder → `POST /folders/{id}/restore`; file → `POST /files/{id}/restore`.
- Không thấy → `NOT_FOUND`, exit 1. In `Restored <name>`.

### 5b.6 search

```text
filvault search <q>
```

- `GET /search?q=` → in folders (prefix `/`) rồi files (tên + size).

### 5b.7 storage

```text
filvault storage
```

- `GET /storage` → in `usedBytes of quotaBytes` (dùng `formatBytes`).

## 5c. Lệnh slice 3 chi tiết

### 5c.1 cd

```text
filvault cd <name>     # vào folder con theo tên
filvault cd ..         # lên cha
filvault cd            # về root
```

- `cd <name>`: `GET /browser?folderId=<current>` → tìm folder con theo `name` (khớp chính xác). Không thấy → `NOT_FOUND`, exit 1. Thấy → lưu `currentFolderId` = id folder đó.
- `cd ..`: `GET /browser?folderId=<current>` → lấy `parentId` của folder hiện tại (từ field `folder.parentId`). Lưu `currentFolderId` = parentId (rỗng = root).
- `cd` (không arg): xóa `currentFolderId` (về root).
- In `Now in <name>` (hoặc `Now in /` khi về root).

### 5c.2 pwd

```text
filvault pwd
```

- In `currentFolderId` nếu có, ngược lại in `/`.

## 6. Test

TDD theo slice. Ưu tiên:

| Tầng | Gì |
|---|---|
| HTTP client | parse lỗi, refresh-on-401 retry (dùng `httptest.Server` fake) |
| Token store | đọc/ghi file, path theo OS |
| Command | `login`/`logout`/`ls`/`whoami`/`upload`/`download`/`mkdir`/`rm`/`mv`/`trash`/`restore`/`search`/`storage` gọi đúng endpoint + parse response (fake server) |

Không cần test thật MinIO/S3 — upload/download dùng fake server trả presigned URL.

## 7. Không làm ở slice này

```text
sync, watcher, hash
glob, resumable, multipart
sharing, versioning, devices
permanent delete (rm -f)
```
