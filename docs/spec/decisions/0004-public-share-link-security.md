# ADR 0004 — Bảo mật public share link (S4)

- Trạng thái: **Accepted** (2026-08-23) — **implemented cùng ngày**
- Ngày: 2026-08-23
- Spec: [../09-phase-2-features.md](../09-phase-2-features.md) §4
- Tiến độ: [../10-phase-2-status.md](../10-phase-2-status.md) — S4 Done; tiếp S6

## Bối cảnh

S4 cho phép chủ file tạo **link công khai**: ai có link đều tải được file mà không cần tài khoản. Đây là lần đầu sản phẩm phục vụ dữ liệu user cho client **ẩn danh**, nên bề mặt tấn công mới gồm:

1. Dò link (token guessing / enumeration).
2. Link rò rỉ tiếp tục hoạt động khi đã hết hạn hoặc file đã bị xóa.
3. Abuse băng thông: quét link công khai để tải lại liên tục.
4. Nhầm lẫn với tính năng share nội bộ theo email (Phase 5, spec [10-sharing](../10-sharing.md)) **đã được implement** — module `share` hiện chiếm route `GET /shares` và bảng `shares`.

ADR 0003 quyết định hướng chung (hash token, 404 chung, rate limit) nhưng chưa chốt số cụ thể. ADR này chốt chi tiết trước khi code, theo rule "không code S4 trước ADR bảo mật".

## Quyết định

### 1. Token

- Token = **32 byte random** (`crypto/rand`), encode hex → chuỗi 64 ký tự URL-safe. Sinh tại `POST /files/{id}/share`, trả plaintext **đúng một lần**; server không lưu plaintext.
- DB lưu **SHA-256 hash** (cột `token_hash TEXT UNIQUE`) — lộ DB không lộ link. Cùng triết lý `refresh_tokens.token_hash` (spec 03 §3.2); tái dùng pattern `hashRefresh` ở `auth/token.go`.
- Lookup public: query theo `token_hash`. Không cần so sánh thời gian hằng (constant-time): giá trị tra cứu là hash ngẫu nhiên 256-bit, không dò được qua timing khi đã đi qua index unique.

### 2. TTL & thu hồi

- Hạn chọn khi tạo: **không hạn (NULL) | 1 giờ | 24 giờ | 7 ngày**. Không cho custom TTL tự do.
- Cột `expires_at TIMESTAMPTZ NULL`; check hết hạn **tại thời điểm request** (`expires_at > now()`), không cần cron vô hiệu hóa.
- Thu hồi = set `revoked_at = now()`. Link thu hồi/hết hạn/file trash/xóa vĩnh viễn → từ chối phục vụ ngay lập tức.

### 3. Điều kiện link hợp lệ (mọi request public)

```sql
revoked_at IS NULL
AND (expires_at IS NULL OR expires_at > now())
AND files.deleted_at IS NULL      -- không đang trash
AND files.status = 'READY'
```

Không khớp tất cả → `404 NOT_FOUND`.

### 4. Mã lỗi chống dò

- Mọi thất bại của endpoint public (token sai, đã thu hồi, hết hạn, file trash/đã xóa) trả **chung một mã**: `404 {"error":{"code":"NOT_FOUND","message":"Not found"}}`. Không phân biệt nguyên nhân trong response, không log token vào access log ở mức info.
- Endpoint owner (tạo/thu hồi/list) vẫn trả lỗi nghiệp vụ chi tiết như các module khác (`INVALID_STATE`, `NOT_FOUND`, …) vì đã qua auth.

### 5. Rate limit

- Áp dụng cho **2 endpoint public** `/api/v1/public/shares/*`, giới hạn theo IP (`c.ClientIP()`).
- Giới hạn mặc định: **30 req/phút/IP** với burst 60 (token bucket), hằng số config `FILVAULT_PUBLIC_SHARE_RATE_LIMIT_PER_MIN` (default 30). Vượt → `429 TOO_MANY_REQUESTS`.
- Implement bằng middleware in-process (map IP → bucket + mutex, dọn định kỳ) — đủ cho single-instance hiện tại; Redis chỉ cân nhắc khi chạy multi-instance (không làm ở Phase này).
- Endpoint owner **không** rate limit riêng (đã nằm sau auth + email-verified).

### 6. Presign tải xuống

- Public download trả presigned URL **TTL 5 phút** (dùng sẵn `config.DownloadPresignTTL`). File gốc tải trực tiếp từ object store, API không proxy bytes.
- Metadata public chỉ gồm `name`, `mimeType`, `sizeBytes`, `expiresAt`. **Không** trả id nội bộ, owner, folder, thumbnail URL.

### 7. Phân tách với share nội bộ (Phase 5)

Module `share` (bảng `shares`, route `/shares`, `/shared/*`) là share theo email và **đang hoạt động** — không đụng tới. Public link dùng:

- Bảng riêng `share_links` (kèm cột `recipient_user_id CHAR(26) NULL REFERENCES users(id)` giành chỗ cho S5, Phase này luôn NULL).
- Route owner đặt dưới **file resource** để tránh đụng `GET /shares` hiện có:
  - `POST   /api/v1/files/{id}/share`
  - `DELETE /api/v1/files/{id}/share`
  - `GET    /api/v1/share-links` (list link đang hoạt động của mình)
- Route public: `GET /api/v1/public/shares/{token}` và `GET /api/v1/public/shares/{token}/download`.

> Lệch so với spec 09 §4.2 (định `GET /shares` cho list public links): do route đó đã bị share nội bộ chiếm, đổi thành `/share-links` và cập nhật spec 09 theo ADR này.

### 8. Giới hạn nghiệp vụ

- Tối đa **20 link hoạt động** mỗi user (check khi tạo, `CONFLICT` khi vượt).
- Mỗi file tối đa **1 link hoạt động**: tạo lại khi đã có → thu hồi link cũ rồi tạo link mới (không báo lỗi).
- File phải `READY` + không trash khi tạo link, ngược lại `INVALID_STATE` / `NOT_FOUND`.

### 9. Web public page `/s/:token`

- Route Vue ngoài shell (không bottom nav, không auth guard): card giữa màn — tên, loại · size, nút Download, caption "Shared via Filvault".
- Lỗi mọi loại hiển thị chung: "This link is not available" (DESIGN.md §9a).

## Hệ quả

- Migration `00008_share_links.up/down.sql`: bảng như mục 7 + unique index `share_links_token_hash`, index `share_links_owner_created (owner_id, created_at DESC)`; schema_test thêm 1 lượt rollback.
- Config thêm `FILVAULT_PUBLIC_SHARE_RATE_LIMIT_PER_MIN` (+ default trong `.env.example`).
- Spec 09 §4.2 sửa route list thành `GET /share-links`; ghi chú nguồn quyết định = ADR này.
- Rate limiter là code mới trong `internal/platform/httpx` — kèm unit test riêng (bucket refill, 429).
- Không có: password-protect, share folder qua link, analytics truy cập, CDN/cache. Mở lại bằng ADR khi có nhu cầu thật.

## Ghi chú triển khai

| Hạng mục | Giá trị chốt |
|---|---|
| Token length | 32 byte random → hex 64 ký tự |
| Hash | SHA-256 hex, lưu `token_hash UNIQUE` |
| TTL choices | NULL / 1h / 24h / 7d |
| Presign TTL | 5 phút (download), không cache |
| Rate limit | 30 req/phút/IP, burst 60, 429 khi vượt |
| Lỗi public | `404 NOT_FOUND` chung một mã |
| Cap | 20 link/user; 1 link/file (tạo lại = revoke cũ) |
