# 11 — Spec trải nghiệm Chat (Messenger-like)

Trạng thái: **Accepted / C1-C3 MVP implemented**. Đây là một **product surface riêng** sau Phase 2 web features. Không thay thế Files/Vault hoặc Photos.

Nguồn nền: spec 03 (schema), spec 04 (API), spec 05 (architecture), spec 09 (Phase 2 web), `DESIGN.md`.

## 0. Mục tiêu

Thêm một trải nghiệm chat giống Messenger:

- Chat là màn riêng, layout/UX riêng, không trộn vào Files hoặc Photos.
- Người dùng nhắn text và gửi ảnh/video/file trong hội thoại.
- Ảnh gửi trong chat **vẫn được lưu vào storage như file bình thường**, có metadata tên file, ngày gửi/upload, người gửi, và sau này hiện được trong Photos.
- Photos vẫn là trải nghiệm thư viện media riêng: timeline/album/cover/favorites. Photos không hiển thị bubble chat, thread, hoặc nội dung tin nhắn.
- Files/Vault vẫn là quản lý file/folder. File chat có thể xem trong Files ở khu vực hệ thống hoặc filter riêng, nhưng không làm hỏng model folder hiện tại.

## 1. Nguyên tắc phân tách surface

| Surface | Mục đích | Không làm |
|---|---|---|
| Vault / Files | quản lý file/folder, upload/download, search, trash | không hiển thị thread chat |
| Photos | timeline media + album, trải nghiệm xem ảnh/video | không hiển thị message bubble |
| Chat | hội thoại realtime/asynchronous, gửi text/media | không thay thế Photos hoặc Files |

Một file ảnh gửi qua Chat có **hai mặt**:

1. Là attachment của message trong Chat.
2. Là media item trong Photos (vì vẫn là file READY, mime ảnh/video, có ngày).

## 2. Phase triển khai đề xuất

| Slice | Nội dung | Ghi chú |
|---|---|---|
| C1 | Chat shell + conversation list + message list text-only | Done: `/chat`, conversation/message API, manual refresh |
| C2 | Gửi ảnh/file attachment bằng upload flow hiện có | Done: reuse upload session/object store |
| C3 | Attachment media xuất hiện trong Photos + metadata chat | Done ở mức MVP: file chat là `READY` media nên vào Photos timeline |
| C4 | Realtime nhẹ (SSE hoặc WebSocket) | ADR riêng trước khi chọn |
| C5 | Search trong chat + media gallery theo conversation | Done: `/messages/search?q=` + `/media` |
| C6 | Chat là bare surface riêng: route `meta.bare`, full-screen Messenger layout ngoài shell Vault | Done: `/chat` không bottom/side nav, mobile master–detail + Back |
| C7 | Polish motion + API đáp ứng UI: cursor pagination (`before`/`limit`, trả `hasMore`/`nextBefore`), preview tin cuối (`?includePreview=true`), optimistic send, TransitionGroup bubble, jump-to-latest, composer auto-grow | Done |

C4 (realtime) cần ADR riêng trước khi làm.

## 3. Database dự kiến

Chỉ tạo khi bắt đầu C1/C2.

```text
conversations (
  id CHAR(26) PK,
  owner_id CHAR(26) NOT NULL REFERENCES users(id),
  title TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  archived_at TIMESTAMPTZ NULL
)

messages (
  id CHAR(26) PK,
  conversation_id CHAR(26) NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  owner_id CHAR(26) NOT NULL REFERENCES users(id),
  body TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ NULL
)

message_attachments (
  id CHAR(26) PK,
  message_id CHAR(26) NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  file_id CHAR(26) NOT NULL REFERENCES files(id),
  original_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE(message_id, file_id)
)
```

### 3.1 File gửi trong chat

- Reuse bảng `files`: `name`, `mime_type`, `size_bytes`, `created_at`, `updated_at`, `status`.
- Thêm cột khi vào C2:

```text
ALTER TABLE files ADD COLUMN source TEXT NOT NULL DEFAULT 'vault';
ALTER TABLE files ADD COLUMN source_ref_id CHAR(26) NULL;
```

Giá trị:

- `source='vault'`: file upload từ Files như hiện tại.
- `source='chat'`: file upload từ Chat; `source_ref_id` trỏ message id hoặc conversation id (chốt khi implement).

Photos lấy media bằng mime ảnh/video và `status='READY'`, không loại trừ `source='chat'`.

## 4. API dự kiến

```text
GET  /api/v1/chat/conversations
POST /api/v1/chat/conversations              { title? }
GET  /api/v1/chat/conversations/{id}/messages?before=&limit=
POST /api/v1/chat/conversations/{id}/messages { body, fileIds? }
POST /api/v1/chat/attachments/upload-sessions { name, size, contentType, conversationId }
POST /api/v1/chat/attachments/{fileId}/complete
```

Quy tắc:

- Conversation luôn thuộc owner hiện tại. ID lạ/không thuộc user → `404`.
- `body` có thể rỗng nếu có attachment; không có body và không attachment → `400`.
- Attachment dùng allowlist file hiện tại (spec 07 §3).
- Complete attachment tạo/ghim file READY và liên kết vào message.
- File chat bị trash trong Files → message vẫn còn nhưng attachment hiển thị trạng thái “Unavailable”.

## 5. UI/UX Chat

Chat là **bare surface**: route `/chat` có `meta: { bare: true }` để `AppShell` không render shell Filvault (bottom nav/side nav/header/storage bar). Trải nghiệm độc lập, truy cập trực tiếp qua URL, layout full-screen kiểu Messenger. Chi tiết visual chuẩn hoá trong `DESIGN.md` §9a "Chat".

### 5.1 Desktop

- 3 cột kiểu Messenger:
  - trái: conversations list + search
  - giữa: thread messages
  - phải: details / shared media (ẩn được)
- Composer cố định dưới thread: input multiline, attach button, send button.
- Bubble:
  - outgoing align right, accent/ink background
  - attachment image preview trong bubble, tap mở lightbox
  - file non-image hiển thị card icon + name + size

### 5.2 Mobile

- Conversation list là màn đầu.
- Chọn conversation → thread full-screen.
- Header có back, title, menu.
- Composer bám dưới, tránh bàn phím, safe-area aware.

### 5.3 Motion/accessibility

- Thread append message: fade/slide nhẹ, respect `prefers-reduced-motion`.
- Composer send: optimistic bubble pending → sent/failed.
- Ảnh upload có progress rõ ràng.
- Bubble có text selectable, attachment có aria-label tên file.

## 6. Quan hệ với Photos

Kỳ vọng của sản phẩm:

- Ảnh gửi trong Chat vẫn xuất hiện trong Photos timeline theo ngày upload/gửi.
- Photos có thể có filter sau này: All / Camera uploads / Chat media, nhưng mặc định vẫn là timeline media.
- Album có thể thêm ảnh từ chat như ảnh thường.
- Chat không hiển thị album UI; Photos không hiển thị conversation UI.

## 7. Không có trong bản đầu

```text
multi-user realtime group chat
read receipts
typing indicator
message reactions
end-to-end encryption
voice/video call
AI chat memory ngoài metadata file
```

Những mục trên mở bằng ADR riêng.

## 8. ADR cần có trước khi code

1. Realtime transport: polling vs SSE vs WebSocket.
2. File source model: `files.source/source_ref_id` hay bảng riêng `chat_files`.
3. Retention: xóa conversation có xóa message không; có xóa file attachment không.
4. Nếu sau này multi-user: permission model và sharing boundary.

