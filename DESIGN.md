---
version: current
name: Filvault-design-system
description: Utility UI for personal cloud storage — Cal.com-like canvas and confident hierarchy, selectable accent themes with light/dark appearance, mobile-first shell with Material 3 navigation and action patterns.
references:
  visual: Cal.com DESIGN.md (VoltAgent/awesome-design-md) — canvas, density, radius hierarchy, secondary surfaces
  interaction: Material Design 3 — bottom navigation, FAB, modal bottom sheets, touch targets
  product: Google Drive / Photos / Files — file list + photo grid dual surface, simplified top-level tabs

colors:
  ink: "#111827"
  body: "#374151"
  muted: "#6b7280"
  muted-soft: "#9ca3af"
  canvas: "#ffffff"
  surface: "#ffffff"
  surface-soft: "#f8fafc"
  surface-card: "#f3f4f6"
  hairline: "#e5e7eb"
  hairline-soft: "#f3f4f6"
  accent: "#2563eb"
  accent-hover: "#1d4ed8"
  accent-soft: "#eff6ff"
  danger: "#dc2626"
  danger-soft: "#fee2e2"
  warning: "#d97706"
  success: "#059669"
  on-accent: "#ffffff"
  on-ink: "#ffffff"
  primary-cta: "#2563eb"
  primary-cta-hover: "#1d4ed8"
  overlay: "rgba(17, 24, 39, 0.45)"

typography:
  font-display: "Plus Jakarta Sans, system-ui, sans-serif"
  font-body: "Plus Jakarta Sans, system-ui, sans-serif"
  font-mono: "ui-monospace, SFMono-Regular, Menlo, monospace"
  display: 28px / 600 / -0.02em
  title-lg: 22px / 600 / -0.01em
  title-md: 18px / 600
  title-sm: 16px / 600
  body: 16px / 400 / 1.5
  body-sm: 14px / 400 / 1.45
  caption: 12px / 500 / 1.4
  button: 14px / 600

rounded:
  sm: 6px
  md: 8px
  lg: 12px
  xl: 16px
  pill: 9999px

spacing:
  xxs: 4px
  xs: 8px
  sm: 12px
  md: 16px
  lg: 24px
  xl: 32px
  xxl: 48px
  shell-pad: 16px
  content-max: 1100px

touch:
  min-target: 44px
  preferred-target: 48px
  fab-size: 56px

motion:
  ease-standard: cubic-bezier(0.2, 0, 0, 1)
  ease-emphasized-decelerate: cubic-bezier(0.05, 0.7, 0.1, 1)
  ease-emphasized-accelerate: cubic-bezier(0.3, 0, 0.8, 0.15)
  duration-short: 100ms
  duration-medium: 200ms
  duration-long: 300ms

breakpoints:
  mobile: 0–767px
  tablet: 768–1023px
  desktop: 1024px+
---

# Filvault DESIGN.md

Nguồn sự thật cho **giao diện web hiện tại** (`apps/web`). Agent và người làm UI đọc file này trước khi thêm màn / component.

Tham chiếu visual: [Cal.com DESIGN.md](https://github.com/VoltAgent/awesome-design-md) (canvas sáng, hierarchy rõ, radius có tầng).  
Tham chiếu interaction: [Material Design 3](https://m3.material.io/) (bottom nav, FAB, bottom sheet).  
Nghiệp vụ / ràng buộc sản phẩm: [`docs/spec/07-phase-1-business.md`](docs/spec/07-phase-1-business.md), [`docs/spec/09-phase-2-features.md`](docs/spec/09-phase-2-features.md).

> **Trạng thái:** tài liệu này mô tả sản phẩm **đã ship** (light/dark, color themes, sharing, chat, vault, …), không chỉ Phase 1.

---

## 1. Visual Theme & Atmosphere

Filvault là **personal cloud** (My Files + Photos + Vault + Chat trên cùng một kho). UI phải đọc như phần mềm tiện ích đáng tin — gọn, một accent có thể đổi, không marketing hero.

**Tone:** calm utility, confident hierarchy, mobile-first.  
**Density:** list/file rows hơi dày (Cal.com product fragment feel); Photos grid thoáng hơn.  
**Motion:** ngắn (100–300ms), chỉ để xác nhận sheet mở/đóng, FAB press, toast, và phản hồi nhấn trên row/ô ảnh — không decorative, không trượt cả trang khi đổi tab.

### Motion tokens (Material 3)

| Token | Value | Dùng cho |
|-------|-------|----------|
| `--ease-standard` | `cubic-bezier(0.2, 0, 0, 1)` | Mọi transition mặc định |
| `--ease-emphasized-decelerate` | `cubic-bezier(0.05, 0.7, 0.1, 1)` | Sheet/trang đi vào |
| `--ease-emphasized-accelerate` | `cubic-bezier(0.3, 0, 0.8, 0.15)` | Sheet/trang rời đi |
| `--duration-short` | `100ms` | Press state, hover màu |
| `--duration-medium` | `200ms` | Fade, item move, toast ra |
| `--duration-long` | `300ms` | Sheet slide-up, toast vào |

### Motion tokens Filvault (nhấn/sheet/toast)

| Token | Value | Use |
|-------|-------|-----|
| `--motion-press` | `150ms` | Nút, FAB, row, ô ảnh |
| `--motion-enter` | `250ms` | Sheet / toast vào màn |
| `--motion-exit` | `200ms` | Sheet / toast ra (nhanh hơn vào) |

Rules:
- Chỉ animate `transform` + `opacity`; cấm `transition: all`.
- Vào chậm hơn ra (enter ≥ leave duration) — cảm giác phản hồi nhanh.
- Press feedback: `.btn` scale 0.97, `.row.tappable` scale 0.98 khi `:active`.
- Mọi motion phải được vô hiệu hoá bởi global `prefers-reduced-motion` guard trong `main.css`.
- Route change: `<Transition name="page" mode="out-in">` trong `AppShell` — fade-out 100ms, fade-in + translateY(8px) 200ms.
- List/grid thay đổi (xóa, move, search): `<TransitionGroup name="row">` — item rời đi scale-fade 100ms, item còn lại trượt vào chỗ (`move`) 200ms.
- Grid Photos / album list xuất hiện lần đầu: class `.appear` + stagger từ `lib/motion.ts` (`cellDelay`, 30ms/cell, tối đa 240ms).
- Bottom/side nav: indicator pill trượt bằng `transform: scale` (200ms emphasized-decelerate) — không animate layout.
- FAB: hover nâng shadow, press scale 0.94 + hạ shadow (100ms).
- Optimistic UI cho thao tác phá hủy (delete/restore): gỡ row khỏi list **trước** khi gọi API để TransitionGroup chạy move animation; lỗi thì `loadBrowser()`/`load()` rollback kèm error alert. Toast chỉ hiện khi API thành công.
- Router: luôn có `scrollBehavior` — reset về đầu trang khi điều hướng, khôi phục vị trí khi back/forward (`savedPosition`).
- Logout và mọi action thoát app phải đi qua confirm sheet.
- Bottom sheet bắt buộc có focus trap (Tab loop trong panel) + trả focus về trigger khi đóng.
- Upload: progress bar component (`UploadProgress`, `role="progressbar"` + `aria-live="polite"`), fill animate bằng `transform: scaleX` — không animate `width`. Files hỗ trợ drag & drop với overlay "Drop files to upload".
- Component dùng `<Transition>`: `BottomSheet`, `ToastHost`.

**Key characteristics**
- Canvas sáng (`{colors.canvas}`) hoặc tối khi `[data-theme='dark']`; surface phụ theo `--surface-soft` / `--surface-card`.
- Accent chromatic **có thể đổi** qua color theme (`data-color-theme`); mặc định ship là `default` (blue `#2563eb`). Theme `teal` (`#0d9488`) giữ bản sắc Filvault gốc.
- Primary destructive / confirm mạnh dùng `--danger` / ink CTA — không rainbow.
- Font: **Plus Jakarta Sans** (display + body). Không Inter / Roboto / Arial / system-only stack.
- Radius: controls `{rounded.md}` (8px), cards/sheets `{rounded.lg}`–`{rounded.xl}` (12–16px). Không pill CTA hàng loạt.
- **Dark mode đã ship.** Không purple glow marketing. Không glassmorphism.

---

## 2. Appearance & Theme System

### 2.1 Hai trục độc lập

| Trục | DOM / storage | Mô tả |
|------|---------------|-------|
| **AppearanceMode** (`system` \| `light` \| `dark`) | `data-theme="dark"` trên `<html>` khi resolved dark; `localStorage.filvault.appearanceMode` | Đảo semantic surface/text tokens qua `[data-theme='dark']` trong `main.css`. `system` theo `prefers-color-scheme` (reactive khi app mở). |
| **Color theme** (accent palette) | `data-color-theme="<id>"`; `localStorage.filvault.colorTheme` | Ghi đè `--accent`, `--accent-hover`, `--accent-soft`, `--primary-cta`, sidebar/header tints — **không** điều khiển light/dark appearance. |

**Contract (#59):** mọi thay đổi appearance (Settings, Theme Center, Chat quick toggle) phải đi qua `useTheme()` — không mutate `document.documentElement.dataset` hay `localStorage` rời rạc.

| API (`useTheme()`) | Role |
|--------------------|------|
| `appearanceMode` | `system` \| `light` \| `dark` (persisted) |
| `resolvedIsDark` / `isDarkMode` | Computed: appearance hiện tại sau resolve `system` |
| `setAppearanceMode(mode)` | Canonical setter — Settings / Theme Center radiogroup |
| `toggleResolvedAppearance()` | Chat quick toggle: flip resolved light ↔ dark |
| `currentColorTheme`, `applyColorTheme(id)` | Accent palette only |

Bootstrap trước Vue mount: inline script trong `index.html` + `hydrateAppearance()` trong `main.ts` (migrate legacy `filvault.theme`, `filvault.followSystemDark`, `colorTheme=dark`).

### 2.2 Color themes có sẵn

Định nghĩa trong `lib/theme.ts`, CSS trong `main.css` (`[data-color-theme='…']` + cặp dark appearance):

| Category | IDs |
|----------|-----|
| Colors | `default`, `cyan`, `teal`, `sage`, `sunshine`, `peach`, `lavender`, `pearl`, `pebble`, `material` |
| Seasonal | `spring`, `summer`, `autumn`, `winter` |

- **`dark` không còn là color-theme option** — legacy `colorTheme=dark` migrate về accent `default`; appearance dark giữ qua `appearanceMode`.
- Appearance và accent **độc lập**: có thể dùng `lavender` accent trên light hoặc dark.

UI: **Settings** (appearance radiogroup `system`/`light`/`dark`) + **Theme Center** (`/settings/theme`, `ThemeView.vue` — appearance + accent grid).

### 2.3 Semantic tokens — quy tắc bất biến

Các token **không được** bị color theme làm mờ nghĩa:

| Token | Role | Rule |
|-------|------|------|
| `--danger` / `--danger-soft` | Xóa vĩnh viễn, lỗi form/API | Luôn đỏ semantic; không dùng `--accent` cho destructive |
| `--warning` | Cảnh báo quota, favorite star fill | Vàng/cam semantic; duy nhất “điểm màu thứ hai” được phép ngoài accent |
| `--success` | Vault unlocked, toast success, quota OK | Xanh semantic; không trùng accent |
| `--on-accent` / `--on-ink` | Text trên nút filled | Phải đủ contrast trên nền tương ứng (light + dark + mỗi color theme) |

Color theme **có thể** ghi đè: `--accent`, `--accent-hover`, `--accent-soft`, `--primary-cta`, `--sidebar-bg`, `--header-bg`, `--hairline` (tint nhẹ). **Không** ghi đè `--danger`, `--success`, `--warning`.

### 2.4 Khi dùng `var(--accent)` vs màu cố định

| Loại | Quy tắc | Ví dụ |
|------|---------|-------|
| **(a) Semantic token** | Luôn `var(--…)` | Nav active, FAB, focus ring, storage fill, link hành động, CTA trong shell |
| **(b) File/media category** | Màu cố định theo MIME/loại, có thể tách lib | Icon Word/PDF/ảnh trong Files, Shared |
| **(c) Brand exception** | Documented, scoped CSS var riêng | Chat: `--chat-accent`, bubble outgoing — Messenger-style, **không** dùng cho shell chung |

**PR review:** shell/shared components không hardcode `#0084ff` hoặc `rgba(13,148,136,…)` — dùng `var(--accent)` hoặc `color-mix(in srgb, var(--accent) …)`.

### 2.5 Contrast checklist (Designer/QA)

- Auth form, list row, bottom sheet trên **light + dark**.
- Ít nhất một accent theme **high-chroma** (`lavender`, `sunshine`, `peach`).
- Destructive action vẫn đọc là danger khi accent là vàng/hồng/tím.
- Focus-visible ring 2px `--accent` vẫn thấy trên `--canvas` / `--surface-card`.

---

## 3. Color Palette & Roles (base light)

Giá trị `:root` / light trước khi ghi đè theme. Dark: xem `[data-theme='dark']` trong `main.css`.

| Token | Hex (light default) | Role |
|-------|---------------------|------|
| `ink` | `#111827` | Tiêu đề, primary text, CTA đậm |
| `body` | `#374151` | Body copy |
| `muted` | `#6b7280` | Meta, breadcrumb, placeholder |
| `canvas` | `#ffffff` | Nền trang / sheet |
| `surface-soft` | `#f8fafc` | Storage bar, nested bands, shell bg |
| `surface-card` | `#f3f4f6` | Empty state bg, inactive chip |
| `hairline` | `#e5e7eb` | Border 1px |
| `accent` | `#2563eb` (theme `default`) | Brand action: FAB, active nav, focus ring, storage fill |
| `accent-hover` | `#1d4ed8` | Hover/press accent |
| `accent-soft` | `#eff6ff` | Selected row tint, nav indicator |
| `danger` | `#dc2626` | Delete forever, lỗi |
| `danger-soft` | `#fee2e2` | Danger alert border/bg |
| `warning` | `#d97706` | Quota warning, favorites |
| `success` | `#059669` | Success states |
| `overlay` | `rgba(17,24,39,0.45)` | Backdrop bottom sheet / modal |

**Rules**
- Primary constructive action trên Files mobile = **FAB accent** (Upload); tablet/desktop ≥768px = toolbar button (FAB ẩn).
- Confirm nguy hiểm = danger text + secondary “Cancel”.
- Storage bar fill = accent; ≥90% quota → warning; ≥100% → danger.

---

## 4. Component Interaction States

Dùng làm **tiêu chí review PR** cho control/list/sheet. Mọi state phải có visual + a11y khi áp dụng.

| State | Visual | A11y / behavior |
|-------|--------|-----------------|
| **default** | Token nền/viền chuẩn | — |
| **hover** | Chỉ khi `@media (hover: hover)` — đổi bg/border nhẹ | Không thay thế focus |
| **focus-visible** | `outline: 2px solid var(--accent); outline-offset: 2px` | Bắt buộc cho button, link, input, nav item |
| **pressed / :active** | Scale 0.97 (`.btn`) / 0.98 (`.row.tappable`); FAB 0.94 | `prefers-reduced-motion`: bỏ scale |
| **selected** | `router-link-active`, segment/chip active: `--accent` text + `--accent-soft` bg | `aria-selected` / `aria-current` khi là tab/segment |
| **disabled** | `opacity: 0.55`, `cursor: not-allowed` | `disabled` hoặc `aria-disabled` |
| **loading** | Skeleton (`.skeleton`) hoặc spinner; nút giữ label + `aria-busy="true"` | Không double-submit |
| **empty** | `EmptyState`: icon + 1 câu + 1 CTA | Không illustration phức tạp |
| **error** | `.error`: `--danger-soft` bg + border; form field invalid | `role="alert"` cho banner |
| **success** | Toast success / inline notice | ToastHost, không block UI |
| **destructive** | `.btn.danger` hoặc label danger trong confirm sheet | Confirm sheet bắt buộc trước xóa vĩnh viễn |
| **permission-denied** | Copy ngắn + disabled action (Vault locked, share revoked) | Không leak chi tiết bảo mật |

Shared primitives: `.btn`, `.row.tappable`, `.card`, `.field`, `BottomSheet`, `ActionSheet`, `ConfirmSheet`, `UploadFab`, `EmptyState`, `ToastHost`, `GlobalConfirm`.

---

## 5. Typography Rules

Family: **Plus Jakarta Sans** (load via Google Fonts hoặc self-host). Fallback: `system-ui, sans-serif`.

| Role | Size | Weight | Use |
|------|------|--------|-----|
| Display / page title | 28px (mobile 24px) | 600 | “My Files”, “Photos” |
| Title md | 18px | 600 | Section (Albums, date group) |
| Title sm | 16px | 600 | Sheet title, card title |
| Body | 16px | 400 | Form, empty copy |
| Body sm | 14px | 400 | Row meta, helper |
| Caption | 12px | 500 | Badge, storage numbers |
| Button | 14px | 600 | Mọi button label |

Letter-spacing display: nhẹ âm (`-0.02em`). Không ALL-CAPS trừ badge kỹ thuật ngắn.

---

## 6. Component Stylings

### Auth card (Login / Register / Verify)

Full-screen centered card — **không** bottom nav. Cal.com-like: canvas trên nền `surface-soft`, hierarchy rõ.

**States:** default, focus (accent ring 2px), disabled/loading (`aria-busy`), error alert, success notice (verify resend).

**Accessibility:** `aria-labelledby` form ↔ h1; labels luôn visible; autocomplete đúng; không placeholder-only.

Component: `AuthCard.vue` + `LoginView`, `RegisterView`, `VerifyEmailView`.

### Overview (`/`)

Hub sau login. **Không** lặp toàn bộ shell nav. Shortcuts: Files, Photos, Shared, Chat (+ quick Trash từ card màu).

Trash và Vault **không** phải tab nav chính — vào qua Settings shortcut hoặc Files/Vault flow.

Component: `OverviewView.vue`. Logo F / wordmark → `/`.

### Buttons
- **Primary ink** (`primary-cta`): auth submit, “Save”, confirm không phá hủy. Height ≥44px; radius `{rounded.md}`.
- **Primary accent**: hiếm — verify hoặc brand moment. Prefer FAB cho Upload mobile.
- **Secondary**: border `{hairline}`, bg canvas.
- **Danger**: text/border danger; fill đỏ đặc chỉ “Delete forever”.
- **Ghost / linkish**: text accent hoặc muted; min hit area 44px.

### Inputs
- Height ≥44px; border `{hairline}`; focus ring 2px `{accent}`.

### List row (Files / Trash / Search / Shared)
- Full-width tap row; mobile → `⋯` action sheet; tablet/desktop ≥768 có thể inline 1–2 action.

### Photo grid
- `auto-fill`, min cell ~108–120px; `PhotoThumb` + thumbnail URL — **không** `<img>` original trong grid.

### Upload progress / FAB / Toast / Empty / Storage bar
- FAB shadow dùng `color-mix` với `var(--accent)`; ẩn FAB từ tablet upward (`≥768px`, toolbar thay thế).

### Bottom sheet / Action sheet / Confirm
- Backdrop `{overlay}`; focus trap; stacked full-width buttons trong confirm.

---

## 7. Layout & Responsive Shell

Source of truth: `lib/shellNav.ts` (`SHELL_BREAKPOINTS`, `SHELL_NAV_WIDTH`), `AppShell.vue`, CSS tokens `--bp-tablet` / `--bp-desktop` / `--nav-rail-w` / `--nav-sidebar-w` trong `main.css`.

### 7.1 Breakpoints (shell chung — khớp responsive-shell #48 / #6)

| Form factor | Viewport | Nav chrome | Page header | Upload |
|-------------|----------|------------|-------------|--------|
| **Mobile** | `≤767px` (`max-width: 767px`) | Bottom nav (5 tab) | Compact header (F + title + chat + avatar) | FAB |
| **Tablet** | `768px–1023px` | **80px icon rail** (`--nav-rail-w` / `SHELL_NAV_WIDTH.rail`), sticky; icons only (labels visually hidden) | **Giữ** compact page header | Toolbar (FAB ẩn) |
| **Desktop** | `≥1024px` (`min-width: 1024px`) | **220px expanded sidebar** (`--nav-sidebar-w` / `SHELL_NAV_WIDTH.sidebar`) với labels + profile block | **Ẩn** mobile header; title trong page content | Toolbar primary |

```text
Mobile (≤767)                    Tablet (768–1023)              Desktop (≥1024)
┌─────────────────────┐         ┌─┬──────────────────┐         ┌────────┬──────────────┐
│ [F] Filvault [chat] │         │F│ [F] title [chat] │         │ F Filv │ (no mobile   │
│     Page title      │         │█│ Storage          │         │ Home   │  header)     │
│ Storage             │         │█│ Main             │         │ Files  │ Storage      │
│ Main                │         │█│                  │         │ …      │ Main + title │
│ [FAB]               │         │█│                  │         │ avatar │ in view      │
│ Bottom nav (5)      │         │P│                  │         └────────┴──────────────┘
└─────────────────────┘         └─┴──────────────────┘
 80px rail = icons only           header KEPT                 220px sidebar + labels
 bottom nav + safe-area pad       bottom nav hidden            bottom nav hidden
```

**Không** gộp tablet với desktop: tablet là **icon rail 80px + header**, desktop là **sidebar 220px không header**.

### 7.2 Top-level shell navigation (`SHELL_NAV`)

Cùng 5 destinations trên bottom nav (mobile) và side nav (tablet rail / desktop sidebar):

1. **Home** (`/`) — Overview  
2. **Files** — browser + search + upload + vault entry  
3. **Photos** — timeline + albums  
4. **Shared** — shared with me  
5. **Settings** — prefs, theme link, trash shortcut, activity, share links  

**Không** trong shell nav: Trash (shortcut Settings/Overview), Profile (avatar → `/profile`), Chat (header shortcut hoặc Overview; surface bare).

Auth / verify: full-screen card, không shell.

### 7.3 Spacing
- Page padding: 16px mobile (`--space-md`); tablet `md` + `lg` horizontal; desktop 24px (`--space-lg`).
- Content max-width: `1100px`.
- `padding-bottom` shell-main ≥ `--bottom-nav-h` + `env(safe-area-inset-bottom)` — **mobile only** (tablet/desktop: `padding-bottom: 0`).

---

## 8. Depth & Elevation

| Level | Use |
|-------|-----|
| 0 | Canvas, list rows flat + hairline |
| 1 | FAB shadow, sticky header |
| 2 | Bottom sheet / modal |
| Overlay | Dim behind sheet |

Không multi-layer card stack hàng loạt.

---

## 9. Surface Map (sản phẩm hiện tại)

| Surface | Route | Shell | Shared system | Ghi chú |
|---------|-------|-------|---------------|---------|
| **Auth** | `/login`, `/register`, `/verify-email` | Bare | Auth card, ink CTA, `.error` | Không nav |
| **Overview** | `/` | Full | Cards, recent lists, quota | Shortcuts Chat/Trash |
| **Files** | `/files` | Full | List row, FAB, search bar, sort ActionSheet, `SearchFilterSheet`, filter chips, upload queue, batch bar | Search/filter URL-sync (`filesSearchState`); MIME icons = category (b); xem §9.4 |
| **Vault** | `/vault` | Full | PIN sheets, `MediaLightbox`, `--success`/`--warning` lock | Auto-lock on hidden |
| **Photos** | `/photos`, `/photos/albums/:id` | Full | Photo grid, albums, lightbox | Thumb only in grid |
| **Shared with me** | `/shared` | Full | List row, presign download | |
| **Trash** | `/trash` | Full | Optimistic restore/delete | Entry: Settings/Overview |
| **Settings** | `/settings` | Full | Cards, appearance radiogroup, activity, share links | `setAppearanceMode` → `useTheme()` |
| **Theme** | `/settings/theme` | Full | Theme grid, follow-system | Preview = product mock (xem follow-up) |
| **Profile** | `/profile` | Full | Form, avatar upload | Avatar in nav footer |
| **Chat** | `/chat`, `/chat/:id` | **Bare** (`meta.bare`) | Bubbles, composer, attachment queue, shared-media tabs, call modal | Breakpoint + interaction/a11y §9.1 |
| **Public share** | `/s/:token` | Bare (guest) | Centered card, ink download | Lỗi generic |

### 9.1 Chat — intentional exception

- **Không** dùng `AppShell` chrome.
- Layout: rail + thread + media panel (desktop).
- Breakpoints **riêng** (không thay shell §7):
  - Mobile `≤767`: master–detail (rail **hoặc** thread).
  - Tablet `768–1199`: 2 cột.
  - Desktop `≥1200`: 3 cột (media panel).
- `--chat-accent` / bubble outgoing: brand Messenger blue — **scoped Chat only**, không lan sang shell.
- Appearance quick toggle: `useTheme().toggleResolvedAppearance()` — không mutate DOM/storage rời (§2).

#### Composer vs attachment upload (#33 / #60)

Source: `ChatView.vue` — `sendingText` chỉ gate **gửi text** (textarea, Send, Like); **không** gate attach hay sticker.

| Control | Disabled khi `sendingText`? | Disabled khi upload đang chạy? |
|---------|----------------------------|--------------------------------|
| Attach (`triggerAttachment`) | **Không** | **Không** |
| Sticker picker | **Không** | **Không** |
| Textarea / Send / Like | **Có** | **Không** |

- User có thể **soạn tin và đính kèm file** trong lúc attachment upload đang chạy nền.
- Attachment queue **single-flight**: `processNextInAttachmentQueue()` chỉ cho **một** upload `uploading` tại một thời điểm; các file còn lại `queued`; khi hoàn tất / hủy / lỗi → dequeue tiếp theo. Không duplicate concurrent upload cho cùng queue.
- Upload gắn `conversationId` gốc — đổi thread không mất/cancel queue sai conversation.

#### Screen-reader upload progress

- **Một** live region cho upload: `UploadProgress` (`class="sr-only"`, `role="progressbar"`, `aria-live="polite"`) với `attachmentUploadProgressItems` + `attachmentAggregateProgress`.
- **Không** thêm `aria-live` cạnh tranh trên pending bubble hay hidden attachment-queue list (`attachment-queue sr-only` đã bỏ).

#### Shared-media tabs (a11y)

- Tab pattern WAI-ARIA: `role="tab"` + `aria-selected` + `aria-controls` → `role="tabpanel"` có `aria-labelledby`.
- Keyboard (`onMediaTabKeydown`): `ArrowLeft` / `ArrowRight` cycle tabs; `Home` → tab đầu; `End` → tab cuối (media panel desktop + shared-media trong info sidebar).
- Áp dụng cho media / file / link tabs trong media panel và chat-info shared media.

#### Desktop pane resizers (a11y)

- Rail và info sidebar: `role="separator"`, `tabindex="0"`, `aria-orientation="vertical"`, `aria-valuenow` / `aria-valuemin` / `aria-valuemax`.
- Pointer drag + double-click reset (`resetRailWidth` / `resetInfoWidth`).
- Keyboard (`onRailResizerKeydown` / `onInfoResizerKeydown`): arrow keys resize theo bước; `Home` reset default; `End` → max width.

### 9.2 Vault

- PIN setup/unlock/change; file list khi unlocked.
- Locked: warning semantic; unlocked: success semantic.
- Preview qua shared `MediaLightbox` + presign download API.

### 9.3 Sharing (đã ship)

- Share link sheet, public page, shared-with-me, activity log trong Settings — theo spec 09.

### 9.4 Files — search & filter (đã ship, #21 / #58)

- **Sort** (icon `sort`, ActionSheet) và **search filters** (icon `sliders`, `SearchFilterSheet`) là hai affordance tách biệt; ARIA label i18n (`sortFilesAria` / `searchFiltersAria`).
- Filter sheet (`SearchFilterSheet` trên `BottomSheet`): type (`all` / `image` / `video` / `document` / `archive` / `folder`), sort (`relevance` / `name` / `date` / `size`), order (`asc` / `desc`), date range `from`/`to`, optional folder scope (`folderId`).
- Validate date range trước Apply — `from > to` → inline `role="alert"`; không gọi API.
- Active filter chips map 1:1 với applied state (`getActiveFilterChipKeys`); mỗi chip removable; **Clear filters** reset filter state nhưng **giữ** search query `q`.
- URL sync (`filesSearchState.ts`): `q`, `type`, `sort`, `order`, `from`, `to`, `sfolderId` — refresh / back / forward khôi phục trạng thái.
- Empty search: copy theo nguyên nhân (query only, filters only, hoặc cả hai) + recovery (`clearSearch`, `clearFilters`).
- Browse mode: sort local (`updatedAt` / `name` / `size`); search mode: API query qua `buildSearchApiQueryString`.

---

## 10. Do's and Don'ts

### Do
- Mobile-first: thiết kế 375px trước.
- Sheet thay `prompt`/`confirm`.
- Touch target ≥44px.
- `useTheme()` cho appearance (`setAppearanceMode`) và accent (`applyColorTheme`).
- Photos: thumbnail + presign cho original.
- Test light + dark + một accent theme trước merge UI.

### Don't
- Đừng viết “no dark mode” — đã ship.
- Đừng mô tả tablet (768–1023) như sidebar 220px — tablet là **80px icon rail + page header**.
- Đừng hardcode accent shell (`#0084ff`, teal rgba) khi có `var(--accent)`.
- Đừng dùng accent cho destructive/success.
- Đừng `<img>` original trong Photos grid.
- Đừng >5 item bottom nav (hiện tại = 5, không thêm tab).
- Đừng FAB cho navigation.
- Đừng emoji làm icon hệ thống.

---

## 11. Entitlements & premium presentation (current policy)

**Chưa có mô hình premium / subscription.** Cho đến khi có nguồn entitlement thật (API hoặc billing):

- **Không** hiển thị badge PRO, crown, hoặc affordance “khóa” trên profile, theme, hoặc settings.
- **Không** đánh dấu theme là premium (`isPro`) trong định nghĩa hoặc UI nếu chưa gate hành vi tương ứng.
- **Không** thêm CTA upgrade, luồng mua giả, hoặc trạng thái “locked” chỉ mang tính trang trí.
- Tất cả theme trong Theme Center **áp dụng tự do**; Settings và Theme Center phải **nhất quán** (cùng trạng thái free, không PRO).

Khi monetization sẵn sàng: thêm entitlement source, gate theme/affordance theo quyền thật, và cập nhật mục này — không giữ UI premium placeholder.

---

## 12. Agent Prompt Guide

```text
Follow DESIGN.md (Filvault). Cal.com-like utility UI, semantic tokens via CSS variables,
AppearanceMode system/light/dark + selectable accent color themes (useTheme), mobile-first shell:
≤767 bottom nav (5 tabs) + header; 768–1023 80px icon rail + header;
≥1024 220px sidebar, no mobile header. SHELL_BREAKPOINTS / SHELL_NAV_WIDTH in shellNav.ts.
avatar → Profile, chat → bare /chat, FAB upload on Files (mobile ≤767 only),
bottom sheets not prompt/confirm, component states per §4,
danger/success/warning never replaced by accent, touch ≥44px.
Chat may use --chat-accent; shell must use var(--accent).
Chat §9.1: composer stays usable during attachment upload; single-flight queue;
one UploadProgress SR region; shared-media tabs + desktop resizers keyboard-accessible.
No PRO badges, crowns, or premium placeholders until real entitlements exist (§11).
```

Quick tokens: `--accent`, `--ink`, `--canvas`, `--danger`, `--success`, radius 8px controls / 12–16px cards, Plus Jakarta Sans.

---

## 13. Implementation notes (`apps/web`)

1. Tokens → CSS variables: `src/assets/main.css` (`:root`, `[data-theme='dark']`, `[data-color-theme='…']`).
2. Theme logic: `src/lib/theme.ts` (`AppearanceMode`, `useTheme`, `hydrateAppearance`, `THEMES`); bootstrap inline trong `index.html`.
3. Shell: `AppShell.vue` + `lib/shellNav.ts` (`SHELL_BREAKPOINTS`, `SHELL_NAV_WIDTH`, `isShellNavActive`).
4. Shared: `BottomSheet`, `ActionSheet`, `ConfirmSheet`, `UploadFab`, `EmptyState`, `ToastHost`, `MediaLightbox`.
5. Tests contract: `AppShell.test.ts`, `shellNav.test.ts`, `theme.test.ts`, `dark-mode.test.ts`, `vault-view.test.ts`, `chat.test.ts`, `ChatView.component.test.ts` (composer/upload/a11y §9.1).
6. Không thêm UI library nặng trừ khi duyệt — CSS + Vue SFC.

### Follow-up debt (GitHub issues riêng — không giấu trong prose)

- Chuẩn hoá MIME/category colors vào một module.
- Chat: giảm hardcode `#0084ff` ngoài `--chat-accent` contract.
- Theme Center: tab Icons/Display chưa ship; preview mock calendar → Filvault surfaces.
- `EmojiPicker` / `CallModal` fallback colors.

---

*Chốt hướng visual: **A — Cal.com-like** (2026-08-17). Cập nhật contract sản phẩm hiện tại: 2026-09-09 (#8). Shell 3 form-factor: 2026-09-09 (#48 / #54). AppearanceMode unified: 2026-09-09 (#59). Chat composer/upload a11y: 2026-09-10 (#60 / #33).*
