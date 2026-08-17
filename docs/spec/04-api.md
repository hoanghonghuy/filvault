# 04 — API spec (Phase 1)

Trạng thái: **chốt cho Phase 1**. Endpoint sharing / sync / devices **không** có ở phase này.

JSON camelCase. Auth: `Authorization: Bearer <access_token>`. Prefix: `/api/v1`.

## 1. Quy ước

### 1.1 Thành công

- `200` đọc / update
- `201` tạo
- `204` xóa / logout không body

### 1.2 Lỗi

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "File not found"
  }
}
```

| HTTP | `code` |
|---|---|
| 400 | `VALIDATION_ERROR` |
| 401 | `UNAUTHORIZED` |
| 403 | `FORBIDDEN` / `EMAIL_NOT_VERIFIED` / `REGISTER_DISABLED` |
| 404 | `NOT_FOUND` |
| 409 | `CONFLICT` (tên trùng, restore trùng tên) |
| 413 / 409 | `QUOTA_EXCEEDED` / `FILE_TOO_LARGE` |
| 409 | `UPLOAD_EXPIRED` / `INVALID_STATE` |
| 500 | `INTERNAL` |

`message` đủ để UI hiện; không leak SQL / stack.

### 1.3 Authz

Public (không Bearer): `register`, `login`, `refresh`, `verify-email`, `resend-verification`.

Còn lại cần access token. Chưa `email_verified_at`: chỉ `me`, `logout`, `refresh`. Mọi Files/Photos/Trash/Search → `403 EMAIL_NOT_VERIFIED`.

Resource chỉ của owner. Không thấy / không download file người khác → `404` (không `403` để khỏi dò ID).

Nhiều refresh token / user (nhiều máy). Logout: revoke **đúng** token client gửi.

### 1.4 Token

| | TTL (gợi ý) | Nơi sống |
|---|---|---|
| Access JWT | 15 phút | memory client |
| Refresh | 7 ngày | httpOnly cookie **hoặc** body; server lưu hash |

Phase 1 chấp nhận refresh trong JSON body cho dễ CLI sau này. Cookie httpOnly có thể thêm khi siết web.

Logout: revoke refresh token hiện tại.

## 2. Auth & user

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
POST /api/v1/auth/verify-email
POST /api/v1/auth/resend-verification
GET  /api/v1/users/me
PATCH /api/v1/users/me
POST /api/v1/users/me/password
```

**Register** `{ email, password, displayName, inviteCode }` → `201` `{ user, accessToken, refreshToken }`.

- Sai/thiếu invite → `403 FORBIDDEN` (không nói mã đúng hay sai chi tiết)
- `FILVAULT_INVITE_CODE` trống → `403 REGISTER_DISABLED`
- Email đã có → `409 CONFLICT`
- User tạo ra **chưa** verify (`emailVerified: false`)

**Login** `{ email, password }` → `200` cùng shape kể cả chưa verify (để client vào màn hình nhập mã). Sai mật khẩu → `401` (không nói “email không tồn tại”).

**Verify** `{ email, code }` → `204`. Sai/hết hạn: `400 VALIDATION_ERROR` (không leak).

**Resend** `{ email }` → luôn `204`. Nếu user tồn tại và chưa verify: tạo mã mới, gửi qua Mailer (SMTP hoặc console).

**Me** → `{ id, email, displayName, emailVerified, storageUsed, storageQuota, trashAutoDeleteEnabled, trashRetentionDays, createdAt }`. Không trả `passwordHash`.

**Patch me** `{ displayName?, trashAutoDeleteEnabled?, trashRetentionDays? }`.

**Đổi mật khẩu** `{ currentPassword, newPassword }` → revoke mọi refresh token, `200` `{ accessToken, refreshToken }`. Sai mật khẩu hiện tại → `401`.

Password: **argon2id**, tối thiểu 8 ký tự.

## 3. Folders

```text
POST   /api/v1/folders
GET    /api/v1/folders/{id}
PATCH  /api/v1/folders/{id}
DELETE /api/v1/folders/{id}
POST   /api/v1/folders/{id}/restore
```

**Create** `{ name, parentId? }` → `201`. `parentId` omit/null = root. Cha không tồn tại / không thuộc user → `404`. Tên trùng sibling → `409`.

**Patch** `{ name?, parentId? }` — rename và/hoặc move. Move không được tạo chu trình (folder thành con cháu của chính nó). Move vào folder đang trong trash → `INVALID_STATE`.

**Delete** soft. Folder còn con chưa xóa → `409 CONFLICT`.

Body folder:

```json
{
  "id": "...",
  "parentId": null,
  "name": "Documents",
  "createdAt": "...",
  "updatedAt": "..."
}
```

## 4. Files & upload

```text
POST /api/v1/files/upload-sessions
POST /api/v1/files/{id}/complete
GET  /api/v1/files/{id}
GET  /api/v1/files/{id}/download
PATCH /api/v1/files/{id}
DELETE /api/v1/files/{id}
POST /api/v1/files/{id}/restore
```

### 4.1 Tạo session

`POST /api/v1/files/upload-sessions`

```json
{
  "name": "photo.jpg",
  "size": 2849512,
  "contentType": "image/jpeg",
  "folderId": null
}
```

Server:

1. Authz folder (nếu có)
2. Kiểm tra allowlist (extension + contentType) và `size <= MaxFileSizeBytes` (100 MiB)
3. Kiểm tra quota (`storage_used + size`)
4. Tạo file `PENDING`, `object_key`, `upload_expires_at`
5. `ObjectStore.CreateUploadURL`

`201`:

```json
{
  "fileId": "...",
  "uploadUrl": "https://...",
  "expiresAt": "..."
}
```

Client `PUT uploadUrl` **thẳng** object storage (không qua Go). Header `Content-Type` khớp `contentType` đã ký.

Hết hạn presign / hủy: `DELETE /api/v1/files/{id}` khi `PENDING` — xóa row, cố gắng xóa object nếu đã lên; không đụng `storage_used`.

### 4.2 Complete

`POST /api/v1/files/{id}/complete`

Không tin body size từ client. Server `Head` object:

- Không có object / hết hạn session → `FAILED` + lỗi
- Size thật > quota còn lại → `QUOTA_EXCEEDED`, `FAILED`
- OK → `status=READY`, `size_bytes` = size thật, `storage_used += size thật`, xóa `upload_expires_at`

`200` metadata file.

### 4.3 Download

`GET /api/v1/files/{id}/download` → `{ "downloadUrl", "expiresAt" }`. Chỉ `READY` và chưa trash. Client GET/PUT URL đó tới object storage.

Presign download/upload: **5–15 phút** (hằng số, gợi ý 15 phút upload / 5 phút download).

### 4.4 Patch / delete file

Patch: `{ name?, folderId? }` — rename/move. Cùng rule tên trùng. Không đổi `object_key`.

Delete: soft, chỉ `READY`. `PENDING` delete = abort (§4.1).

## 5. Browser, trash, search, storage

```text
GET /api/v1/browser?folderId=
GET /api/v1/trash
DELETE /api/v1/trash/{type}/{id}
GET /api/v1/search?q=
GET /api/v1/storage
```

**Browser:** `folderId` omit = root. Trả folders + files `READY` chưa xóa + breadcrumb. Đây là API chính của UI “My Files”.

```json
{
  "folder": null,
  "breadcrumb": [],
  "folders": [],
  "files": []
}
```

`files[]` gồm `id, name, mimeType, sizeBytes, updatedAt` — không lộ `object_key` nếu không cần. Phase 1 có thể trả `objectKey` cho debug; production web không cần hiện.

**Trash:** list file/folder có `deleted_at`. Restore dùng `POST .../restore` ở trên.

**Permanent delete:** `DELETE /api/v1/trash/files/{id}` hoặc `.../folders/{id}`. Folder vĩnh viễn: chỉ khi không còn con trong trash **hoặc** xóa hết con trước (Phase 1: folder trống mới được xóa mềm → permanent cũng chỉ folder trống). File: xóa object + row + giảm quota.

**Search:** `q` bắt buộc. Trả `{ folders, files }` theo spec 03 §5. Không search trong trash.

**Storage:** `{ usedBytes, quotaBytes }`.

## 6. Photos

```text
GET    /api/v1/photos/timeline?before=&limit=
GET    /api/v1/photos/albums
POST   /api/v1/photos/albums
GET    /api/v1/photos/albums/{id}
PATCH  /api/v1/photos/albums/{id}
DELETE /api/v1/photos/albums/{id}
POST   /api/v1/photos/albums/{id}/items
DELETE /api/v1/photos/albums/{id}/items/{fileId}
```

**Timeline:** file `READY`, chưa trash, mime ảnh hoặc video. Nhóm theo ngày `created_at` (UTC date). Không trả URL object. `limit` mặc định 50.

```json
{
  "groups": [
    {
      "date": "2026-08-14",
      "items": [
        { "id": "...", "name": "a.jpg", "mimeType": "image/jpeg", "sizeBytes": 123, "createdAt": "..." }
      ]
    }
  ],
  "nextBefore": "..."
}
```

**Albums:** `{ name }` tạo; patch `{ name }`; delete album không xóa file.

`GET /photos/albums` trả danh sách album kèm `itemCount` (số item ảnh/video `READY`, chưa trash trong album):

```json
{
  "albums": [
    { "id": "...", "name": "Trip", "itemCount": 3, "createdAt": "...", "updatedAt": "..." }
  ]
}
```

`GET /photos/albums/{id}` trả cùng shape trên cộng `items` (danh sách item).

**Items:** `{ fileIds: ["..."] }`. File không phải ảnh/video / không phải của user / trash → `404` hoặc `VALIDATION_ERROR`. Trùng trong album → bỏ qua hoặc `409` (chốt: **idempotent**, trùng = no-op).

Nghiệp vụ đầy đủ: spec 07 §4.

## 7. Không có ở Phase 1

```text
/shares  /devices  /sync
multipart upload  resumable
public object URL
Cognito
```

## 8. ID trên URL

Mọi `{id}` là ULID. Sai format → `400 VALIDATION_ERROR`, không `500`.
