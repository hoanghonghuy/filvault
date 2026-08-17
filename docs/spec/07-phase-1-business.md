# 07 — Nghiệp vụ Phase 1 (đã chốt)

Trạng thái: **chốt**. Schema/API phải khớp file này. ADR: [decisions/0002-phase-1-business.md](decisions/0002-phase-1-business.md)

Tích hợp app khác (ví dụ quản lý trường mầm non) **không** thuộc Phase 1. Làm sau khi sản phẩm Filvault hoàn chỉnh. App khác dùng **cả nền tảng** (file, folder, Photos, trash, quota, …), không phải chỉ API ảnh. Phase 1 chỉ **dành chỗ trong spec**, không thêm cột `tenant_id` / `app_id`, không API key.

## 1. Hình dạng sản phẩm

Drive cá nhân: 1 tài khoản = files của đúng user đó.

Web Phase 1 có **hai mặt** trên **cùng một kho**:

```text
My Files          cây folder, mọi file trong allowlist
Photos            ảnh + video của user (lọc theo mime), kiểu Terabox / Google Photos
```

Upload một file ảnh/video → vừa hiện ở folder (My Files) vừa hiện ở Photos. Không copy object, không nhân metadata.

## 2. Tài khoản

| Mục | Chốt |
|---|---|
| Register | Bắt buộc `inviteCode` khớp `FILVAULT_INVITE_CODE` (env). So sánh constant-time. Env trống → register tắt (`REGISTER_DISABLED`) |
| User đầu tiên | Không ngoại lệ: vẫn cần mã trong env |
| Email | Phải verify mới dùng Files/Photos. Register xong **chưa** dùng kho |
| Verify | Mã 6 số, TTL 15 phút, lưu hash, không plaintext |
| Mailer | Có SMTP trong env → gửi mail. Không SMTP → in mã ra **console** (log). Không gửi đồng thời hai nơi |
| SMTP sau này | Cổng `Mailer`; adapter `ses` (AWS) chỉ thiết kế, chưa code |
| Login nhiều máy | Nhiều refresh token song song |
| Profile | Đổi `displayName` + đổi mật khẩu. **Không** đổi email Phase 1 |
| Đổi mật khẩu | Revoke **mọi** refresh token, trả cặp token mới |
| Password | argon2id, tối thiểu 8 ký tự |

Chưa verify: gọi được `me`, `verify-email`, `resend-verification`, `logout`, `refresh`. Mọi API folder/file/photos/trash/search → `403 EMAIL_NOT_VERIFIED`.

Resend / verify theo email: không tiết lộ email có tồn tại hay không (verify sai mã → `401`/`400` chung; resend luôn `204`).

## 3. File được nhận

Một file tối đa **100 MiB** (`104857600` bytes) — hằng số `MaxFileSizeBytes`. Phase 1 **không** multipart.

Phải khớp **cả** extension lẫn `contentType` theo bảng dưới. Lệch nhau hoặc ngoài list → `VALIDATION_ERROR` lúc tạo session.

| Nhóm | Extension | MIME |
|---|---|---|
| Ảnh | `jpg` `jpeg` | `image/jpeg` |
| | `png` | `image/png` |
| | `gif` | `image/gif` |
| | `webp` | `image/webp` |
| | `heic` `heif` | `image/heic` / `image/heif` |
| Video | `mp4` | `video/mp4` |
| | `mov` | `video/quicktime` |
| | `webm` | `video/webm` |
| Tài liệu | `pdf` | `application/pdf` |
| | `doc` | `application/msword` |
| | `docx` | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| | `xls` | `application/vnd.ms-excel` |
| | `xlsx` | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` |
| | `ppt` | `application/vnd.ms-powerpoint` |
| | `pptx` | `application/vnd.openxmlformats-officedocument.presentationml.presentation` |
| Text | `txt` | `text/plain` |
| | `md` | `text/markdown` |
| | `csv` | `text/csv` |
| Nén | `zip` | `application/zip` |

Không nhận: `svg`, `exe`, `sh`, `bat`, `js`, … (mọi thứ không nằm bảng). File 0 byte **được** nếu type hợp lệ.

**Photos** chỉ lấy mime ảnh + video trong bảng. PDF/office/text/zip chỉ ở My Files.

## 4. Photos: timeline + album

Không thumbnail server Phase 1. Grid **cấm** `<img src>` trỏ original (file tới 100MB). Placeholder theo mime; click mới lấy presign để xem/tải.

Timeline nhóm theo **ngày upload** (`created_at` UTC→local hoặc UTC, chốt UTC date trên API). Ngày chụp EXIF = khi có worker (Phase 3).

Album:

- Chỉ metadata: tên + danh sách `file_id` (ảnh/video READY, chưa trash, đúng owner)
- Thêm vào album **không** move file khỏi folder
- Xóa album **không** xóa file
- File vào trash: ẩn khỏi timeline/album (row `album_items` có thể giữ)
- Permanent delete file: xóa `album_items` (FK cascade)
- Tên album không trùng trong cùng user (chưa xóa)

Upload từ UI Photos = cùng `upload-sessions`, `folderId` null (root) trừ khi UI chọn folder.

## 5. Thùng rác

User có setting:

```text
trashAutoDeleteEnabled    mặc định = config hệ thống (gợi ý false)
trashRetentionDays        mặc định = config hệ thống (gợi ý 30), tối thiểu 1
```

Bật thì file/folder đã soft-delete quá số ngày → xóa vĩnh viễn (đúng rule permanent: file giảm quota; folder chỉ khi trống).

Chạy bằng **ticker trong process API** (không SQS/Lambda). Interval gợi ý 1 giờ.

## 6. Dành chỗ tích hợp app khác (không code)

Sau khi Filvault xong, hướng mặc định **dự kiến** (chưa chốt): app khác gọi **cùng HTTP API** của sản phẩm — không phải một “Photos API” tách.

Ví dụ trường mầm non có thể upload ảnh **và** lưu PDF, tổ chức folder, trash, quota, album… tùy app cần. Upload ảnh chỉ là một use case, không thu hẹp phạm vi tích hợp.

Không nhúng thư viện Go, không copy runtime.

Phase 1 không thêm:

```text
tenant_id, app_id, api_keys
```

Khi làm: auth adapter mới (API key / OAuth), resource vẫn `owner` — có thể là user hoặc service principal. Schema hiện tại cố ý chỉ `owner_id` = user. API ổn định ở Phase 1 chính là chỗ app khác sẽ cắm vào sau.

## 7. Stack Go (kỹ thuật, chốt cùng phase)

```text
HTTP:        Gin
Postgres:    pgx v5 (pgxpool), SQL thuần — không ORM, không database/sql
Log:         slog (structured)
Validate:    biên handler (input HTTP) → service nhận kiểu sạch
Error:       lỗi domain có code; handler map sang JSON spec 04 — một chỗ, không fmt.Errorf rải handler
```

SQL chỉ trong `platform/postgres`.
