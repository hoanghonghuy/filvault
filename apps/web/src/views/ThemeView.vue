<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Icon from '@/components/AppIcon.vue'
import { useUiStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
import { THEMES, useTheme, type ThemeDef } from '@/lib/theme'

const router = useRouter()
const ui = useUiStore()
const { t } = useI18n()
const { currentColorTheme, followSystemDark, applyColorTheme, setFollowSystemDark } = useTheme()

const activeTab = ref<'theme' | 'icons' | 'display'>('theme')
const previewTheme = ref<ThemeDef | null>(null)
const previewOpen = ref(false)

const colorThemes = computed(() => THEMES.filter((item) => item.category === 'colors'))
const seasonalThemes = computed(() => THEMES.filter((item) => item.category === 'seasonal'))

function openPreview(theme: ThemeDef) {
  previewTheme.value = theme
  previewOpen.value = true
}

function closePreview() {
  previewOpen.value = false
}

function applyTheme(theme: ThemeDef) {
  applyColorTheme(theme.id)
  ui.showToast(`${t.value.themeApplied}: ${themeName(theme)}`, 'success')
  closePreview()
}

function themeName(theme: ThemeDef): string {
  const key = theme.nameKey as keyof typeof t.value
  return (t.value[key] as string) || theme.nameDefault
}

function handleFollowSystemToggle() {
  setFollowSystemDark(!followSystemDark.value)
}
</script>

<template>
  <div class="theme-page">
    <!-- Top Navigation Header -->
    <header class="theme-header">
      <button type="button" class="btn icon-only back-btn" :aria-label="t.back" @click="router.back()">
        <Icon name="arrow-left" :size="22" />
      </button>

      <!-- Segmented Top Tabs -->
      <div class="header-tabs" role="tablist">
        <button
          type="button"
          role="tab"
          class="header-tab-pill"
          :class="{ active: activeTab === 'theme' }"
          :aria-selected="activeTab === 'theme'"
          @click="activeTab = 'theme'"
        >
          {{ t.themeTabThemes }}
        </button>
        <button
          type="button"
          role="tab"
          class="header-tab-pill"
          :class="{ active: activeTab === 'icons' }"
          :aria-selected="activeTab === 'icons'"
          @click="activeTab = 'icons'"
        >
          {{ t.themeTabIcons }}
        </button>
        <button
          type="button"
          role="tab"
          class="header-tab-pill"
          :class="{ active: activeTab === 'display' }"
          :aria-selected="activeTab === 'display'"
          @click="activeTab = 'display'"
        >
          {{ t.themeTabDisplay }}
        </button>
      </div>

      <div class="header-spacer"></div>
    </header>

    <!-- Follow System Dark Mode Card -->
    <section class="card system-dark-card">
      <label class="toggle-row">
        <span class="toggle-label">{{ t.themeFollowSystem }}</span>
        <div class="switch-wrapper">
          <input
            :checked="followSystemDark"
            type="checkbox"
            class="switch-input"
            @change="handleFollowSystemToggle"
          />
          <span class="switch-track" aria-hidden="true"></span>
        </div>
      </label>
    </section>

    <!-- Section 1: Color Palette (Dòng màu sắc) -->
    <section class="card theme-section">
      <h2 class="section-title">{{ t.themeColorPalette }}</h2>
      <div class="swatches-grid">
        <button
          v-for="theme in colorThemes"
          :key="theme.id"
          type="button"
          class="swatch-item"
          :class="{ active: currentColorTheme === theme.id }"
          @click="openPreview(theme)"
        >
          <div class="swatch-box" :style="{ background: theme.swatchGradient }">
            <span v-if="currentColorTheme === theme.id" class="swatch-check" aria-hidden="true">
              <Icon name="check" :size="14" />
            </span>
            <span v-if="theme.isPro" class="swatch-pro-badge" aria-hidden="true">
              <Icon name="crown" :size="11" />
            </span>
          </div>
          <span class="swatch-label">{{ themeName(theme) }}</span>
        </button>
      </div>
    </section>

    <!-- Section 2: Seasonal Series (Loạt Mùa) -->
    <section class="card theme-section">
      <h2 class="section-title">{{ t.themeSeasonal }}</h2>
      <div class="seasonal-grid">
        <button
          v-for="theme in seasonalThemes"
          :key="theme.id"
          type="button"
          class="seasonal-card"
          :class="{ active: currentColorTheme === theme.id }"
          @click="openPreview(theme)"
        >
          <div class="seasonal-swatch-box" :style="{ background: theme.swatchGradient }">
            <span class="seasonal-crown-badge" aria-hidden="true">
              <Icon name="crown" :size="12" />
            </span>
            <span v-if="currentColorTheme === theme.id" class="swatch-check" aria-hidden="true">
              <Icon name="check" :size="14" />
            </span>
          </div>
          <span class="swatch-label">{{ themeName(theme) }}</span>
        </button>
      </div>
    </section>

    <!-- Theme Live Preview Modal (Matching Photo 1) -->
    <Teleport to="body">
      <Transition name="preview-fade">
        <div v-if="previewOpen && previewTheme" class="preview-overlay" role="dialog" aria-modal="true">
          <!-- Preview Header -->
          <header class="preview-top-bar">
            <button type="button" class="btn icon-only back-btn" :aria-label="t.back" @click="closePreview">
              <Icon name="arrow-left" :size="22" />
            </button>
            <h1 class="preview-header-title">{{ themeName(previewTheme) }}</h1>
            <div class="header-spacer"></div>
          </header>

          <!-- Preview Mockup Card Frame -->
          <div class="preview-content-area">
            <div class="mockup-frame" :style="{ background: previewTheme.previewBg }">
              <!-- Mockup Calendar / Header -->
              <div class="mockup-header-row">
                <span class="mockup-month">tháng 9</span>
                <span class="mockup-dots" aria-hidden="true">
                  <Icon name="more" :size="18" />
                </span>
              </div>

              <!-- Mini Day Names Matrix -->
              <div class="mockup-week-row">
                <span>Th 2</span>
                <span>Th 3</span>
                <span>Th 4</span>
                <span>Th 5</span>
                <span>Th 6</span>
                <span>Th 7</span>
                <span>CN</span>
              </div>

              <!-- Mini Dates Grid -->
              <div class="mockup-dates-grid">
                <span class="faded">31</span>
                <span>1</span>
                <span>2</span>
                <span>3</span>
                <span>4</span>
                <span>5</span>
                <span>6</span>
                <span>7</span>
                <span class="active-date-circle" :style="{ background: previewTheme.todayAccent, color: '#fff' }">
                  8
                </span>
                <span>9</span>
                <span>10</span>
                <span class="highlight-date">11</span>
                <span>12</span>
                <span>13</span>
                <span>14</span>
                <span>15</span>
                <span>16</span>
                <span>17</span>
                <span>18</span>
                <span>19</span>
                <span>20</span>
                <span>21</span>
                <span>22</span>
                <span>23</span>
                <span>24</span>
                <span>25</span>
                <span>26</span>
                <span>27</span>
              </div>

              <!-- Mockup Today Checklist Card -->
              <div class="mockup-today-card">
                <h3 class="mockup-today-title">{{ t.themePreviewToday }}</h3>
                <div class="mockup-task-item">
                  <span class="mockup-checkbox"></span>
                  <span class="mockup-task-text">{{ t.themePreviewTask1 }}</span>
                </div>
                <div class="mockup-task-item">
                  <span class="mockup-checkbox"></span>
                  <span class="mockup-task-text">{{ t.themePreviewTask2 }}</span>
                </div>
                <div class="mockup-task-item">
                  <span class="mockup-checkbox"></span>
                  <span class="mockup-task-text">{{ t.themePreviewTask3 }}</span>
                </div>
              </div>

              <!-- Mockup FAB Button -->
              <div class="mockup-fab" :style="{ background: previewTheme.todayAccent }">
                <Icon name="plus" :size="20" />
              </div>

              <!-- Mockup Bottom Bar -->
              <div class="mockup-bottom-nav">
                <span class="nav-dot">
                  <Icon name="check" :size="16" />
                </span>
                <span class="nav-dot active-pill" :style="{ background: previewTheme.todayAccent, color: '#fff' }">
                  8
                </span>
                <span class="nav-dot">
                  <Icon name="home" :size="16" />
                </span>
                <span class="nav-dot">
                  <Icon name="folder" :size="16" />
                </span>
                <span class="nav-dot">
                  <Icon name="settings" :size="16" />
                </span>
              </div>
            </div>
          </div>

          <!-- Bottom Action CTA (Sử dụng) -->
          <footer class="preview-footer-cta">
            <button
              type="button"
              class="btn cta-apply-btn"
              :style="{ background: previewTheme.accentColor }"
              @click="applyTheme(previewTheme)"
            >
              {{ currentColorTheme === previewTheme.id ? t.themeInUse : t.themeApply }}
            </button>
          </footer>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.theme-page {
  padding-bottom: var(--space-xl);
}

.theme-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-md);
  padding: var(--space-xs) 0;
}

.back-btn {
  color: var(--ink);
  padding: 8px;
}

.header-spacer {
  width: 36px;
}

.header-tabs {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.header-tab-pill {
  border: none;
  background: none;
  font-size: 1rem;
  font-weight: 600;
  color: var(--muted);
  padding: 6px 4px;
  cursor: pointer;
  position: relative;
  transition: color var(--duration-short) var(--ease-standard);
}

.header-tab-pill.active {
  color: var(--ink);
  font-weight: 700;
}

.header-tab-pill.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 20%;
  right: 20%;
  height: 3px;
  border-radius: var(--radius-pill);
  background: var(--accent);
}

/* System Dark Card */
.system-dark-card {
  margin-bottom: var(--space-md);
  padding: var(--space-md);
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-xl);
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  min-height: 36px;
}

.toggle-label {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--ink);
}

.switch-wrapper {
  position: relative;
  width: 44px;
  height: 24px;
}

.switch-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-track {
  position: relative;
  display: block;
  width: 44px;
  height: 24px;
  border-radius: var(--radius-pill);
  background: var(--hairline);
  transition: background-color var(--duration-short) var(--ease-standard);
}

.switch-track::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--canvas);
  box-shadow: 0 1px 3px rgba(17, 24, 39, 0.25);
  transition: transform var(--duration-short) var(--ease-standard);
}

.switch-input:checked + .switch-track {
  background: var(--accent);
}

.switch-input:checked + .switch-track::after {
  transform: translateX(20px);
}

/* Theme Sections */
.theme-section {
  margin-bottom: var(--space-md);
  padding: var(--space-md);
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-xl);
}

.section-title {
  margin: 0 0 var(--space-md);
  font-size: 1rem;
  font-weight: 700;
  color: var(--ink);
}

/* Swatches Grid */
.swatches-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-sm);
}

.swatch-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  border: none;
  background: none;
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-lg);
  transition: transform var(--duration-short) var(--ease-standard);
}

.swatch-item:active {
  transform: scale(0.95);
}

.swatch-box {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  max-width: 68px;
  border-radius: 16px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.swatch-check {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #ffffff;
  color: #111827;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.swatch-pro-badge {
  position: absolute;
  bottom: 6px;
  right: 6px;
  padding: 2px 5px;
  border-radius: var(--radius-pill);
  background: #f59e0b;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
}

.swatch-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--ink);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 72px;
}

/* Seasonal Grid */
.seasonal-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

@media (min-width: 600px) {
  .seasonal-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

.seasonal-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  border: none;
  background: none;
  cursor: pointer;
  padding: 4px;
}

.seasonal-swatch-box {
  position: relative;
  width: 100%;
  height: 60px;
  border-radius: 16px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.seasonal-crown-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #f59e0b;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Full Preview Overlay (Matching Photo 1) */
.preview-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: var(--canvas);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  padding: 0 var(--space-md) var(--space-md);
}

.preview-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  flex-shrink: 0;
}

.preview-header-title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--ink);
}

.preview-content-area {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-sm) 0;
}

.mockup-frame {
  width: 100%;
  max-width: 330px;
  border-radius: 28px;
  border: 4px solid var(--hairline);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
  padding: 20px 16px 16px;
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 480px;
}

.mockup-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.mockup-month {
  font-size: 1.125rem;
  font-weight: 700;
  color: #111827;
}

.mockup-dots {
  color: #6b7280;
}

.mockup-week-row {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  font-size: 0.6875rem;
  font-weight: 600;
  color: #9ca3af;
  margin-bottom: 8px;
}

.mockup-dates-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  font-size: 0.75rem;
  row-gap: 8px;
  color: #374151;
  margin-bottom: 16px;
}

.mockup-dates-grid span {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 24px;
}

.mockup-dates-grid .faded {
  color: #d1d5db;
}

.mockup-dates-grid .highlight-date {
  color: #d97706;
  font-weight: 700;
}

.active-date-circle {
  width: 24px;
  height: 24px;
  margin: 0 auto;
  border-radius: 50%;
  font-weight: 700;
}

.mockup-today-card {
  background: #ffffff;
  border-radius: 18px;
  padding: 14px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.04);
  margin-bottom: auto;
}

.mockup-today-title {
  margin: 0 0 10px;
  font-size: 0.875rem;
  font-weight: 700;
  color: #111827;
}

.mockup-task-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.mockup-task-item:last-child {
  margin-bottom: 0;
}

.mockup-checkbox {
  width: 15px;
  height: 15px;
  border-radius: 4px;
  border: 1.5px solid #d1d5db;
  flex-shrink: 0;
}

.mockup-task-text {
  font-size: 0.75rem;
  color: #374151;
  line-height: 1.3;
}

.mockup-fab {
  position: absolute;
  right: 20px;
  bottom: 60px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.18);
}

.mockup-bottom-nav {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding-top: 10px;
  margin-top: 14px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.nav-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.nav-dot.active-pill {
  width: 32px;
  height: 24px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 700;
}

/* CTA Apply Button */
.preview-footer-cta {
  padding: var(--space-sm) 0 var(--space-xs);
  flex-shrink: 0;
}

.cta-apply-btn {
  width: 100%;
  min-height: 48px;
  border-radius: var(--radius-pill);
  font-size: 1rem;
  font-weight: 700;
  color: #ffffff;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
  transition: transform var(--duration-short) var(--ease-standard);
}

.cta-apply-btn:active {
  transform: scale(0.98);
}

.preview-fade-enter-active,
.preview-fade-leave-active {
  transition: opacity var(--duration-medium) var(--ease-standard), transform var(--duration-medium) var(--ease-standard);
}

.preview-fade-enter-from,
.preview-fade-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
