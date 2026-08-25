# ADR 0004 — Chat là surface riêng, media reuse storage

- Trạng thái: **Accepted**
- Ngày: 2026-08-25
- Spec: [../11-chat-experience.md](../11-chat-experience.md)

## Bối cảnh

Filvault hiện có hai trải nghiệm chính:

- Vault / Files: quản lý file/folder.
- Photos: timeline media + album.

Người dùng muốn thêm chat giống Messenger: nhắn text, gửi ảnh, và ảnh gửi qua chat vẫn được lưu/ghi nhớ tên + ngày để sau này xem được trong Photos. Đồng thời Chat, Photos và quản lý tệp phải là các trải nghiệm riêng, không lẫn UI/UX.

## Quyết định

1. Chat là product surface riêng, có route/layout riêng và không dùng list UI của Files.
2. Attachment trong Chat reuse object storage + upload completion flow hiện có; không tạo storage riêng cho chat.
3. Ảnh/video gửi qua Chat vẫn là `files` READY nên Photos timeline nhìn thấy như media bình thường.
4. Chat lưu quan hệ message ↔ file bằng `message_attachments`; message giữ context hội thoại, file giữ metadata storage.
5. Không chọn realtime transport ngay. C1–C3 có thể dùng refresh/polling đơn giản; WebSocket/SSE cần ADR riêng trước khi code.

## Hệ quả

- Cần thêm schema chat (`conversations`, `messages`, `message_attachments`) khi bắt đầu C1/C2.
- Có thể cần thêm metadata source cho `files` để phân biệt upload từ Vault và Chat.
- Photos không biết message bubble; Chat không biết album UI.
- Xóa/retention conversation và attachment là quyết định riêng trước khi code delete nâng cao.

