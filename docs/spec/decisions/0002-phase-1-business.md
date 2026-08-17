# ADR 0002 — Nghiệp vụ Phase 1

- Trạng thái: **Accepted**
- Ngày: 2026-08-14
- Spec: [../07-phase-1-business.md](../07-phase-1-business.md)

## Bối cảnh

Cần siết nghiệp vụ Phase 1 trước khi code. Filvault về sau có thể phục vụ app khác (ví dụ quản lý trường mầm non) — **toàn bộ** khả năng lưu trữ, không chỉ upload ảnh — nhưng việc đó **sau khi sản phẩm hiện tại hoàn chỉnh**. Web Phase 1 cần cả quản lý tệp và quản lý ảnh (Terabox / Google Photos).

## Quyết định

1. Phase 1 = Drive cá nhân + tab Photos (cùng kho file). Không `tenant_id` / API key.
2. Photos = lọc ảnh/video + timeline theo ngày upload + album (tham chiếu). Không thumbnail server; không nhúng original vào grid.
3. Register: mã mời env. Email phải verify (SMTP hoặc console). Nhiều refresh token. Đổi displayName + password, không đổi email.
4. Allowlist MIME/extension + max 100 MiB / file. Có video.
5. Trash: setting theo user, default từ config; ticker in-process tự xóa nếu bật.
6. Go: Gin, pgx SQL thuần, slog, validate ở biên, error code chuẩn.
7. Tích hợp app khác: cùng full API (files/folders/photos/trash/…); chưa chốt auth (API key/OAuth); không library; làm sau khi sản phẩm xong.

## Hệ quả

- Thêm bảng `albums` / `album_items` và vài cột trên `users` ngay Phase 1.
- Mailer là cổng (console | smtp); SES để sau.
- Scope Phase 1 lớn hơn concept gốc (thêm Photos/album/verify/invite) nhưng vẫn không CLI, không worker AWS, không thumbnail.
