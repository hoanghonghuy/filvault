<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import Icon from '@/components/AppIcon.vue'
import { useUiStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
import { THEMES, useTheme, type ThemeDef } from '@/lib/theme'

const router = useRouter()
const ui = useUiStore()
const { t } = useI18n()
const { appearanceMode, currentColorTheme, resolvedIsDark, applyColorTheme, setAppearanceMode } = useTheme()

const appearanceModes = [
  { id: 'system' as const, labelKey: 'appearanceModeSystem' },
  { id: 'light' as const, labelKey: 'appearanceModeLight' },
  { id: 'dark' as const, labelKey: 'appearanceModeDark' },
]

const previewTheme = ref<ThemeDef | null>(null)
const previewOpen = ref(false)
const previewDialog = ref<HTMLElement | null>(null)
const previewCloseButton = ref<HTMLButtonElement | null>(null)
let previewOpener: HTMLElement | null = null

const colorThemes = computed(() => THEMES.filter((item) => item.category === 'colors'))
const seasonalThemes = computed(() => THEMES.filter((item) => item.category === 'seasonal'))

function themeName(theme: ThemeDef): string {
  const key = theme.nameKey as keyof typeof t.value
  return (t.value[key] as string) || theme.nameDefault
}

async function openPreview(theme: ThemeDef) {
  previewOpener = document.activeElement instanceof HTMLElement ? document.activeElement : null
  previewTheme.value = theme
  previewOpen.value = true
  await nextTick()
  previewCloseButton.value?.focus()
}

async function closePreview() {
  if (!previewOpen.value) return
  previewOpen.value = false
  await nextTick()
  previewOpener?.focus()
  previewOpener = null
}

function applyTheme(theme: ThemeDef) {
  applyColorTheme(theme.id)
  ui.showToast(`${t.value.themeApplied}: ${themeName(theme)}`, 'success')
  void closePreview()
}

function chooseAppearanceMode(mode: 'system' | 'light' | 'dark') {
  setAppearanceMode(mode)
}

function handlePreviewKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    void closePreview()
    return
  }
  if (event.key !== 'Tab' || !previewDialog.value) return

  const focusable = Array.from(
    previewDialog.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((element) => !element.hasAttribute('hidden') && element.getAttribute('aria-hidden') !== 'true')

  if (focusable.length === 0) {
    event.preventDefault()
    previewDialog.value.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last?.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first?.focus()
  }
}

onBeforeUnmount(() => {
  previewOpener = null
})
</script>

<template>
  <div class="theme-page">
    <header class="theme-header">
      <button type="button" class="btn icon-only back-btn" :aria-label="t.back" @click="router.back()">
        <Icon name="arrow-left" :size="22" />
      </button>
      <h1 class="theme-header-title">{{ t.themeTitle }}</h1>
      <div class="header-spacer" aria-hidden="true"></div>
    </header>

    <section class="card system-dark-card">
      <h2 class="section-title">{{ t.appearanceModeLabel }}</h2>
      <div class="appearance-mode-row" role="radiogroup" :aria-label="t.appearanceModeLabel">
        <button
          v-for="mode in appearanceModes"
          :key="mode.id"
          type="button"
          class="appearance-mode-btn"
          :class="{ active: appearanceMode === mode.id }"
          role="radio"
          :aria-checked="appearanceMode === mode.id"
          @click="chooseAppearanceMode(mode.id)"
        >
          {{ (t as any)[mode.labelKey] }}
        </button>
      </div>
      <p class="appearance-mode-hint">
        {{ appearanceMode === 'system' ? t.appearanceModeSystemHint : resolvedIsDark ? t.appearanceModeDarkHint : t.appearanceModeLightHint }}
      </p>
      <p class="appearance-mode-hint color-theme-hint">{{ t.colorThemeAppearanceHint }}</p>
    </section>

    <section class="card theme-section">
      <h2 class="section-title">{{ t.themeColorPalette }}</h2>
      <div class="swatches-grid">
        <button
          v-for="theme in colorThemes"
          :key="theme.id"
          type="button"
          class="swatch-item"
          :class="{ active: currentColorTheme === theme.id }"
          :aria-pressed="currentColorTheme === theme.id"
          @click="openPreview(theme)"
        >
          <span class="swatch-box" :style="{ background: theme.swatchGradient }">
            <span v-if="currentColorTheme === theme.id" class="swatch-check" aria-hidden="true"><Icon name="check" :size="14" /></span>
          </span>
          <span class="swatch-label">{{ themeName(theme) }}</span>
        </button>
      </div>
    </section>

    <section class="card theme-section">
      <h2 class="section-title">{{ t.themeSeasonal }}</h2>
      <div class="seasonal-grid">
        <button
          v-for="theme in seasonalThemes"
          :key="theme.id"
          type="button"
          class="seasonal-card"
          :class="{ active: currentColorTheme === theme.id }"
          :aria-pressed="currentColorTheme === theme.id"
          @click="openPreview(theme)"
        >
          <span class="seasonal-swatch-box" :style="{ background: theme.swatchGradient }">
            <span v-if="currentColorTheme === theme.id" class="swatch-check" aria-hidden="true"><Icon name="check" :size="14" /></span>
          </span>
          <span class="swatch-label">{{ themeName(theme) }}</span>
        </button>
      </div>
    </section>

    <Teleport to="body">
      <Transition name="preview-fade">
        <div
          v-if="previewOpen && previewTheme"
          ref="previewDialog"
          class="preview-overlay"
          role="dialog"
          aria-modal="true"
          aria-labelledby="theme-preview-title"
          tabindex="-1"
          @keydown="handlePreviewKeydown"
        >
          <header class="preview-top-bar">
            <button ref="previewCloseButton" type="button" class="btn icon-only back-btn" :aria-label="t.back" @click="closePreview">
              <Icon name="arrow-left" :size="22" />
            </button>
            <h2 id="theme-preview-title" class="preview-header-title">{{ themeName(previewTheme) }}</h2>
            <div class="header-spacer" aria-hidden="true"></div>
          </header>

          <div class="preview-content-area">
            <div class="storage-preview" :style="{ background: previewTheme.previewBg, '--preview-accent': previewTheme.todayAccent }">
              <div class="storage-preview-shell">
                <div class="storage-preview-brand">
                  <span class="storage-preview-logo" aria-hidden="true"><Icon name="folder" :size="18" /></span>
                  <strong>Filvault</strong>
                  <span class="storage-preview-more" aria-hidden="true"><Icon name="more" :size="18" /></span>
                </div>
                <div class="storage-preview-search" aria-hidden="true">
                  <Icon name="search" :size="15" />
                  <span>{{ t.files }}</span>
                </div>
              </div>

              <div class="storage-preview-body">
                <div class="storage-preview-heading">
                  <strong>{{ t.files }}</strong>
                  <span class="storage-preview-chip">{{ themeName(previewTheme) }}</span>
                </div>

                <div class="storage-preview-grid" aria-hidden="true">
                  <div class="storage-preview-folder selected-preview-item">
                    <span class="preview-item-icon"><Icon name="folder" :size="20" /></span>
                    <span>{{ t.folders }}</span>
                  </div>
                  <div class="storage-preview-media">
                    <span class="preview-photo-shape"></span>
                    <span class="preview-media-line"></span>
                  </div>
                </div>

                <div class="storage-preview-file" aria-hidden="true">
                  <span class="preview-file-icon"><Icon name="file" :size="19" /></span>
                  <span class="preview-file-copy">
                    <span class="preview-file-title">Filvault.pdf</span>
                    <span class="preview-file-meta">2.4 MB</span>
                  </span>
                  <Icon name="more" :size="18" />
                </div>

                <div class="storage-preview-progress" aria-hidden="true">
                  <span class="preview-progress-copy"><span>{{ t.files }}</span><span>68%</span></span>
                  <span class="preview-progress-track"><span class="preview-progress-value"></span></span>
                </div>
              </div>

              <div class="storage-preview-nav" aria-hidden="true">
                <span class="preview-nav-item active-preview-nav"><Icon name="folder" :size="18" /></span>
                <span class="preview-nav-item"><Icon name="image" :size="18" /></span>
                <span class="preview-nav-item"><Icon name="settings" :size="18" /></span>
              </div>
            </div>
          </div>

          <footer class="preview-footer-cta">
            <button type="button" class="btn cta-apply-btn" :style="{ background: previewTheme.accentColor }" @click="applyTheme(previewTheme)">
              {{ currentColorTheme === previewTheme.id ? t.themeInUse : t.themeApply }}
            </button>
          </footer>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.theme-page { padding-bottom: var(--space-xl); }
.theme-header { display:grid; grid-template-columns:44px minmax(0,1fr) 44px; align-items:center; gap:var(--space-sm); margin-bottom:var(--space-md); padding:var(--space-xs) 0; }
.back-btn { color:var(--ink); padding:8px; min-width:44px; min-height:44px; }
.header-spacer { width:44px; height:44px; }
.theme-header-title,.preview-header-title { margin:0; min-width:0; color:var(--ink); font-size:1.125rem; font-weight:700; line-height:1.3; text-align:center; overflow-wrap:anywhere; }
.system-dark-card,.theme-section { margin-bottom:var(--space-md); padding:var(--space-md); background:var(--surface); border:1px solid var(--hairline); border-radius:var(--radius-xl); }
.section-title { margin:0 0 var(--space-md); font-size:1rem; font-weight:700; color:var(--ink); }
.appearance-mode-row { display:flex; gap:var(--space-xs); margin-top:var(--space-sm); }
.appearance-mode-btn { flex:1; min-width:0; min-height:44px; border:1px solid var(--hairline); background:var(--surface-soft); color:var(--ink); border-radius:var(--radius-lg); padding:10px 8px; font-size:.875rem; font-weight:600; cursor:pointer; }
.appearance-mode-btn.active { background:var(--accent-soft); border-color:var(--accent); color:var(--accent); }
.appearance-mode-hint { margin:var(--space-sm) 0 0; font-size:.8125rem; color:var(--muted); line-height:1.45; }
.color-theme-hint { margin-top:var(--space-xs); }
.swatches-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:var(--space-sm); }
.swatch-item,.seasonal-card { display:flex; flex-direction:column; align-items:center; gap:6px; min-width:0; min-height:44px; border:0; background:none; color:var(--ink); cursor:pointer; padding:4px; border-radius:var(--radius-lg); }
.swatch-item:focus-visible,.seasonal-card:focus-visible { outline:3px solid var(--accent); outline-offset:2px; }
.swatch-box { position:relative; width:100%; aspect-ratio:1; max-width:68px; border-radius:16px; box-shadow:0 2px 6px rgba(0,0,0,.08); }
.swatch-check { position:absolute; top:6px; right:6px; width:20px; height:20px; border-radius:50%; background:#fff; color:#111827; display:flex; align-items:center; justify-content:center; box-shadow:0 1px 3px rgba(0,0,0,.2); }
.swatch-label { font-size:.75rem; font-weight:500; color:var(--ink); text-align:center; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; max-width:100%; }
.seasonal-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:var(--space-sm); }
.seasonal-swatch-box { position:relative; width:100%; height:60px; border-radius:16px; box-shadow:0 2px 6px rgba(0,0,0,.08); }

.preview-overlay { position:fixed; inset:0; z-index:1000; background:var(--canvas); display:flex; flex-direction:column; overflow-y:auto; padding:max(env(safe-area-inset-top),var(--space-xs)) max(env(safe-area-inset-right),var(--space-md)) max(env(safe-area-inset-bottom),var(--space-md)) max(env(safe-area-inset-left),var(--space-md)); }
.preview-top-bar { display:grid; grid-template-columns:44px minmax(0,1fr) 44px; align-items:center; gap:var(--space-sm); min-height:56px; flex-shrink:0; }
.preview-content-area { flex:1; display:flex; align-items:center; justify-content:center; padding:var(--space-sm) 0; }
.storage-preview { --preview-accent:var(--accent); width:min(100%,720px); min-height:430px; border:1px solid rgba(17,24,39,.13); border-radius:28px; box-shadow:0 18px 50px rgba(0,0,0,.14); display:grid; grid-template-rows:auto 1fr auto; overflow:hidden; color:#111827; }
.storage-preview-shell { padding:16px; border-bottom:1px solid rgba(17,24,39,.09); background:rgba(255,255,255,.78); backdrop-filter:blur(10px); }
.storage-preview-brand { display:grid; grid-template-columns:36px minmax(0,1fr) 32px; align-items:center; gap:10px; }
.storage-preview-logo { width:36px; height:36px; border-radius:11px; display:flex; align-items:center; justify-content:center; background:var(--preview-accent); color:#fff; }
.storage-preview-more { display:flex; justify-content:center; color:#64748b; }
.storage-preview-search { margin-top:12px; min-height:38px; border-radius:13px; background:rgba(255,255,255,.82); display:flex; align-items:center; gap:8px; padding:0 12px; color:#64748b; font-size:.8rem; }
.storage-preview-body { padding:18px 16px; }
.storage-preview-heading,.preview-progress-copy { display:flex; align-items:center; justify-content:space-between; gap:12px; }
.storage-preview-chip { max-width:50%; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; padding:5px 9px; border-radius:999px; background:var(--preview-accent); color:#fff; font-size:.7rem; font-weight:700; }
.storage-preview-grid { display:grid; grid-template-columns:1.15fr .85fr; gap:12px; margin-top:16px; }
.storage-preview-folder,.storage-preview-media { min-height:110px; border-radius:18px; background:rgba(255,255,255,.82); border:1px solid rgba(17,24,39,.08); padding:14px; }
.storage-preview-folder { display:flex; flex-direction:column; justify-content:space-between; font-size:.8rem; font-weight:700; }
.selected-preview-item { outline:3px solid color-mix(in srgb,var(--preview-accent) 72%,white); outline-offset:-3px; }
.preview-item-icon,.preview-file-icon { color:var(--preview-accent); }
.storage-preview-media { display:flex; flex-direction:column; justify-content:flex-end; gap:10px; background:linear-gradient(145deg,color-mix(in srgb,var(--preview-accent) 24%,white),rgba(255,255,255,.9)); }
.preview-photo-shape { width:54px; height:42px; border-radius:13px; background:var(--preview-accent); opacity:.78; }
.preview-media-line { width:70%; height:8px; border-radius:999px; background:rgba(17,24,39,.16); }
.storage-preview-file { margin-top:12px; min-height:62px; display:grid; grid-template-columns:36px minmax(0,1fr) 24px; align-items:center; gap:10px; padding:10px 12px; border-radius:16px; background:rgba(255,255,255,.82); border:1px solid rgba(17,24,39,.08); }
.preview-file-copy { min-width:0; display:flex; flex-direction:column; gap:2px; }
.preview-file-title { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-size:.8rem; font-weight:700; }
.preview-file-meta { color:#64748b; font-size:.68rem; }
.storage-preview-progress { margin-top:14px; padding:12px; border-radius:16px; background:rgba(255,255,255,.7); font-size:.72rem; font-weight:700; }
.preview-progress-track { display:block; height:8px; margin-top:8px; border-radius:999px; overflow:hidden; background:rgba(17,24,39,.1); }
.preview-progress-value { display:block; width:68%; height:100%; background:var(--preview-accent); border-radius:inherit; }
.storage-preview-nav { min-height:56px; display:flex; align-items:center; justify-content:space-around; border-top:1px solid rgba(17,24,39,.09); background:rgba(255,255,255,.82); }
.preview-nav-item { width:40px; height:34px; border-radius:12px; display:flex; align-items:center; justify-content:center; color:#64748b; }
.active-preview-nav { background:var(--preview-accent); color:#fff; }
.preview-footer-cta { padding:var(--space-sm) 0 0; flex-shrink:0; }
.cta-apply-btn { width:100%; min-height:48px; border-radius:var(--radius-pill); font-size:1rem; font-weight:700; color:#fff; border:0; cursor:pointer; box-shadow:0 4px 14px rgba(0,0,0,.15); }
.preview-fade-enter-active,.preview-fade-leave-active { transition:opacity var(--duration-medium) var(--ease-standard),transform var(--duration-medium) var(--ease-standard); }
.preview-fade-enter-from,.preview-fade-leave-to { opacity:0; transform:translateY(20px); }

@media (min-width:600px) { .seasonal-grid { grid-template-columns:repeat(4,minmax(0,1fr)); } }
@media (min-width:768px) { .storage-preview { grid-template-columns:190px 1fr; grid-template-rows:1fr; min-height:440px; } .storage-preview-shell { border-bottom:0; border-right:1px solid rgba(17,24,39,.09); } .storage-preview-brand { grid-template-columns:36px minmax(0,1fr); } .storage-preview-more { display:none; } .storage-preview-search { margin-top:24px; } .storage-preview-body { padding:26px; } .storage-preview-nav { flex-direction:column; justify-content:flex-start; gap:12px; padding-top:110px; border-top:0; border-right:1px solid rgba(17,24,39,.09); grid-column:1; grid-row:1; pointer-events:none; background:transparent; } }
@media (prefers-reduced-motion:reduce) { .preview-fade-enter-active,.preview-fade-leave-active { transition:none; } .preview-fade-enter-from,.preview-fade-leave-to { transform:none; } }
</style>
