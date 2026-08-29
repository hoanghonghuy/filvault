# ADR 0005 — Production Direct Chat

- Trạng thái: **Accepted**
- Ngày: 2026-08-28
- Spec: [../11-chat-experience.md](../11-chat-experience.md)

## Bối cảnh

Chat hiện tại là hội thoại một chủ sở hữu, phù hợp MVP nhưng chưa có participant,
sender, quyền truy cập attachment cho người nhận, hoặc cơ chế đồng bộ bền vững.
Mục tiêu tiếp theo là direct message production giữa hai người dùng đã xác minh,
không biến Chat thành một ACL mở cho toàn bộ Vault.

## Quyết định

1. **Mô hình direct message**
   - Một conversation direct có đúng hai active members.
   - Cặp user dùng `direct_key` canonical và chỉ có một conversation active.
   - User chỉ có thể mở DM với email chính xác của user đã xác minh.
   - Self-DM, recipient không tồn tại hoặc chưa xác minh dùng response riêng tư,
     không tiết lộ trạng thái tài khoản.
   - Dữ liệu legacy single-owner được giữ tương thích; owner được backfill thành
     member và message cũ được xem như do owner gửi.

2. **Authorization**
   - Membership là source of truth cho đọc conversation, message, search, media,
     stream và download attachment trong Chat.
   - Endpoint Chat trả `404` cho conversation/attachment không thuộc member để
     tránh dò ID.
   - Endpoint Files/Vault tiếp tục kiểm tra `files.owner_id`; không nới quyền
     generic download.

3. **Ownership và album chung**
   - File vật lý, quota và quyền Trash/Purge vẫn thuộc sender.
   - Mọi attachment gửi vào DM tự động được thêm vào album chung của conversation.
   - Album chung có hai collaborator là hai conversation members và được mở từ
     Chat/Photos theo permission của conversation.
   - Khi sender sửa/gỡ message hoặc Trash/Purge file nguồn, item trong album chung
     bị ẩn hoặc unavailable theo cùng lifecycle; không tạo bản sao và không tính
     quota lần hai.

4. **Message mutation**
   - Message có `sender_id`, `client_message_id`, `edited_at` và tombstone remove.
   - Chỉ sender được sửa hoặc gỡ message của chính mình trong 15 phút kể từ lúc
     tạo.
   - Sửa hiển thị `Edited`; gỡ hiển thị `Message removed`; nội dung cũ không được
     trả qua API/UI.
   - REST mutation dùng idempotency key/client ID. Retry cùng payload trả kết quả
     cũ; reuse key với payload khác trả conflict.

5. **Realtime**
   - Release production dùng REST cho write và SSE cho server-to-client event.
   - `chat_events` là event log durable, được ghi trong cùng transaction với
     message/attachment mutation.
   - PostgreSQL notification hoặc polling chỉ dùng để đánh thức stream; không là
     nguồn sự thật.
   - Client reconnect theo cursor, replay event đã bỏ lỡ, deduplicate theo event
     ID/message ID và refetch delta khi cần.
   - Stream yêu cầu authentication, kiểm membership và không truyền JWT dài hạn
     qua query string.

6. **Purge và upload cleanup**
   - Purge object storage dùng at-least-once job processing.
   - Job claim bằng row lock/lease; retry có backoff, giới hạn attempt, dead state
     và metrics/alert.
   - Object không tồn tại được xem là purge thành công.
   - Upload `PENDING` hết hạn và orphan object phải được worker dọn.
   - API không tự chạy cleanup theo từng replica trong production.

7. **Không thuộc release này**

   - Group chat.
   - Presence, typing indicator, read receipt.
   - End-to-end encryption.
   - Copy file sang quota riêng của recipient.
   - Message restore sau khi remove.

## Hệ quả

- Cần migration additive và rollout expand/contract; không sửa migration đã áp dụng.
- Cần endpoint Chat-scoped cho download attachment của recipient.
- Cần tách worker khỏi API và chuẩn bị deployment/backup/observability tương ứng.
- Album dùng quan hệ cộng tác hiện có của Photos/Vault nếu tương thích; nếu không,
  conversation membership là quyền tối thiểu riêng cho album chung.
