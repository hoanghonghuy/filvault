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
| [08-phase-1-status.md](08-phase-1-status.md) | Checklist bàn giao Phase 1; việc tiếp theo: smoke test tay |
| [09-phase-2-features.md](09-phase-2-features.md) | Spec tính năng Phase 2 web: search filter, album cover, favorites, share link, activity |
| [decisions/0001-aws-first-provider-agnostic.md](decisions/0001-aws-first-provider-agnostic.md) | ADR: AWS-first, provider-agnostic |
| [decisions/0002-phase-1-business.md](decisions/0002-phase-1-business.md) | ADR: nghiệp vụ Phase 1 |
| [decisions/0003-phase-2-features-scope.md](decisions/0003-phase-2-features-scope.md) | ADR: phạm vi tính năng Phase 2 (Proposed) |
| [`../../DESIGN.md`](../../DESIGN.md) | Design system web Phase 1 (Cal.com-like + teal, mobile-first) |

Phase 1 **đã có spec và đã implement** (tiến độ bàn giao: [08-phase-1-status.md](08-phase-1-status.md)). Tính năng Phase 2 web đã có spec đề xuất: [09-phase-2-features.md](09-phase-2-features.md) + [ADR 0003](decisions/0003-phase-2-features-scope.md).

## Chưa viết (đúng lúc)

- Chi tiết bảo mật share link (token TTL, rate limit số cụ thể) — bổ sung ADR khi bắt đầu slice S4
- Chi tiết UI/endpoint share user nội bộ (S5) — viết khi vào slice
- Database/API cho versioning, sync
- Terraform / RDS deploy spec
- JobQueue + worker spec (Phase 3)
- CLI / sync spec
- Tích hợp app khác (full API + API key/tenant) — sau khi sản phẩm Filvault hoàn chỉnh

## Quy ước

- **Thiết kế sẵn ≠ implement sẵn.** Interface (giao diện lập trình) và mapping được ghi trong spec; code adapter chỉ viết khi phase cần.
- Tên dịch vụ AWS giữ nguyên tiếng Anh. Thuật ngữ khó có giải thích ngắn trong ngoặc.
- Mọi thay đổi kiến trúc mới phải có ADR trong `decisions/`.
