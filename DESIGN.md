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

Nguồn sự thật cho **giao diện web Phase 1** (`apps/web`). Agent và người làm UI đọc file này trước khi thêm màn / component.

Tham chiếu visual: [Cal.com DESIGN.md](https://github.com/VoltAgent/awesome-design-md) (canvas trắng, hierarchy rõ, radius có tầng).  
Tham chiếu interaction: [Material Design 3](https://m3.material.io/) (bottom nav, FAB, bottom sheet).  
Nghiệp vụ / ràng buộc sản phẩm: [`docs/spec/07-phase-1-business.md`](docs/spec/07-phase-1-business.md).

---

## 1. Visual Theme & Atmosphere

Filvault là **personal cloud** (My Files + Photos trên cùng một kho). UI phải đọc như phần mềm tiện ích đáng tin — sáng, gọn, một accent, không marketing hero.

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
| `--motion-press` | 150ms | Nút, FAB, row, ô ảnh |
| `--motion-enter` | 250ms | Sheet / toast vào màn |
| `--motion-exit` | 200ms | Sheet / toast ra (nhanh hơn vào) |

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

### Overview (`/`)

Hub sau login. **Không** lặp 4 tab nav (Files/Photos/Trash/Settings đã ở bottom/side). Trash và Settings không phải lối vào chính.

**Anatomy**
```text
[Name]                          ← title-lg, ink
[used of quota]                 ← caption muted (từ users/me)
┌────────────┐ ┌────────────┐
│ My Files   │ │ Photos     │   ← 2 destination cards
│ Browse…    │ │ Timeline…  │
└────────────┘ └────────────┘
Recent files              See all
  rows (hoặc 1 dòng empty + link)
Recent photos             See all
  PhotoThumb grid (thumbnailUrl; fallback mime icon — không <img> original)
```

| Token | Value | Ghi chú |
|-------|-------|---------|
| Hero title | 22px mobile / 28px desktop | `displayName` |
| Cards | 2 cột, radius-lg, border hairline | chỉ Files + Photos |
| Card min-height | 96px | chạm dễ |
| Section gap | 32px (`space-xl`) | |
| Empty | 1 dòng + link accent | không EmptyState lồng card |

**Anti-patterns:** 4 ô trùng bottom nav; card lồng card; empty dashed lớn; FAB trên Overview.

Component: `OverviewView.vue`. Logo F / wordmark → `/`.

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
- Cell = `PhotoThumb`: `thumbnailUrl` khi có (server thumb), không thì icon theo mime — **không** `<img>` original.
- Tap → sheet: Xem / Tải / Thêm vào album / Xóa (nếu context cho phép).

### Upload progress (Files)
- Khi đang upload: `LinearProgress` (track hairline 4px, fill accent) + label + % — không chỉ text “Uploading… N%”.
- `role="progressbar"` + `aria-valuenow` / min / max.

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
│ [F] Filvault        │               │ Side │ (no top header)  │
│     My Files        │               │ nav  │ Storage          │
│ Storage (compact)   │               │ 4    │ Page title in    │
│ Main (scroll)       │               │ tabs │ main             │
│ [FAB]               │               │      │                  │
│ Bottom nav (4)      │               └──────┴──────────────────┘
└─────────────────────┘               FAB → toolbar button
```

Mobile header: mark 32px (ô **F** là nút Home, chạm ≥44px) + tên app caption muted + page title ink + **avatar initials** (phải, chạm ≥44px → `/profile`). Bấm **F** (mobile) hoặc wordmark sidebar (desktop) → `/` Overview. Title Files = “My Files”. Header ẩn ≥768 — title nằm trong view; avatar nằm đáy side nav. Overview **không** thêm tab thứ 5 vào bottom nav.

### Top-level destinations (đúng M3: 3–5, chỉ navigation)
**Home / Overview** — vào bằng logo (ô F / wordmark), không phải tab.  
**Profile** — vào bằng avatar (header mobile / side-nav desktop), không phải tab.

Bottom / side nav (4):
1. **Files** — browser + search + upload  
2. **Photos** — timeline + albums  
3. **Trash**  
4. **Settings** — trash prefs, media preview prefs (không chứa account)

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
- Photos: `PhotoThumb` + thumbnail URL / mime fallback; xem original qua presign on demand (spec 07).
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
| Overview (`/`) | Greeting + quota, 2 destination cards (Files/Photos), recent files/photos | me, browser, photos/timeline |
| Files | Browser, search + filter sheet, create folder, upload (progress bar + drag & drop), rename/move/delete file & folder | browser, folders, files, search |
| Photos | Timeline (+ load more), albums CRUD, add/remove items, open/download | photos/* |
| Album detail | Grid + manage | albums/:id, items |
| Trash | Restore / delete forever (optimistic UI) | trash, restore |
| Settings | trash prefs, thumbnail prefs | users/me |
| Profile (`/profile`) | avatar initials, displayName, password, logout | users/me, password |

---

## 9a. Surface mới Phase 2 (spec 09)

### FilterSheet (Files search)
- `BottomSheet` title "Filter"; các nhóm option dạng pill chọn 1: Type (All/Image/Video/Doc/Archive/Folder), Sort (Relevance/Name/Date/Size), Date range (2 input date).
- Nút "Apply" ink full-width; "Reset" ghost bên trái. State filter sống ở query string (`?q=&type=...`) để share/reload giữ được.

### Album cover
- Card album: khung ảnh bìa 1:1 radius-lg phía trên tên; trống → icon photos trên `surface-card`.
- Set/Remove cover nằm trong action sheet item (icon `image` / `restore`), không có nút riêng ngoài grid.

### Share link
- Action sheet file thêm "Share link" (icon `share`) → `ShareSheet`: chọn hạn (segmented Forever/1h/24h/7d) → tạo → hiển thị URL + nút Copy + "Revoke link" danger.
- Settings section "Shared links": list tên file + hạn + revoke (confirm trước khi thu hồi).

### Favorites
- Icon sao: outline khi chưa, fill `warning` (#d97706) khi đã — điểm màu thứ hai duy nhất được phép.
- Files root: segment "Favorites | All" trên list; Overview ưu tiên section Favorites khi có dữ liệu.

### Public share page
- Canvas `surface-soft`, card trắng radius-xl max-width 420px giữa màn: brand mark F, tên file (truncate 2 dòng), meta loại · size, nút Download ink full-width, caption "Shared via Filvault".
- Lỗi (token sai/hết hạn/thu hồi): cùng layout, icon alert, message chung "This link is not available" — không phân biệt nguyên nhân.

### Activity log
- List trong Settings: icon theo type + targetName (1 dòng truncate) + thời gian tương đối muted; "Load more" pattern như timeline.

### Shared with me (S5)
- Entry từ Overview: card thứ 3 trong destinations (icon `users`), **không** thêm tab bottom nav.
- `/shared`: list dòng = icon loại + tên + "by {owner} · {thời gian tương đối}"; file → download qua presign; folder → browse 1 cấp ngay trong trang (nút Back về list). Empty state icon `users` "Nothing shared with you yet".
- Owner share theo email: action sheet file "Share with user" → sheet input email; email lạ vẫn thành công (toast "Invitation sent").

> **Trạng thái (2026-08-23): toàn bộ surface §9a đã implement** — S1 filter, S2 album cover, S3 favorites, S4 share link + public page, S5 shared-with-me + share user nội bộ, S6 activity log. Chi tiết hợp đồng: spec [09](docs/spec/09-phase-2-features.md), tiến độ [10](docs/spec/10-phase-2-status.md).

### Chat (`/chat` — bare surface, spec 11)
- **Surface riêng hoàn toàn**: không dùng shell Filvault (không bottom nav/side nav/header/storage bar). Route có `meta: { bare: true }`; `AppShell` bỏ chrome khi route bare. Truy cập trực tiếp bằng URL `/chat`.
- Layout full-screen kiểu Messenger: rail hội thoại (trái) + thread (giữa) + media panel (phải, ≥1024px).
- Mobile `<768`: master–detail — hiện rail HOẶC thread, không xếp dọc; nút Back (≥44px) quay lại rail.
- Rail: header "Chats" + brand mark F (link về `/`) + input tạo chat mới + list conversation (avatar tròn chữ đầu, title truncate, ngày muted).
- Thread: header avatar + title + Refresh; search pill; bubble ink đậm bo góc lớn (gửi bên phải); composer pill "Aa" + attach `+` + send ink.
- Media panel: ảnh/video đã gửi trong chat (tên + ngày), click mở download — tách khỏi Photos timeline của Vault.
- Touch targets ≥44px; focus ring accent 2px; safe-area top/bottom; motion theo token chung.

---

## 10. Agent Prompt Guide

Khi generate / sửa UI:

```text
Follow DESIGN.md (Filvault). Cal.com-like light utility UI, teal accent #0d9488,
Plus Jakarta Sans, mobile-first bottom nav (Files/Photos/Trash/Settings),
avatar → Profile (not a 5th tab), FAB upload on Files, bottom sheets instead of prompt/confirm,
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
