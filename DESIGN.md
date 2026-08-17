---
version: alpha
name: Filvault-design-system
description: Light utility UI for personal cloud storage — Cal.com-like white canvas and confident hierarchy, Filvault teal as the single chromatic accent, mobile-first shell with Material 3 navigation and action patterns.
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
  accent: "#0d9488"
  accent-hover: "#0f766e"
  accent-soft: "#ccfbf1"
  danger: "#dc2626"
  danger-soft: "#fee2e2"
  warning: "#d97706"
  success: "#059669"
  on-accent: "#ffffff"
  on-ink: "#ffffff"
  primary-cta: "#111827"
  primary-cta-hover: "#1f2937"
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

breakpoints:
  mobile: 0–767px
  tablet: 768–1023px
  desktop: 1024px+
---

# Filvault DESIGN.md

Nguồn sự thật cho **giao diện web Phase 1** (`apps/web`). Agent và người làm UI đọc file này trước khi thêm màn / component.

Tham chiếu visual: [Cal.com DESIGN.md](https://github.com/VoltAgent/awesome-design-md) (canvas trắng, hierarchy rõ, radius có tầng).  
Tham chiếu interaction: [Material Design 3](https://m3.material.io/) (bottom nav, FAB, bottom sheet).  
Nghiệp vụ / ràng buộc sản phẩm: [`docs/spec/07-phase-1-business.md`](docs/spec/07-phase-1-business.md).

---

## 1. Visual Theme & Atmosphere

Filvault là **personal cloud** (My Files + Photos trên cùng một kho). UI phải đọc như phần mềm tiện ích đáng tin — sáng, gọn, một accent, không marketing hero.

**Tone:** calm utility, confident hierarchy, mobile-first.  
**Density:** list/file rows hơi dày (Cal.com product fragment feel); Photos grid thoáng hơn.  
**Motion:** ngắn (150–250ms), chỉ để xác nhận sheet mở/đóng, FAB press, toast — không decorative.

**Key characteristics**
- Canvas trắng (`{colors.canvas}`), surface phụ xám rất nhạt (`{colors.surface-soft}` / `{colors.surface-card}`).
- Một accent chromatic: teal Filvault (`{colors.accent}` — `#0d9488`). Dùng cho FAB, tab active, link hành động, progress storage.
- Primary destructive / confirm mạnh dùng ink (`{colors.primary-cta}`) hoặc danger — không rainbow.
- Font: **Plus Jakarta Sans** (display + body). Không Inter / Roboto / Arial / system-only stack.
- Radius: controls `{rounded.md}` (8px), cards/sheets `{rounded.lg}`–`{rounded.xl}` (12–16px). Không pill CTA hàng loạt.
- Không dark mode Phase 1. Không purple glow. Không glassmorphism.

---

## 2. Color Palette & Roles

| Token | Hex | Role |
|-------|-----|------|
| `ink` | `#111827` | Tiêu đề, primary text, CTA đậm (Cal.com-like black button) |
| `body` | `#374151` | Body copy |
| `muted` | `#6b7280` | Meta, breadcrumb, placeholder |
| `canvas` | `#ffffff` | Nền trang / sheet |
| `surface-soft` | `#f8fafc` | Storage bar, nested bands |
| `surface-card` | `#f3f4f6` | Empty state bg, inactive chip |
| `hairline` | `#e5e7eb` | Border 1px |
| `accent` | `#0d9488` | Brand action: FAB, active nav, focus ring, storage fill |
| `accent-hover` | `#0f766e` | Hover/press accent |
| `accent-soft` | `#ccfbf1` | Selected row tint, soft badge |
| `danger` | `#dc2626` | Delete forever, lỗi |
| `danger-soft` | `#fee2e2` | Danger button border/bg nhẹ |
| `overlay` | `rgba(17,24,39,0.45)` | Backdrop bottom sheet / modal |

**Rules**
- Primary constructive action trên màn Files = **FAB accent** (Upload), không nhân bản nhiều nút primary cùng lúc.
- Confirm nguy hiểm = danger text + ink secondary “Cancel”.
- Storage bar fill = accent; ≥90% quota → warning; ≥100% → danger.

---

## 3. Typography Rules

Family: **Plus Jakarta Sans** (load via `fonts.google.com` hoặc self-host). Fallback: `system-ui, sans-serif`.

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

## 4. Component Stylings

### Auth card (Login / Register / Verify)

Full-screen centered card — **không** bottom nav. Cal.com-like: canvas trắng trên nền `surface-soft`, hierarchy rõ.

**Anatomy**
```text
[Brand: mark + “Filvault”]     ← trên card, căn giữa
┌─────────────────────────┐
│ Title (h1)              │
│ Subtitle (muted, 1 câu) │
│ Form fields (label trên)│
│ Error alert (nếu có)    │
│ Primary ink CTA (full)  │
│ ─────────────────────── │
│ Footer switch link      │
└─────────────────────────┘
```

| Token | Value | Ghi chú |
|-------|-------|---------|
| Shell max-width | 420px | §8 responsive |
| Card padding | 24px (`space-lg`) | |
| Card radius | 16px (`radius-xl`) | |
| Field min height | 44px | touch target |
| CTA | ink `#111827`, full width | §4 Buttons |
| Error | `danger-soft` bg + border | `role="alert"` |
| Link switch | accent + underline | hit area ≥44px |

**States:** default, focus (accent ring 2px), disabled/loading (opacity + `aria-busy`), error alert, success notice (verify resend).

**Accessibility:** `aria-labelledby` form ↔ h1; labels luôn visible; autocomplete đúng (`username`, `current-password`, `new-password`, `one-time-code`); không placeholder-only. Màn ngắn: page scroll từ trên, không cắt field đầu.

**Anti-patterns:** không bottom nav; không nhiều CTA primary; không placeholder thay label; không lỗi API raw; không `autocapitalize` invite (mã phân biệt hoa/thường).

Verify: tách state Verifying / Resend — không đổi nhãn cả hai nút cùng lúc.

Component: `AuthCard.vue` + views `LoginView`, `RegisterView`, `VerifyEmailView`.

### Buttons
- **Primary ink** (`primary-cta`): dùng cho auth submit, “Save”, confirm không phá hủy. Height ≥44px mobile; radius `{rounded.md}`; text `{on-ink}`.
- **Primary accent**: hiếm — chỉ khi action mang brand (vd. “Verify”). Prefer FAB cho Upload.
- **Secondary**: border `{hairline}`, bg canvas, text ink.
- **Danger**: text/border danger; không fill đỏ đặc trừ “Delete forever” trong sheet.
- **Ghost / linkish**: text accent hoặc muted; min hit area vẫn 44px.

### Inputs
- Height ≥44px; padding 12px 14px; border `{hairline}`; focus ring 2px `{accent}` (không outline browser mặc định).
- Label trên field (`field` stack), không placeholder-only.

### List row (Files / Trash / Search)
- Full-width tap row: icon/type + name (1 dòng truncate) + meta phụ.
- Actions **không** xếp 3–4 nút ngang trên mobile → một nút `⋯` mở **Action sheet**.
- Desktop (≥768): có thể hiện 1–2 action phụ + `⋯`.

### Photo grid
- `auto-fill`, min cell ~108–120px mobile; gap `{spacing.xs}`–`{spacing.sm}`.
- Cell = **placeholder** theo mime (không `<img>` original — spec Phase 1).
- Tap → sheet: Xem / Tải / Thêm vào album / Xóa (nếu context cho phép).

### Bottom sheet / Action sheet
- Backdrop `{overlay}`; sheet bg canvas; top radius `{rounded.xl}`; drag handle 32×4px muted.
- Max height ~90vh; safe-area padding đáy.
- Confirm sheet: title + short body + stacked full-width buttons (primary / cancel / danger).

### FAB
- Size `{touch.fab-size}` (56px); bg `{accent}`; icon trắng; shadow nhẹ (elevation 1).
- Chỉ trên **Files** (và có thể Photos nếu upload media — Phase 1: Files là chính).
- Vị trí: bottom-end, trên bottom nav (~16px + safe-area).

### Toast / snackbar
- Bottom, trên nav; bg ink; text on-ink; auto-dismiss ~3s; một toast tại một thời điểm.

### Empty state
- Icon đơn giản + 1 câu + 1 CTA (không illustration phức tạp).

### Storage bar
- Mobile: track + “used / quota” caption; có thể rút label “Storage”.
- Reload sau upload / permanent delete.

---

## 5. Layout Principles

### Shell
```text
Mobile (<768)                         Tablet/Desktop (≥768)
┌─────────────────────┐               ┌──────┬──────────────────┐
│ Header (brand+title)│               │ Side │ Header           │
│ Storage (compact)   │               │ nav  │ Storage          │
│                     │               │ 4    │                  │
│ Main (scroll)       │               │ tabs │ Main             │
│                     │               │      │                  │
│ [FAB]               │               └──────┴──────────────────┘
│ Bottom nav (4)      │               FAB → toolbar button
└─────────────────────┘
```

### Top-level destinations (đúng M3: 3–5, chỉ navigation)
1. **Files** — browser + search + upload  
2. **Photos** — timeline + albums  
3. **Trash**  
4. **Settings**

Auth / verify: full-screen card giữa, không bottom nav.

### Spacing
- Page padding ngang: `{spacing.shell-pad}` (16px) mobile; 24px desktop.
- Content max-width: `{spacing.content-max}` (1100px) căn giữa.
- Section gap: `{spacing.lg}`–`{spacing.xl}`.
- `padding-bottom` main ≥ bottom-nav height + FAB clear + `env(safe-area-inset-bottom)`.

### Breadcrumb (Files)
- Mobile: `← Parent` hoặc truncate giữa `Root / … / Current`.
- Desktop: full trail.

---

## 6. Depth & Elevation

| Level | Use |
|-------|-----|
| 0 | Canvas, list rows flat + hairline |
| 1 | FAB shadow, sticky header optional |
| 2 | Bottom sheet / modal |
| Overlay | Dim behind sheet |

Không multi-layer card stack. Row = border, không drop-shadow hàng loạt.

---

## 7. Do's and Don'ts

### Do
- Mobile-first: thiết kế 375px trước, mở rộng lên.
- Thay `window.prompt` / `confirm` bằng sheet.
- Touch target ≥44px (ưu tiên 48px).
- Photos: placeholder + presign on demand (spec 07).
- Một việc chính mỗi màn; overflow vào sheet.
- Báo lỗi API bằng copy ngắn + code ẩn trong detail nếu cần (`QUOTA_EXCEEDED`, `409`).

### Don't
- Đừng dùng Inter / Roboto / Arial làm font chính.
- Đừng purple / indigo gradient theme; đừng cream+serif “AI brochure”.
- Đừng `<img src>` original trong grid Photos.
- Đừng nhồi >5 item vào bottom nav.
- Đừng dùng FAB cho navigation.
- Đừng dark mode Phase 1.
- Đừng emoji làm icon hệ thống (SVG đơn giản).

---

## 8. Responsive Behavior

| Breakpoint | Nav | Upload | Actions |
|------------|-----|--------|---------|
| `<768` | Bottom nav | FAB | Action sheet |
| `768–1023` | Side rail hoặc bottom giữ + wider | Toolbar + optional FAB | Sheet hoặc menu |
| `≥1024` | Side nav persistent | Toolbar primary | Inline + menu |

- Photo grid: 3 cột hẹp → 4–6 cột rộng.
- Auth card: max-width 420px, luôn center.
- Keyboard mở: sheet tránh bị che (visual viewport / padding).

---

## 9. Screen map (Phase 1 — đủ API)

| Screen | Must-have UI | API |
|--------|--------------|-----|
| Login / Register | Form ink CTA | auth |
| Verify email | OTP 6 số + resend | verify / resend |
| Files | Browser, search, create folder, upload, rename/move/delete file & folder | browser, folders, files, search |
| Photos | Timeline (+ load more), albums CRUD, add/remove items, open/download | photos/* |
| Album detail | Grid + manage | albums/:id, items |
| Trash | Restore / delete forever | trash, restore |
| Settings | displayName, trash prefs, password, logout | users/me, password |

---

## 10. Agent Prompt Guide

Khi generate / sửa UI:

```text
Follow DESIGN.md (Filvault). Cal.com-like light utility UI, teal accent #0d9488,
Plus Jakarta Sans, mobile-first bottom nav (Files/Photos/Trash/Settings),
FAB upload on Files, bottom sheets instead of prompt/confirm,
no original <img> in Photos grid, touch targets ≥44px, no purple/dark-mode.
```

Quick tokens:
- Accent: `#0d9488`
- Ink CTA: `#111827`
- Canvas: `#ffffff`
- Hairline: `#e5e7eb`
- Radius control: `8px` · card/sheet: `12–16px`
- Font: Plus Jakarta Sans

---

## 11. Implementation notes (`apps/web`)

1. Map tokens → CSS variables trong `src/assets/main.css`.
2. Shell: refactor `AppShell.vue` (bottom nav + desktop side).
3. Shared: `BottomSheet`, `ActionSheet`, `ConfirmSheet`, `Fab`, `EmptyState`, `Toast`.
4. Không thêm UI library nặng Phase 1 trừ khi được duyệt — CSS + Vue SFC trước.
5. Mọi màn mới phải khớp §9 và gap API đã liệt kê trong kế hoạch W1–W6.

---

*Chốt hướng visual: **A — Cal.com-like** (2026-08-17).*
