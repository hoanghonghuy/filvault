# Filvault Design System

**Status:** current product contract  
**Applies to:** `apps/web` across mobile, tablet, and desktop  
**Product/release truth:** [`docs/CURRENT_PRODUCT_STATE.md`](docs/CURRENT_PRODUCT_STATE.md)

This document defines the current UI/UX contract for Filvault. It replaces the former Phase-1-only design snapshot. Historical phase documents remain historical records and must not override this contract or the current product-state handoff.

Filvault is a personal cloud product: files, photos/media, sharing, trash/recovery, personal vault, chat/calls, account/settings, and appearance all belong to one coherent application. External products such as Google Drive, Dropbox, OneDrive, TeraBox, Messenger, and Telegram are pattern references only; do not copy their visual identity.

## 1. Product design principles

1. **Product-first utility.** Core actions must be obvious before decorative polish: browse, upload, organize, preview/download, share, recover, secure, and communicate.
2. **Mobile-first, not mobile-only.** Mobile optimizes reachability and progressive disclosure; tablet uses available width; desktop adds density, persistent navigation where appropriate, hover/focus affordances, keyboard efficiency, drag/drop, and multi-column layouts when useful.
3. **Progressive disclosure.** Keep primary surfaces calm. Secondary or destructive actions belong in contextual menus, sheets, drawers, or dialogs rather than competing with the primary task.
4. **Semantic consistency.** Use shared tokens and shared interaction patterns. Feature-specific brand/color exceptions must be explicit and scoped.
5. **State completeness.** Every significant surface must account for loading, empty, error, success, disabled, destructive, permission-denied, offline/retry where relevant, progress, duplicate/conflict, overflow, and long content.
6. **Accessible by default.** Keyboard, focus-visible, touch targets, labels, contrast, reduced motion, and screen-reader states are release criteria, not optional polish.
7. **Reactive localization.** User-facing copy, errors, dates/times, and ARIA labels follow the active locale. Do not compare behavior against translated display strings; use stable IDs/keys.

## 2. Appearance and themes

Filvault supports three app appearance modes through the canonical `useTheme()` service:

- `system`
- `light`
- `dark`

Color theme and appearance mode are independent. Current color themes include default, cyan, teal, sage, sunshine, peach, lavender, pearl, pebble, material, and seasonal spring/summer/autumn/winter variants.

### Rules

- Do not implement a second global dark-mode state inside a feature. All app-level appearance changes go through `useTheme()`.
- Theme selection may change the accent family, but **must not redefine semantic danger/success/warning meaning**.
- Dark mode is a first-class supported state. Any statement that Filvault has “no dark mode” is obsolete.
- Components should consume semantic CSS variables instead of baking a selected theme's hex value into feature CSS.
- Color alone must not be the only indicator of selected, error, success, destructive, or disabled state.

## 3. Semantic color contract

Prefer semantic variables already exposed by the web app, including the following roles where available:

- `--accent`, `--accent-soft`: app accent and soft selected/hover surfaces.
- text roles: primary/body/muted equivalents from shared styles.
- surface roles: page/canvas, card/surface, elevated/soft surface, hairline/border.
- semantic status roles: danger, warning, success and their soft variants.
- overlays and focus-visible rings.

### Chat exception

Chat is an intentional scoped brand exception. Use:

- `--chat-accent` for chat-owned accent semantics;
- `--chat-accent-secondary` / `--chat-surface-tint` where the conversation theme defines them;
- `var(--accent)` as the app-level fallback when an explicit chat accent is unavailable.

Do not spread raw Messenger blue (`#0084ff`) into shared shell/components. Existing migration work is tracked in #52.

### File/media category colors

MIME/file-category colors are informational category colors, not app accent or semantic status colors. They may remain fixed by category, but duplicated maps should converge on a shared source (#51). Do not route danger/success/warning through MIME colors.

## 4. Typography, spacing, and shape

Use the current shared typography/tokens from app styles. Maintain a clear hierarchy rather than view-local arbitrary sizes.

Recommended hierarchy:

- page/display title: 24–28px, semibold;
- section title: 18–22px, semibold;
- component/card title: 16px, semibold;
- body: 14–16px;
- metadata/caption: 12–14px.

Spacing follows a small consistent scale based around 4/8/12/16/24/32/48px. Prefer shared spacing variables/classes where present rather than introducing near-duplicate one-off values.

Controls should generally use modest radii; cards/sheets may use larger radii. Pill shapes are appropriate for chips, badges, compact segmented controls, or true pill actions—not as the default shape for every CTA.

## 5. Responsive contract

Use the product's established responsive behavior and keep these intent bands explicit:

### Mobile

- Single primary content column.
- Reachable primary actions and bottom-sheet/action-sheet disclosure where appropriate.
- Minimum interactive target: 44px; prefer 48px for primary touch controls.
- No hover-only information or actions.
- Long names/content must truncate or wrap intentionally without pushing critical controls off-screen.

### Tablet

- Use horizontal room instead of stretching a phone layout edge-to-edge.
- Grid/list density may increase.
- Drawers, split panes, or side-by-side content are allowed when they improve task continuity.
- Preserve touch usability even when pointer input is present.

### Desktop

- Do not render a scaled-up mobile screen.
- Use denser lists, persistent navigation/secondary panels where useful, keyboard/focus states, hover affordances, drag-and-drop, context actions, and multi-column layouts where the task benefits.
- Keep readable content widths for settings/forms rather than stretching text across the viewport.

## 6. Shared component patterns

### Navigation and shell

- Current shell navigation is the product contract; feature pages must not invent competing global navigation.
- The active destination must be visually and accessibly identifiable.
- Header actions use semantic accent tokens and must retain focus-visible states.

### Buttons/actions

- One visually dominant constructive action per local decision point whenever possible.
- Secondary actions use neutral/outlined/ghost treatment.
- Destructive actions must be explicitly destructive in label, treatment, and confirmation when data loss/revocation is meaningful.
- Disabled and loading states must prevent duplicate submission/action.

### Forms

- Visible labels; placeholders are not labels.
- Errors are actionable and localized; do not surface raw API/internal error strings directly.
- Loading/submitting state exposes disabled/`aria-busy` behavior where appropriate.
- Preserve autocomplete and password/OTP semantics.

### Modal / dialog / bottom sheet / drawer

- Trap focus when modal.
- Return focus to the trigger on close where practical.
- Escape/back behavior must not silently execute destructive work.
- Mobile may use a sheet where desktop uses a centered dialog or side panel, provided semantics stay equivalent.
- Destructive confirmation copy states the object/action clearly.

### Lists, grids, and file cards

- List view prioritizes density, metadata, keyboard movement, and bulk selection.
- Grid view prioritizes recognizability/preview while keeping filename and selection state usable.
- Selection must not depend on color alone.
- Context actions must remain discoverable on touch and keyboard, not hover only.

### Upload and progress

- Upload exposes progress and terminal success/failure.
- Retry/cancel are available when the underlying operation supports them.
- Desktop drag/drop is additive; file picker remains available.
- Do not fake completion before durable success is known.

### Empty/loading/error states

- Loading: skeleton/progress when useful; avoid layout jumps.
- Empty: explain the state and provide the next meaningful action if one exists.
- Error: explain recovery/retry, not raw implementation details.
- Partial failure: preserve successfully completed work and identify failed items.

## 7. Core journey review matrix

Any UI-heavy PR should consider the affected portion of this journey:

`onboarding/auth → upload → organize → search → preview/download → share → trash/recovery → settings/account/security`

Also review cross-cutting surfaces when relevant:

- Photos/media preview
- Personal Vault
- Chat/calls
- Theme/appearance
- Activity/history

For each affected surface, verify:

- mobile / tablet / desktop intent;
- light / dark appearance;
- active color themes do not break semantic contrast;
- loading / empty / error / success / disabled / destructive states;
- keyboard / focus-visible / touch behavior;
- localized visible copy and ARIA labels;
- overflow and long filenames/content.

## 8. Accessibility contract

- Interactive touch targets: minimum 44px unless a platform-native exception is justified.
- Keyboard users can reach and operate all primary actions.
- Focus-visible state must remain visible against every supported appearance/theme.
- Icon-only controls require an accessible name.
- Form controls require labels and useful error association.
- Dialog/sheet semantics and focus behavior must be correct.
- Respect `prefers-reduced-motion`; motion must not be required to understand state.
- Contrast must be reviewed under light/dark and representative accent themes.

## 9. Motion

Motion confirms state change; it is not decorative spectacle.

- Prefer transform/opacity for short transitions.
- Avoid `transition: all`.
- Respect reduced-motion globally.
- Loading/progress motion must not imply completion prematurely.
- Optimistic UI must have a deterministic rollback/reconciliation path on failure.

## 10. Feature-specific contracts

### Files / Search

Core actions: upload, create folder, organize, search/filter/sort, preview/download, move/share/trash, bulk selection. Search/filter state should remain understandable when zero results occur because filters—not because the library is empty.

### Sharing

Differentiate “Shared with me” and “My shares” clearly. Public-link copy/revoke/expiry/permission states must be localized, destructive revoke must be explicit, and dates must be locale aware. Current remaining localization work is tracked in #168.

### Personal Vault

Vault actions must communicate the security boundary clearly without implying stronger guarantees than implemented. Move-in/move-out errors must follow active locale; remaining Files→Vault localization is tracked in #160.

### Chat and Calls

Conversation personalization is scoped to the conversation. App appearance is global and uses `useTheme()`. Chat-owned accent styling uses the `--chat-accent` contract. Call controls must retain clear active/muted/disabled/error states and usable touch targets.

### Settings / Theme Center

Make scope explicit: app appearance mode, app color theme, account/security, and per-feature personalization are different concepts. Global appearance changes go through the canonical theme service.

## 11. PR design/QA gate

Before merging UI-impacting work, Reviewer + QA + Product Designer should verify the exact current head rather than author summary alone.

Minimum gate:

- acceptance criteria and product intent met;
- no material regression in core flow;
- mobile/tablet/desktop intent reviewed for UI-heavy changes;
- light/dark/theme contrast considered;
- loading/empty/error/destructive/permission states considered where relevant;
- keyboard/focus/touch accessibility considered;
- localization/date/ARIA impact considered;
- tests/CI for the scope are green;
- no unresolved material review finding.

A new head invalidates prior exact-head gate evidence.

## 12. Known design-system workstreams

This document is the shared contract, not a claim that all implementation is already conformant. Current tracked convergence work includes:

- #8 — design-system reconciliation parent;
- #16 — mixed-language UX across core journeys;
- #34 — app appearance vs per-chat personalization hierarchy;
- #51 — centralized MIME/category colors;
- #52 — semantic Chat accent contract.

When a new inconsistency is feature-specific or too large for the current PR, create/refine a focused issue with testable acceptance criteria instead of hiding it in this document or expanding into a giant PR.
