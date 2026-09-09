# Filvault — Trạng thái sản phẩm & bàn giao (hiện tại)

Cập nhật: **2026-09-10** (`develop` @ `21e2eeb` — post [#55](https://github.com/hoanghonghuy/filvault/pull/55) / [#7](https://github.com/hoanghonghuy/filvault/issues/7) E2E; trước đó [#56](https://github.com/hoanghonghuy/filvault/pull/56) / [#17](https://github.com/hoanghonghuy/filvault/issues/17), [#48](https://github.com/hoanghonghuy/filvault/pull/48) / [#6](https://github.com/hoanghonghuy/filvault/issues/6)).  
**Nguồn sự thật duy nhất** cho trạng thái sản phẩm, mục tiêu release và việc nên làm tiếp.

> Các checklist Phase 1/2 ([08](08-phase-1-status.md), [10](10-phase-2-status.md)) chỉ còn giá trị **lịch sử** — không dùng phần “việc tiếp theo” ở đó.

---

## Đọc trước khi làm việc

| Câu hỏi | Trả lời |
|---|---|
| Nhánh điều khiển? | **`develop`** — mọi PR feature merge vào đây (protected; xem [Hợp đồng merge](#hợp-đồng-merge-pr--developmain)). |
| Mục tiêu sản phẩm? | **Filvault production-ready** (không còn MVP/local-only). Triển khai, bảo mật, UX polish và QA release là công việc hiện tại — không phải “sau khi xong Phase 2”. |
| Spec kiến trúc / API? | [`docs/spec/README.md`](README.md), [`04-api.md`](04-api.md), [`05-architecture.md`](05-architecture.md) |
| Design system (SoT)? | [`DESIGN.md`](../../DESIGN.md) — [#8](https://github.com/hoanghonghuy/filvault/issues/8) IN_PROGRESS → PR [#54](https://github.com/hoanghonghuy/filvault/pull/54) open (MERGEABLE vs `develop`; shell-language recon với AppShell three-form-factor contract #48; chờ Designer/QA review) |
| Bảo mật / secrets? | [#3](https://github.com/hoanghonghuy/filvault/issues/3) (đã merge [#13](https://github.com/hoanghonghuy/filvault/pull/13)); audit: [`docs/security/`](../security/) |
| Việc tiếp theo nên làm gì? | **Mục [Backlog đang mở](#backlog-đang-mở)** — chỉ link issue/PR; trạng thái mutable lấy từ GitHub, không freeze checklist chi tiết trong doc. |

---

## Chạy local

```bash
cp .env.example .env    # chỉnh giá trị local trong file .env (không commit)
make up                 # Docker: postgres + minio + api + web (+ seed dev nếu bật)
make test               # API tests (cần Postgres + MinIO)
make dev-api            # API trên host :8080
make dev-web            # Vue :5173 (Node ^22)
```

| Dịch vụ | URL |
|---|---|
| Web | `http://localhost:5173` |
| API | `http://localhost:8080` |
| MinIO console | `http://localhost:9002` |

### Tài khoản dev (chỉ local)

`make up` / `make dev-api` tự seed user dev khi `FILVAULT_SEED_DEV_USER=true` (chỉ trong compose local). **Không bật seed dev trên production.**

| Trường | Biến env (xem `.env.example`) |
|---|---|
| Email | `FILVAULT_SEED_EMAIL` |
| Mật khẩu | `FILVAULT_SEED_PASSWORD` |
| Mã mời (register tay) | `FILVAULT_INVITE_CODE` |

---

## Ma trận năng lực

**Chú thích mức độ:**

- **Functional** — luồng chính hoạt động trên stack local; API + web đã có.
- **Polish / hardening** — còn issue design/UX, a11y, edge case hoặc QA chưa đóng.
- **Release-ready** — đáp ứng contract production (config, health, deploy, smoke) — *chưa* đạt cho toàn sản phẩm.

| Nhóm | Functional | Polish / hardening | Release-ready |
|---|---|---|---|
| **Auth & account** — invite, register, verify email, login/refresh, profile, đổi mật khẩu | Có (Phase 1) | [#28](https://github.com/hoanghonghuy/filvault/issues/28), [#27](https://github.com/hoanghonghuy/filvault/issues/27) | Chưa — phụ thuộc epic [#4](https://github.com/hoanghonghuy/filvault/issues/4) |
| **Files & folders** — browser, upload presign, rename/move, search/filter, trash | Có | [#21](https://github.com/hoanghonghuy/filvault/issues/21) IN_PROGRESS → PR [#58](https://github.com/hoanghonghuy/filvault/pull/58) (coherent search & filters; MERGEABLE vs `develop`; chờ human review); upload queue đã merge [#12](https://github.com/hoanghonghuy/filvault/pull/12) | Chưa |
| **Upload queue** — hàng đợi per-file, retry | Có ([#12](https://github.com/hoanghonghuy/filvault/pull/12)) | — | Chưa |
| **Photos & albums** — timeline, album, cover, favorites | Có (Phase 2 S2–S3) | [#23](https://github.com/hoanghonghuy/filvault/issues/23), [#29](https://github.com/hoanghonghuy/filvault/issues/29) | Chưa |
| **Sharing** — public link, share user nội bộ, Shared with me | Có (Phase 2 S4–S5) | [#25](https://github.com/hoanghonghuy/filvault/issues/25), [#24](https://github.com/hoanghonghuy/filvault/issues/24), [#32](https://github.com/hoanghonghuy/filvault/issues/32) | Chưa |
| **Chat & calls** — direct message, attachment, SSE, video call (LiveKit) | Có (C1–C7 + ADR 0005) | [#33](https://github.com/hoanghonghuy/filvault/issues/33), [#30](https://github.com/hoanghonghuy/filvault/issues/30) | Chưa |
| **Vault** — khu vực file cá nhân (`/vault`) | Có (routing [#37](https://github.com/hoanghonghuy/filvault/pull/37); UX media/a11y [#44](https://github.com/hoanghonghuy/filvault/pull/44) / [#18](https://github.com/hoanghonghuy/filvault/issues/18)) | Cross-cutting: [#29](https://github.com/hoanghonghuy/filvault/issues/29) (MediaLightbox) | Chưa |
| **Themes & settings** — dark/system, color themes, Theme Center | Có (cơ bản; free-only UI [#56](https://github.com/hoanghonghuy/filvault/pull/56) / [#17](https://github.com/hoanghonghuy/filvault/issues/17)) | [#15](https://github.com/hoanghonghuy/filvault/issues/15) IN_PROGRESS → PR [#59](https://github.com/hoanghonghuy/filvault/pull/59) (unify appearance state; MERGEABLE vs `develop`; chờ human review), [#14](https://github.com/hoanghonghuy/filvault/issues/14), [#34](https://github.com/hoanghonghuy/filvault/issues/34), [#36](https://github.com/hoanghonghuy/filvault/issues/36) IN_PROGRESS → PR [#57](https://github.com/hoanghonghuy/filvault/pull/57) (`auth.user` canonicality đã xử lý trên PR — PATCH response only, unsaved edits giữ trong SettingsView-local refs; chờ human review) | Chưa |
| **Trash & activity** | Có | [#26](https://github.com/hoanghonghuy/filvault/issues/26) | Chưa |
| **Shell & navigation** — responsive, Overview shortcuts | Có (responsive AppShell [#48](https://github.com/hoanghonghuy/filvault/pull/48) / [#6](https://github.com/hoanghonghuy/filvault/issues/6); no misleading PRO UI [#56](https://github.com/hoanghonghuy/filvault/pull/56) / [#17](https://github.com/hoanghonghuy/filvault/issues/17)) | [#22](https://github.com/hoanghonghuy/filvault/issues/22), [#16](https://github.com/hoanghonghuy/filvault/issues/16) | Chưa |
| **CLI** (`bin/filvault`) | Có (login, ls, upload, download — spec [09-cli](09-cli.md)) | Chưa ưu tiên release | Chưa |
| **Secrets & config hardening** | Có ([#13](https://github.com/hoanghonghuy/filvault/pull/13)) | — | Một phần — xem [#3](https://github.com/hoanghonghuy/filvault/issues/3) |
| **Health / readiness / migration gate** | Có ([#45](https://github.com/hoanghonghuy/filvault/pull/45) / [#41](https://github.com/hoanghonghuy/filvault/issues/41); runtime [#46](https://github.com/hoanghonghuy/filvault/pull/46) / [#42](https://github.com/hoanghonghuy/filvault/issues/42)) | — | Contract trên `develop` |
| **Production deploy & release smoke** | Baseline runtime [#46](https://github.com/hoanghonghuy/filvault/pull/46) / [#42](https://github.com/hoanghonghuy/filvault/issues/42) | — | Epic [#4](https://github.com/hoanghonghuy/filvault/issues/4); [#43](https://github.com/hoanghonghuy/filvault/issues/43) **PARKED** — chờ external Docker/production-like runtime evidence (PR [#49](https://github.com/hoanghonghuy/filvault/pull/49) draft; MERGEABLE vs `develop` nhưng evidence gate chưa đạt) |
| **Browser E2E critical path** — Playwright mobile/tablet/desktop | Có ([#55](https://github.com/hoanghonghuy/filvault/pull/55) / [#7](https://github.com/hoanghonghuy/filvault/issues/7)) | — | Required check trên ruleset; chưa đồng nghĩa production-ready |
| **Repository release integrity** — protected `develop`/`main`, required PR CI + review-thread resolution | Có — ruleset [Protect integration branches](https://github.com/hoanghonghuy/filvault/rules/22665159) (`22665159`; 5 required checks incl. E2E) | [#50](https://github.com/hoanghonghuy/filvault/issues/50) IN_PROGRESS (handoff doc + human close) | Enforced trên `develop`/`main`; issue mở đến khi human đóng |

Phase 1 API slices 0–9 và Phase 2 web slices S1–S6 **đã hoàn thành chức năng** (chi tiết lịch sử: [08](08-phase-1-status.md), [10](10-phase-2-status.md)). Điều đó **không** có nghĩa sản phẩm đã production-ready.

---

## Cổng chất lượng & release blockers

| Cổng | Trạng thái | Tham chiếu |
|---|---|---|
| CI lint / typecheck / test + secret scan + E2E (required on merge) | **Enforced** | Ruleset `22665159`; 5 required contexts (xem [Hợp đồng merge](#hợp-đồng-merge-pr--developmain)) |
| Secret scanning (PR) | Enforced + đã merge hardening | [#3](https://github.com/hoanghonghuy/filvault/issues/3), [#13](https://github.com/hoanghonghuy/filvault/pull/13) |
| Health `/healthz`, readiness `/readyz`, migration gate | Đã merge | [#45](https://github.com/hoanghonghuy/filvault/pull/45), [`docs/ops/health-readiness-migration.md`](../ops/health-readiness-migration.md) |
| Production runtime (TLS edge, reverse proxy, env contract) | **Đã merge** | [#46](https://github.com/hoanghonghuy/filvault/pull/46) / [#42](https://github.com/hoanghonghuy/filvault/issues/42) (`17bad79`) |
| Release smoke + persistence restart + rollback runbook | **PARKED** | [#43](https://github.com/hoanghonghuy/filvault/issues/43) — external Docker/production-like runtime evidence; PR [#49](https://github.com/hoanghonghuy/filvault/pull/49) (draft; MERGEABLE vs `develop`; không undraft / không hạ evidence gate) |
| Protected integration branches + required PR quality gates + review-thread resolution | **Enforced** | Ruleset [22665159](https://github.com/hoanghonghuy/filvault/rules/22665159); [#50](https://github.com/hoanghonghuy/filvault/issues/50) IN_PROGRESS (handoff doc #39/#47 + human close) |
| Browser E2E critical path (`Web — E2E critical path`) | **Đã merge** + **required** on ruleset | [#55](https://github.com/hoanghonghuy/filvault/pull/55) / [#7](https://github.com/hoanghonghuy/filvault/issues/7) |
| Design system đồng bộ surfaces hiện tại | Chưa | [#8](https://github.com/hoanghonghuy/filvault/issues/8) IN_PROGRESS → PR [#54](https://github.com/hoanghonghuy/filvault/pull/54) open (MERGEABLE vs `develop`; shell-language recon với AppShell three-form-factor contract #48; chờ Designer/QA review) |

**Epic điều phối production:** [#4](https://github.com/hoanghonghuy/filvault/issues/4) — #41, #42 và #7/#55 (E2E) đã merge; branch protection + 5 required checks enforced (ruleset `22665159`, [#50](https://github.com/hoanghonghuy/filvault/issues/50) IN_PROGRESS until handoff lands + human close); [#43](https://github.com/hoanghonghuy/filvault/issues/43) **PARKED** on external Docker evidence (PR [#49](https://github.com/hoanghonghuy/filvault/pull/49)).

---

## Hợp đồng merge (PR → `develop`/`main`)

`develop` và `main` được bảo vệ bởi ruleset **[Protect integration branches](https://github.com/hoanghonghuy/filvault/rules/22665159)** (id `22665159`).

| Quy tắc | Chi tiết |
|---|---|
| Luồng merge | Task branch → PR → required CI xanh → unresolved review threads resolved → merge (không push trực tiếp implementation lên `develop`/`main`) |
| Required status checks (strict / up-to-date) | `Security — secret scan`, `API — lint, typecheck, test`, `Web — lint, typecheck, test, build`, `CLI — lint, typecheck, test`, **`Web — E2E critical path`** (thêm sau merge [#55](https://github.com/hoanghonghuy/filvault/pull/55)) |
| Review thread resolution | **Enforced** — unresolved PR review conversations block merge (`required_review_thread_resolution`) |
| Không required | PR Agent (skipped), GitGuardian (optional dashboard), approving-review count (PM comment gates thay APPROVED reviews) |
| Chặn | Force-push (`non_fast_forward`), xóa nhánh |
| Bypass | Repository Admin — emergency only |

[#50](https://github.com/hoanghonghuy/filvault/issues/50) IN_PROGRESS until handoff doc lands (PR [#47](https://github.com/hoanghonghuy/filvault/pull/47) / [#39](https://github.com/hoanghonghuy/filvault/issues/39)) and human closes the issue; ruleset contract above is already enforced.

---

## Backlog đang mở

Tra GitHub để lấy trạng thái mới nhất — doc chỉ giữ **ref** issue/PR, không duplicate checklist mutable. Snapshot: `develop` @ `21e2eeb` (2026-09-10).

### PR đang mở (ưu tiên review/merge)

| PR | Issue | Mô tả ngắn |
|---|---|---|
| [#58](https://github.com/hoanghonghuy/filvault/pull/58) | [#21](https://github.com/hoanghonghuy/filvault/issues/21) | Coherent Files search & filters (open; MERGEABLE vs `develop`; chờ human review) |
| [#59](https://github.com/hoanghonghuy/filvault/pull/59) | [#15](https://github.com/hoanghonghuy/filvault/issues/15) | Unify appearance state (light/dark/system + accent themes) (open; MERGEABLE vs `develop`; chờ human review) |
| [#54](https://github.com/hoanghonghuy/filvault/pull/54) | [#8](https://github.com/hoanghonghuy/filvault/issues/8) | DESIGN.md reconcile (open; MERGEABLE vs `develop`; chờ Designer/QA review) |
| [#57](https://github.com/hoanghonghuy/filvault/pull/57) | [#36](https://github.com/hoanghonghuy/filvault/issues/36) | Settings section-scoped save (open; MERGEABLE vs `develop`; `auth.user` canonicality đã xử lý trên PR; chờ human review) |
| [#49](https://github.com/hoanghonghuy/filvault/pull/49) | [#43](https://github.com/hoanghonghuy/filvault/issues/43) | Release smoke — **PARKED** on external Docker/production-like runtime evidence (draft; MERGEABLE vs `develop`; không merge-ready — không undraft / không hạ evidence gate) |
| [#47](https://github.com/hoanghonghuy/filvault/pull/47) | [#39](https://github.com/hoanghonghuy/filvault/issues/39) | Handoff doc này (draft) |

### Đã merge gần đây (tham chiếu, không làm lại)

| PR | Issue | Nội dung |
|---|---|---|
| [#55](https://github.com/hoanghonghuy/filvault/pull/55) | [#7](https://github.com/hoanghonghuy/filvault/issues/7) | Playwright E2E critical-path smoke (mobile/tablet/desktop) |
| [#56](https://github.com/hoanghonghuy/filvault/pull/56) | [#17](https://github.com/hoanghonghuy/filvault/issues/17) | Remove misleading PRO/premium UI — free-only until entitlements exist |
| [#48](https://github.com/hoanghonghuy/filvault/pull/48) | [#6](https://github.com/hoanghonghuy/filvault/issues/6) | Responsive AppShell & navigation |
| [#46](https://github.com/hoanghonghuy/filvault/pull/46) | [#42](https://github.com/hoanghonghuy/filvault/issues/42) | Production-like runtime baseline + TLS edge proxy |
| [#44](https://github.com/hoanghonghuy/filvault/pull/44) | [#18](https://github.com/hoanghonghuy/filvault/issues/18) | Vault UX — align media preview & a11y |
| [#45](https://github.com/hoanghonghuy/filvault/pull/45) | [#41](https://github.com/hoanghonghuy/filvault/issues/41) | Health/readiness probes + migration release gate |
| [#12](https://github.com/hoanghonghuy/filvault/pull/12) | — | Upload queue (Files) |
| [#13](https://github.com/hoanghonghuy/filvault/pull/13) | [#3](https://github.com/hoanghonghuy/filvault/issues/3) | Secret lifecycle hardening |
| [#37](https://github.com/hoanghonghuy/filvault/pull/37) | [#19](https://github.com/hoanghonghuy/filvault/issues/19) | Vault shortcut routing |
| [#38](https://github.com/hoanghonghuy/filvault/pull/38) | [#20](https://github.com/hoanghonghuy/filvault/issues/20) | Multi-select batch actions |

### Issue backlog (theo ref — tra GitHub cho trạng thái hiện tại)

Nhóm **release / repository integrity:** [#43](https://github.com/hoanghonghuy/filvault/issues/43) **PARKED** → PR [#49](https://github.com/hoanghonghuy/filvault/pull/49) (external Docker/production-like runtime evidence; draft; MERGEABLE vs `develop` — không merge-ready; không undraft / không hạ evidence gate); [#50](https://github.com/hoanghonghuy/filvault/issues/50) IN_PROGRESS (ruleset `22665159` active — 5 required checks incl. `Web — E2E critical path` đã enforced; close after handoff #47 lands + human); epic [#4](https://github.com/hoanghonghuy/filvault/issues/4)

Nhóm **shell & cross-cutting:** [#8](https://github.com/hoanghonghuy/filvault/issues/8) IN_PROGRESS → PR [#54](https://github.com/hoanghonghuy/filvault/pull/54) (MERGEABLE vs `develop`; shell-language recon: ≤767 bottom nav; 768–1023 80px icon rail + page header; ≥1024 220px expanded sidebar — contract #48; chờ Designer/QA review); [#16](https://github.com/hoanghonghuy/filvault/issues/16), [#35](https://github.com/hoanghonghuy/filvault/issues/35)

Nhóm **theo surface (design):** [#21](https://github.com/hoanghonghuy/filvault/issues/21) IN_PROGRESS → PR [#58](https://github.com/hoanghonghuy/filvault/pull/58) (coherent Files search & filters; MERGEABLE vs `develop`; chờ human review); [#15](https://github.com/hoanghonghuy/filvault/issues/15) IN_PROGRESS → PR [#59](https://github.com/hoanghonghuy/filvault/pull/59) (unify appearance state; MERGEABLE vs `develop`; chờ human review); [#36](https://github.com/hoanghonghuy/filvault/issues/36) IN_PROGRESS → PR [#57](https://github.com/hoanghonghuy/filvault/pull/57) (`auth.user` canonicality đã xử lý — removed `mergeUserAfterSectionSave`, `auth.user` = PATCH response only, unsaved edits trong SettingsView-local refs; chờ human review); [#14](https://github.com/hoanghonghuy/filvault/issues/14), [#22](https://github.com/hoanghonghuy/filvault/issues/22)–[#35](https://github.com/hoanghonghuy/filvault/issues/35) — chi tiết AC trên từng issue

Nhóm **PM / handoff:** [#39](https://github.com/hoanghonghuy/filvault/issues/39) IN_PROGRESS → PR [#47](https://github.com/hoanghonghuy/filvault/pull/47) (draft)

### Việc **không** nên làm (đã xong hoặc lỗi thời)

- Bắt đầu Phase 2 slice S1–S6 — **đã Done** (xem [10](10-phase-2-status.md) lịch sử).
- Smoke tay Phase 1 UI như “việc tiếp theo” duy nhất — đã lệch thời; dùng [#43](https://github.com/hoanghonghuy/filvault/issues/43) / [#7](https://github.com/hoanghonghuy/filvault/issues/7) cho QA release.
- Browser E2E critical path ([#7](https://github.com/hoanghonghuy/filvault/issues/7) / [#55](https://github.com/hoanghonghuy/filvault/pull/55)) — **đã merge** trên `develop` (`21e2eeb`); `Web — E2E critical path` là required check trên ruleset.
- Production runtime baseline ([#42](https://github.com/hoanghonghuy/filvault/issues/42) / [#46](https://github.com/hoanghonghuy/filvault/pull/46)) — **đã merge** trên `develop` (`17bad79`).
- Vault UX polish ([#18](https://github.com/hoanghonghuy/filvault/issues/18) / [#44](https://github.com/hoanghonghuy/filvault/pull/44)) — **đã merge** trên `develop`.
- Deploy/AWS “sau Phase 2” — baseline runtime xong; release smoke ([#43](https://github.com/hoanghonghuy/filvault/issues/43)) **PARKED** on external Docker evidence trong epic [#4](https://github.com/hoanghonghuy/filvault/issues/4).
- Branch protection / required CI / review-thread resolution ([#50](https://github.com/hoanghonghuy/filvault/issues/50)) — **đã enforced** qua ruleset `22665159`; không tái cấu hình thủ công trừ khi đổi contract. Issue #50 vẫn mở đến khi handoff doc land + human đóng.
- Responsive AppShell / shell navigation ([#6](https://github.com/hoanghonghuy/filvault/issues/6) / [#48](https://github.com/hoanghonghuy/filvault/pull/48)) — **đã merge** trên `develop` (`48b16cb`).
- Misleading PRO/premium UI ([#17](https://github.com/hoanghonghuy/filvault/issues/17) / [#56](https://github.com/hoanghonghuy/filvault/pull/56)) — **đã merge**; free-only until entitlements exist (`0e156f1`).

---

## Tài liệu lịch sử (chỉ tra cứu)

| File | Nội dung | Ghi chú |
|---|---|---|
| [08-phase-1-status.md](08-phase-1-status.md) | Checklist Phase 1 (2026-08) | **Archival** — slice API 0–9 + smoke ghi chép |
| [10-phase-2-status.md](10-phase-2-status.md) | Checklist Phase 2 S1–S6 (2026-08) | **Archival** — 6/6 slice Done |
| [06-phase-1-plan.md](06-phase-1-plan.md) | Thứ tự slice Phase 1 | Spec/plan, không phải backlog hiện tại |
| [09-phase-2-features.md](09-phase-2-features.md) | Spec Phase 2 (Accepted) | Vẫn là spec; không phải trạng thái release |
| [11-chat-experience.md](11-chat-experience.md) | Spec Chat (Accepted) | C1–C7 Done; polish qua issues #30, #33 |

---

## Quy ước cập nhật

1. Khi merge issue/PR lớn ảnh hình release posture → cập nhật **file này** (không sửa “việc tiếp theo” trong 08/10).
2. Backlog mutable luôn sống trên GitHub issues/PRs — doc chỉ giữ ref và nhóm, không freeze checklist chi tiết.
3. Không ghi password/invite/token cố định trong handoff — chỉ tham chiếu biến `.env` / `.env.example`.
4. Lệch spec → sửa spec (hoặc ADR), không code xong rồi viết ngược.
