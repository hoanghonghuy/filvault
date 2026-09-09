# ADR 0003 — Phạm vi tính năng Phase 2 (web)

- Trạng thái: **Accepted** (2026-08-22)
- Ngày: 2026-08-22
- Spec: [../09-phase-2-features.md](../09-phase-2-features.md)
- Tiến độ: [../10-phase-2-status.md](../10-phase-2-status.md) — S1 Done; tiếp S2

## Bối cảnh

Phase 1 đã bàn giao đủ 6 màn (Auth, Files, Photos, Album, Trash, Settings) và UI/UX nền tảng (motion system, optimistic UI, focus trap — `DESIGN.md`). Cần chốt bộ tính năng kế tiếp cho web trước khi code, vì:

- Search hiện chỉ `ILIKE '%q%'` (spec 03 §5), chưa lọc/sắp xếp được.
- Backend đã có thumbnail presign nhưng album chưa có ảnh bìa.
- Share là tính năng lớn nhất và **thay đổi bề mặt bảo mật** sản phẩm: spec 04 §7 đang liệt kê `/shares` là "không có Phase 1", bảng `share_links` nằm danh sách "không tạo" (spec 03 §6). Không thể thêm im lặng.

## Quyết định

1. Chia 6 slice: S1 search filter/sort → S2 album cover → S3 favorites → S4 share public link → S6 activity log → S5 share user nội bộ. Mỗi slice tự chứa migration + API + UI, làm độc lập.
2. Search **giữ engine ILIKE**, chỉ thêm filter (`type`, `folderId`, khoảng ngày) + sort. `pg_trgm`/FTS hoãn đến khi dữ liệu thật chứng minh cần — tránh tune DB sớm.
3. Album cover = `cover_file_id` nullable; không ghim thì auto chọn item mới nhất. Không tạo bảng riêng.
4. Share làm **public link trước** (người nhận không cần tài khoản), user nội bộ sau (S5 reuse schema). Token lưu **hash** SHA-256 (cùng triết lý refresh token), public endpoint trả `404` chung một mã chống dò, rate limit theo IP.
5. Activity log chỉ ghi sự kiện thay đổi trạng thái (upload/trash/restore/share/password/settings), giữ 90 ngày, dọn bằng ticker sẵn có; ghi fail-open.
6. Không đụng kiến trúc Phase 1: không worker ngoài process, không engine search mới, không quyền write.

## Hệ quả

- Migration chỉ tạo khi slice tương ứng bắt đầu (giữ rule "không tạo bảng trống cho tương lai"); riêng `share_links` có cột `recipient_user_id NULL` giành chỗ từ đầu để S5 không phải migrate lại.
- S4 cần bổ sung ADR chi tiết bảo mật link (token TTL, rate limit số cụ thể) khi bắt đầu slice — **không code S4 trước ADR đó**.
- Web có thêm trang public `/s/:token` không cần login — surface duy nhất nằm ngoài shell; DESIGN.md bổ sung anatomy riêng (§9a).
- Spec 04 §7 và spec 03 §6 được thay thế bởi spec 09 cho các mục share/activity kể từ khi slice tương ứng vào implementation.
- Checklist bàn giao / DoD slice: [10-phase-2-status.md](../10-phase-2-status.md).

## Tiến độ (cập nhật kèm file 10)

| Slice | Ghi chú |
|---|---|
| S1 | Implemented 2026-08-22 — giữ ILIKE + filter/sort theo quyết định 2 |
| S2 | Implemented 2026-08-22 — cover_file_id + auto/pin theo quyết định 3 |
| S3 | Implemented 2026-08-22 — favorites |
| S4 | Implemented 2026-08-23 — theo [ADR 0004](0004-public-share-link-security.md) |
| S5 | Implemented 2026-08-23 — chống dò email + mail mời; UI /shared theo §5 của spec 09 |
| S6 | Implemented 2026-08-23 — activity log 90 ngày, fail-open |
