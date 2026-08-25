# Filvault — Spec

Thư mục này chứa **spec ý tưởng và quyết định thiết kế** đã được chốt để triển khai. Nó được tách từ [`../concept-product-scope_v0.1.md`](../concept-product-scope_v0.1.md) (concept — tài liệu khái niệm v0.1).

## Cách đọc

| Vai trò | Tài liệu | Khi nào thắng nếu lệch nhau |
|---|---|---|
| Phạm vi sản phẩm, Phase 1, roadmap | Concept v0.1 | Concept thắng cho **sản phẩm / phase** |
| Kiến trúc, cổng nhà cung cấp, quyết định kỹ thuật mới | `docs/spec/` | Spec thắng cho **kiến trúc** |

Spec không thay concept. Spec **làm rõ và siết** những chỗ concept còn mở, đặc biệt hướng:

> **AWS-first integration, provider-agnostic core**  
> (tích hợp AWS trước, lõi không lệ thuộc nhà cung cấp)

Nghĩa là: học và dùng AWS thật, nhưng domain (nghiệp vụ) không dính SDK hay dịch vụ cụ thể.

## Mục lục hiện tại

| File | Nội dung |
|---|---|
| [01-idea.md](01-idea.md) | Spec ý tưởng: sản phẩm, mục tiêu, ranh giới |
| [02-provider-strategy.md](02-provider-strategy.md) | Chiến lược nhà cung cấp: cổng, adapter, RDS vs DynamoDB |
| [03-database.md](03-database.md) | Schema Phase 1, status, quota, trash, search |
| [04-api.md](04-api.md) | HTTP Phase 1: auth, upload, browser, lỗi |
| [05-architecture.md](05-architecture.md) | Monorepo, module Go, flow upload, test |
| [06-phase-1-plan.md](06-phase-1-plan.md) | Thứ tự slice + Definition of Done |
| [07-phase-1-business.md](07-phase-1-business.md) | Nghiệp vụ đã chốt: invite, verify, Photos, allowlist, trash |
| [08-phase-1-status.md](08-phase-1-status.md) | Checklist bàn giao Phase 1; smoke test tay |
| [09-phase-2-features.md](09-phase-2-features.md) | Spec tính năng Phase 2 web (Accepted): search, cover, favorites, share, activity |
| [10-phase-2-status.md](10-phase-2-status.md) | Checklist bàn giao Phase 2; S1 done → tiếp S2 |
| [11-chat-experience.md](11-chat-experience.md) | Spec đề xuất Chat kiểu Messenger; tách riêng Vault/Photos |
| [decisions/0001-aws-first-provider-agnostic.md](decisions/0001-aws-first-provider-agnostic.md) | ADR: AWS-first, provider-agnostic |
| [decisions/0002-phase-1-business.md](decisions/0002-phase-1-business.md) | ADR: nghiệp vụ Phase 1 |
| [decisions/0003-phase-2-features-scope.md](decisions/0003-phase-2-features-scope.md) | ADR: phạm vi tính năng Phase 2 (**Accepted**) |
| [decisions/0004-chat-separate-surface.md](decisions/0004-chat-separate-surface.md) | ADR đề xuất: Chat là surface riêng, media reuse storage |
| [`../../DESIGN.md`](../../DESIGN.md) | Design system web (Cal.com-like + teal, mobile-first; §9a Phase 2) |

**Phase 1** đã có spec + implement (checklist: [08](08-phase-1-status.md)).  
**Phase 2 web:** spec [09](09-phase-2-features.md) Accepted; tiến độ [10](10-phase-2-status.md) — **S1 xong**, tiếp **S2**.

## Chưa viết (đúng lúc)

- ADR bảo mật share link (token TTL, rate limit số cụ thể) — **bắt buộc trước khi code S4**
- Chi tiết UI/endpoint share user nội bộ (S5) — viết khi vào slice
- ADR realtime chat (polling/SSE/WebSocket) — viết trước khi code Chat C4
- Database/API cho versioning, sync
- Terraform / RDS deploy spec
- JobQueue + worker spec (Phase 3)
- CLI / sync spec
- Tích hợp app khác (full API + API key/tenant) — sau khi sản phẩm Filvault hoàn chỉnh

## Quy ước

- **Thiết kế sẵn ≠ implement sẵn.** Interface (giao diện lập trình) và mapping được ghi trong spec; code adapter chỉ viết khi phase cần.
- Tên dịch vụ AWS giữ nguyên tiếng Anh. Thuật ngữ khó có giải thích ngắn trong ngoặc.
- Mọi thay đổi kiến trúc mới phải có ADR trong `decisions/`.
- Mỗi slice Phase 2: TDD + DoD trong [09 §0.1](09-phase-2-features.md) và tick [10](10-phase-2-status.md).
