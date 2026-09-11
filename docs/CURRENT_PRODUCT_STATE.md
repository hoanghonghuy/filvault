# Filvault — Current Product & Release State

> **Authoritative current handoff.** This document describes what should be worked on now. Historical phase documents under `docs/spec/*phase*-status.md` are useful records, but their old “next work” sections are not authoritative when they conflict with this file or current GitHub issues/PRs.

Last reconciled: **2026-09-11**  
Control/integration branch: **`develop`**

## Product goal

Filvault is being driven toward a production-ready personal cloud product, not an MVP/demo checkpoint. Product work should prioritize usable end-to-end journeys, security/data integrity, recovery/error states, responsive UX, accessibility and release readiness before low-impact cleanup.

## Current capability state

| Area | Functional state | Current posture |
|---|---|---|
| Auth/account | Login, register, verify email, password recovery/reset, profile/password settings | Functional; recent VI/EN localization and safe auth-error mapping merged |
| Files/folders | Browse, search/filter/sort, create folders, upload/download, move/rename/delete, favorites, batch actions | Functional; ongoing localization/UX hardening |
| Photos/albums/media | Timeline, albums, preview/lightbox, media preferences | Functional; continue UX/accessibility/performance audit |
| Sharing | Public links, Shared With Me / My Shares | Functional; #168 tracks remaining mixed-language/loading/action/ARIA states |
| Chat/calls | Messaging, reactions/customization, calls | Functional; #52/#34 track remaining semantic-accent/personalization hierarchy debt |
| Personal Vault | Protected storage and move flows | Functional; #160 tracks remaining Files → Vault localized failure copy |
| Settings/themes | Profile, password, trash/media preferences, appearance/locale/theme | Functional; major localization pass merged; design contract still needs reconciliation (#8) |
| Trash/activity | Restore/purge and activity surfaces | Functional; continue edge/error/accessibility audit |
| CLI | Existing CLI remains part of CI | Functional; keep green through product changes |
| Production-like runtime | TLS edge, readiness/liveness, migration gate and runtime baseline | Baseline merged; final release-smoke/restart/rollback evidence is blocked on Docker-capable runtime (#43 / PR #49) |

“Functional” does **not** mean release-ready. Open issues and exact-head CI/review evidence remain the source of truth for hardening work.

## Active product workstreams

### P1 — core product / release

- **#16 — mixed-language UX:** IN_PROGRESS. Auth, Call, ShareSheet and Settings slices are merged. Confirmed remaining slices include **#160 Files → Personal Vault** and **#168 Shared With Me / My Shares**.
- **#39 — current product/release handoff:** this document is the replacement current-state entry point. Historical phase handoffs still need explicit archival/reconciliation treatment.
- **#8 — design-system reconciliation:** current theme/dark-mode/product surfaces have outgrown the old Phase-1 design contract.
- **#87 — web dependency vulnerabilities:** READY; remediate with evidence-driven minimal upgrades, never blanket `npm audit fix --force`.
- **#43 / PR #49 — release smoke + restart persistence + rollback:** implementation exists but remains PARKED/draft until a Docker-capable production-like host produces the required runtime evidence.

### Product design follow-ups

- **#52 — Chat semantic accent contract:** EmojiPicker, shell shortcut and CallModal slices are done; `ChatView.vue` remains the confirmed semantic-accent debt.
- **#34 — personalization hierarchy:** clarify app appearance vs per-conversation theme/wallpaper/bubble scope.
- **#51 — MIME/category colors:** consolidate duplicated file-category color mapping without mixing it with semantic status/accent colors.

## Release gates

A change is not DONE merely because code exists. For implementation PRs:

1. Branch from current `develop`; no implementation push directly to `develop`/`main`.
2. Run the maximum relevant local checks available before push.
3. Required CI on the **exact current head** must be green.
4. Review the exact current head from Reviewer + QA + Product Designer perspectives; head changes invalidate prior gate evidence.
5. Do not weaken tests, branch rules, security, accessibility or acceptance criteria to obtain green status.
6. Merge only when the PR is mergeable and material review threads/blockers are resolved.
7. After merge, verify `develop`, reconcile the issue, and unblock/re-scan the next workstream.

## Production posture

The production-like baseline already includes production configuration validation, TLS edge/reverse proxy, migration gating and `/healthz`/`/readyz` contracts. The remaining production epic blocker is runtime proof for #43: authenticated release smoke, persistence after restarting stateless app services, forced-failure diagnostics, and rollback/recovery verification on a Docker-capable host.

Do **not** treat old Phase-1 text saying “deploy/AWS is not next” as current direction.

## Local development

Use `.env` / `.env.example` for local credentials and invite codes. Current handoff documentation must not publish fixed password/invite values.

Typical local commands:

```bash
make up
make test
make dev-api
make dev-web
```

See the root `README.md`, `deploy/production/README.md`, accepted specs/ADRs, and current GitHub issues/PRs for implementation details.

## How to choose the next task

Use this order:

1. Broken/missing core flows.
2. Auth/security/privacy/data integrity/file recovery.
3. End-to-end loading/error/permission/retry UX.
4. Responsive/design-system/accessibility gaps.
5. Performance/reliability/release readiness.
6. Technical cleanup only when it has measurable product impact.

If READY supply is low, audit the actual product and create a small-to-medium vertical slice with explicit goal, scope, testable acceptance criteria, responsive/error states and QA evidence. Avoid giant PRs.
