# Filvault — Spec

> **Current work/handoff source of truth:** [`../CURRENT_PRODUCT_STATE.md`](../CURRENT_PRODUCT_STATE.md).  
> Các tài liệu `08-phase-1-status.md` và `10-phase-2-status.md` là **historical phase records**. Những mục “việc tiếp theo” bên trong phản ánh thời điểm của phase đó và **không** được dùng để override trạng thái sản phẩm/backlog hiện tại.

Thư mục này chứa **spec ý tưởng và quyết định thiết kế** đã được chốt để triển khai. Nó được tách từ [`../concept-product-scope_v0.1.md`](../concept-product-scope_v0.1.md) (concept — tài liệu khái niệm v0.1).

## Cách đọc

| Vai trò | Tài liệu | Khi nào thắng nếu lệch nhau |
|---|---|---|
| Trạng thái hiện tại / việc cần làm tiếp | [`../CURRENT_PRODUCT_STATE.md`](../CURRENT_PRODUCT_STATE.md) + GitHub issues/PRs | Current state thắng cho **handoff / backlog / release posture** |
| Phạm vi sản phẩm lịch sử, Phase 1, roadmap gốc | Concept v0.1 | Concept là record cho **ý định / phase lịch sử** |
| Kiến trúc, cổng nhà cung cấp, quyết định kỹ thuật mới | `docs/spec/` + Accepted ADR | Spec/ADR thắng cho **kiến trúc** |

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
| [08-phase-1-status.md](08-phase-1-status.md) | **Historical:** checklist bàn giao Phase 1; “next work” trong file không còn authoritative |
| [09-cli.md](09-cli.md) | CLI `filvault` (Phase 4): login, whoami, ls, upload, download |
| [09-phase-2-features.md](09-phase-2-features.md) | Spec tính năng Phase 2 web (Accepted): search, cover, favorites, share, activity |
| [10-phase-2-status.md](10-phase-2-status.md) | **Historical:** checklist bàn giao Phase 2; “next work” trong file không còn authoritative |
| [10-sharing.md](10-sharing.md) | Sharing (Phase 5): chia sẻ file/folder cho user khác; đã đồng bộ S5 (anti-probing, mail mời, activity, UI `/shared`) |
| [11-versioning.md](11-versioning.md) | Versioning (Phase 6): lịch sử phiên bản file khi replace |
| [11-chat-experience.md](11-chat-experience.md) | Spec đề xuất Chat kiểu Messenger; tách riêng Vault/Photos |
| [decisions/0001-aws-first-provider-agnostic.md](decisions/0001-aws-first-provider-agnostic.md) | ADR: AWS-first, provider-agnostic |
| [decisions/0002-phase-1-business.md](decisions/0002-phase-1-business.md) | ADR: nghiệp vụ Phase 1 |
| [decisions/0003-phase-2-features-scope.md](decisions/0003-phase-2-features-scope.md) | ADR: phạm vi tính năng Phase 2 (**Accepted**) |
| [decisions/0004-public-share-link-security.md](decisions/0004-public-share-link-security.md) | ADR: bảo mật public share link S4 (**Accepted**) |
| [decisions/0004-chat-separate-surface.md](decisions/0004-chat-separate-surface.md) | ADR đề xuất: Chat là surface riêng, media reuse storage |
| [`../../DESIGN.md`](../../DESIGN.md) | Design system web; current design-contract reconciliation is tracked in #8 |

**Phase 1** và **Phase 2 web** đều có historical implementation/status records. Để biết trạng thái **hiện tại** và việc cần làm tiếp, dùng [`../CURRENT_PRODUCT_STATE.md`](../CURRENT_PRODUCT_STATE.md) và GitHub issues/PRs thay vì các “next work” checklist cũ.

## Chưa viết (đúng lúc)

- Sync spec (CLI slice đầu đã có: [09-cli.md](09-cli.md); API/database sync chưa có)
- ADR realtime chat (polling/SSE/WebSocket) — viết trước khi đổi kiến trúc realtime materially
- Database/API cho sync
- Terraform / RDS deploy spec
- JobQueue + worker spec nếu product direction yêu cầu thêm contract ngoài implementation hiện tại
- Tích hợp app khác (full API + API key/tenant) — sau khi sản phẩm Filvault hoàn chỉnh

> Versioning (Phase 6) **đã spec + implement**: [11-versioning.md](11-versioning.md).

## Quy ước

- **Thiết kế sẵn ≠ implement sẵn.** Interface (giao diện lập trình) và mapping được ghi trong spec; code adapter chỉ viết khi phase cần.
- Tên dịch vụ AWS giữ nguyên tiếng Anh. Thuật ngữ khó có giải thích ngắn trong ngoặc.
- Mọi thay đổi kiến trúc mới phải có ADR trong `decisions/`.
- Historical phase checklists không được dùng làm current backlog khi chúng mâu thuẫn với current-state handoff hoặc GitHub source of truth.
