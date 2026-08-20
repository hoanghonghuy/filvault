# 09 — CLI spec (Phase 4)

Trạng thái: **chốt slice đầu**. Nguồn API: [04-api.md](04-api.md). Kiến trúc: [05-architecture.md](05-architecture.md).

CLI `filvault` nói chuyện với API **qua HTTP** (`/api/v1`), không import package `internal/` của API. Đây là binary độc lập, cùng monorepo.

## 1. Phạm vi slice đầu (core)

```text
filvault login                 # email + password → lưu token
filvault whoami                # GET /users/me
filvault ls [folderId]         # GET /browser (root nếu bỏ trống)
filvault upload <file> [--folder <id>]
filvault download <name>       # tìm theo tên trong folder hiện tại
```

Chưa làm ở slice này: `cd`, `mkdir`, `rm`, `sync`, glob `*.pdf`, sharing, versioning.

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

- `login` ghi đè file. `logout` (slice sau) xóa.
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

### 4.2 whoami

- `GET /users/me` → in `displayName <email>` + `storageUsed/storageQuota` (dùng `formatBytes`).

### 4.3 ls

```text
filvault ls [folderId]
```

- `GET /browser?folderId=` (omit nếu trống) → in folders trước (prefix `/`), rồi files (tên + size).
- Không có folderId → root.

### 4.4 upload

```text
filvault upload <file> [--folder <id>]
```

1. Đọc file local, lấy `size` (bytes) + `contentType` (`mime.TypeByExtension`, fallback `application/octet-stream`).
2. `POST /files/upload-sessions` `{ name, size, contentType, folderId }` → `201` `{ fileId, uploadUrl, expiresAt }`.
3. `PUT uploadUrl` thẳng object storage, header `Content-Type` khớp `contentType`.
4. `POST /files/{fileId}/complete` → `200` metadata.
5. In `Uploaded <name>`.

Lỗi: ngoài allowlist / quá 100 MiB / vượt quota → in `code` + `message` từ body lỗi, exit 1.

### 4.5 download

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
- Exit code: `0` thành công, `1` lỗi nghiệp vụ/HTTP, `2` sai usage (thiếu arg).

## 6. Test

TDD theo slice. Ưu tiên:

| Tầng | Gì |
|---|---|
| HTTP client | parse lỗi, refresh-on-401 retry (dùng `httptest.Server` fake) |
| Token store | đọc/ghi file, path theo OS |
| Command | `ls`/`whoami`/`upload`/`download` gọi đúng endpoint + parse response (fake server) |

Không cần test thật MinIO/S3 — upload/download dùng fake server trả presigned URL.

## 7. Không làm ở slice này

```text
cd / mkdir / rm / mv
sync, watcher, hash
glob, resumable, multipart
sharing, versioning, devices
```
