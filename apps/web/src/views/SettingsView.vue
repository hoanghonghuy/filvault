<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import Icon from '@/components/AppIcon.vue'
import { userInitials } from '@/lib/userInitials'
import type { ActivityEvent, ActivityEventType, ShareLinkInfo, User } from '@/api/types'
import { setLocale, useI18n, type Locale } from '@/lib/i18n'
import { THEMES, useTheme, type ThemeDef } from '@/lib/theme'
import { useStorageUsage } from '@/lib/useStorageUsage'

const auth = useAuthStore()
const ui = useUiStore()
const { locale, t } = useI18n()
const { appearanceMode, currentColorTheme, resolvedIsDark, applyColorTheme, setAppearanceMode } = useTheme()
const fallbackTheme = THEMES[0] as ThemeDef
const currentThemeDef = computed<ThemeDef>(
  () => THEMES.find((item) => item.id === currentColorTheme.value) ?? fallbackTheme,
)

const appearanceModes = [
  { id: 'system' as const, labelKey: 'appearanceModeSystem' },
  { id: 'light' as const, labelKey: 'appearanceModeLight' },
  { id: 'dark' as const, labelKey: 'appearanceModeDark' },
]

function chooseAppearanceMode(mode: 'system' | 'light' | 'dark') {
  setAppearanceMode(mode)
}

const storage = useStorageUsage(formatBytes)
const avatarInitials = computed(() =>
  userInitials(auth.user?.displayName ?? '', auth.user?.email ?? ''),
)

const imageThumbnailsEnabled = ref(auth.user?.imageThumbnailsEnabled ?? true)
const videoThumbnailsEnabled = ref(auth.user?.videoThumbnailsEnabled ?? true)
const trashAutoDeleteEnabled = ref(auth.user?.trashAutoDeleteEnabled ?? false)
const trashRetentionDays = ref(auth.user?.trashRetentionDays ?? 30)
const error = ref('')
const savingSettings = ref(false)
// Persisted in localStorage under filvault.locale
function chooseLocale(next: Locale) {
  setLocale(next)
}

async function saveSettings() {
  error.value = ''
  savingSettings.value = true
  try {
    await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({
        imageThumbnailsEnabled: imageThumbnailsEnabled.value,
        videoThumbnailsEnabled: videoThumbnailsEnabled.value,
        trashAutoDeleteEnabled: trashAutoDeleteEnabled.value,
        trashRetentionDays: Number(trashRetentionDays.value),
      }),
    })
    await auth.loadMe()
    ui.showToast('Settings saved')
  } catch (e) {
    error.value = formatApiError(e, 'Save failed')
  } finally {
    savingSettings.value = false
  }
}

async function logout() {
  const ok = await ui.confirm({
    title: 'Log out?',
    message: 'You will need to sign in again to access your files.',
    confirmLabel: 'Log out',
  })
  if (!ok) return
  await auth.logout()
  window.location.href = '/login'
}

const shareLinks = ref<ShareLinkInfo[]>([])
const linksLoading = ref(false)

async function loadShareLinks() {
  linksLoading.value = true
  try {
    const out = await api<{ links: ShareLinkInfo[] }>('/share-links')
    shareLinks.value = out.links
  } catch {
    shareLinks.value = []
  } finally {
    linksLoading.value = false
  }
}

async function copyLink(link: ShareLinkInfo) {
  try {
    await navigator.clipboard.writeText(window.location.origin + (link.url ?? ''))
    ui.showToast('Link copied')
  } catch {
    ui.showToast('Copy failed', 'info')
  }
}

async function revokeFromSettings(link: ShareLinkInfo) {
  const ok = await ui.confirm({
    title: 'Revoke link?',
    message: `"${link.fileName}" will no longer be shared publicly.`,
    confirmLabel: 'Revoke link',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/files/${link.fileId}/share`, { method: 'DELETE' })
    ui.showToast('Link revoked')
    await loadShareLinks()
  } catch (e) {
    error.value = formatApiError(e, 'Could not revoke link')
  }
}

const activityEvents = ref<ActivityEvent[]>([])
const activityNextBefore = ref<string | undefined>(undefined)
const activityLoading = ref(false)
const activityLoadingMore = ref(false)
const activityError = ref(false)

const ACTIVITY_PAGE_SIZE = 20

async function loadActivity() {
  activityLoading.value = true
  activityError.value = false
  try {
    const page = await api<{ events: ActivityEvent[]; nextBefore?: string }>(
      `/activity?limit=${ACTIVITY_PAGE_SIZE}`,
    )
    activityEvents.value = page.events
    activityNextBefore.value = page.nextBefore
  } catch {
    activityError.value = true
  } finally {
    activityLoading.value = false
  }
}

async function loadMoreActivity() {
  if (!activityNextBefore.value || activityLoadingMore.value) return
  activityLoadingMore.value = true
  try {
    const page = await api<{ events: ActivityEvent[]; nextBefore?: string }>(
      `/activity?limit=${ACTIVITY_PAGE_SIZE}&before=${encodeURIComponent(activityNextBefore.value)}`,
    )
    activityEvents.value.push(...page.events)
    activityNextBefore.value = page.nextBefore
  } catch {
    ui.showToast('Could not load more activity', 'info')
  } finally {
    activityLoadingMore.value = false
  }
}

function activityIcon(type: ActivityEventType): string {
  switch (type) {
    case 'file.uploaded':
      return 'upload'
    case 'file.trashed':
    case 'file.purged':
    case 'folder.trashed':
      return 'trash'
    case 'file.restored':
    case 'folder.restored':
      return 'restore'
    case 'share.created':
    case 'share.revoked':
      return 'share'
    default:
      return 'settings'
  }
}

function activityLabel(type: ActivityEventType): string {
  switch (type) {
    case 'file.uploaded':
      return 'Uploaded'
    case 'file.trashed':
      return 'Moved to trash'
    case 'file.restored':
      return 'Restored'
    case 'file.purged':
      return 'Permanently deleted'
    case 'folder.trashed':
      return 'Folder moved to trash'
    case 'folder.restored':
      return 'Folder restored'
    case 'share.created':
      return 'Share created'
    case 'share.revoked':
      return 'Share revoked'
    case 'password.changed':
      return 'Password changed'
    case 'settings.changed':
      return 'Settings updated'
    default:
      return type
  }
}

function relativeTime(iso: string): string {
  const seconds = Math.round((Date.now() - new Date(iso).getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.round(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.round(hours / 24)
  if (days < 30) return `${days}d ago`
  return new Date(iso).toLocaleDateString()
}

onMounted(() => {
  void storage.reload()
  loadShareLinks()
  loadActivity()
})
</script>

<template>
  <div class="settings-page">
    <h1 class="page-title desktop-only">{{ t.settingsTitle }}</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <!-- Profile Header Card (TeraBox style) -->
    <section class="card profile-header-card">
      <div class="profile-header-main">
        <div class="profile-avatar-box">
          <img
            v-if="auth.user?.avatarUrl"
            :src="auth.user.avatarUrl"
            :alt="auth.user.displayName || 'Avatar'"
            class="profile-avatar-img"
          />
          <span v-else class="profile-avatar-initials">{{ avatarInitials }}</span>
        </div>
        <div class="profile-meta">
          <div class="profile-name-row">
            <h2 class="profile-name">{{ auth.user?.displayName || auth.user?.email || 'Người dùng Filvault' }}</h2>
          </div>
          <p class="profile-email muted">{{ auth.user?.email }}</p>
        </div>
        <RouterLink to="/profile" class="profile-arrow-link" :title="t.profile" :aria-label="t.profile">
          <Icon name="arrow-right" :size="18" />
        </RouterLink>
      </div>
    </section>

    <!-- Storage Card & Shortcuts (TeraBox style "Đám mây của tôi") -->
    <section class="card storage-card">
      <div class="storage-card-header">
        <div class="storage-title-wrap">
          <Icon name="cloud" :size="20" class="storage-cloud-icon" />
          <span class="storage-title">{{ t.myCloud }}</span>
        </div>
        <span v-if="storage.state.value === 'loading'" class="storage-stats-text muted">
          {{ t.storageLoading }}
        </span>
        <span v-else-if="storage.state.value === 'usage-unavailable'" class="storage-stats-text muted">
          {{ t.storageUnavailable }}
        </span>
        <span v-else class="storage-stats-text">
          <strong>{{ formatBytes(storage.usedBytes.value) }}</strong> / {{ formatBytes(storage.quotaBytes.value) }}
        </span>
      </div>

      <template v-if="storage.state.value === 'usage-unavailable'">
        <p class="storage-status muted">{{ t.storageUnavailableHint }}</p>
        <button type="button" class="btn storage-retry-btn" @click="storage.reload()">
          {{ t.retry }}
        </button>
      </template>
      <template v-else-if="storage.state.value !== 'loading'">
        <div
          class="storage-bar-track"
          role="progressbar"
          :aria-valuenow="storage.progressAria.value.valuenow"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-label="storage.progressAria.value.label"
        >
          <div
            class="storage-bar-fill"
            :class="storage.toneClass.value"
            :style="{ width: `${storage.percent.value}%` }"
          />
        </div>
        <p
          v-if="storage.state.value === 'near-quota'"
          class="storage-status warning"
          role="status"
        >
          <Icon name="alert" :size="14" aria-hidden="true" />
          <span>{{ t.storageNearQuota }} — {{ t.storageNearQuotaHint }}</span>
        </p>
        <div
          v-else-if="storage.state.value === 'full-quota'"
          class="storage-status full"
          role="status"
        >
          <Icon name="alert" :size="14" aria-hidden="true" />
          <div class="storage-status-copy">
            <strong>{{ t.storageFullQuota }}</strong>
            <span class="muted">{{ t.storageFullQuotaHint }}</span>
          </div>
          <RouterLink to="/trash" class="storage-action-link">{{ t.storageFreeUp }}</RouterLink>
        </div>
      </template>

      <div class="storage-shortcuts-grid">
        <RouterLink to="/files" class="shortcut-item">
          <div class="shortcut-icon-box files-icon">
            <Icon name="folder" :size="22" />
          </div>
          <span class="shortcut-label">{{ t.navFiles }}</span>
        </RouterLink>

        <RouterLink to="/shared" class="shortcut-item">
          <div class="shortcut-icon-box shared-icon">
            <Icon name="share" :size="22" />
          </div>
          <span class="shortcut-label">{{ t.navShared }}</span>
        </RouterLink>

        <RouterLink to="/trash" class="shortcut-item">
          <div class="shortcut-icon-box trash-icon">
            <Icon name="trash" :size="22" />
          </div>
          <span class="shortcut-label">{{ t.navTrash }}</span>
        </RouterLink>

        <RouterLink to="/files?view=favorites" class="shortcut-item">
          <div class="shortcut-icon-box fav-icon">
            <Icon name="star-filled" :size="22" />
          </div>
          <span class="shortcut-label">{{ t.tabFavorites }}</span>
        </RouterLink>
      </div>
    </section>

    <section class="card section">
      <h2 class="section-title">{{ t.appearance }}</h2>
      <p class="field-hint appearance-mode-label">{{ t.appearanceModeLabel }}</p>
      <div class="appearance-mode-row" role="radiogroup" :aria-label="t.appearanceModeLabel">
        <button
          v-for="mode in appearanceModes"
          :key="mode.id"
          type="button"
          class="btn appearance-mode-btn"
          :class="{ ink: appearanceMode === mode.id }"
          role="radio"
          :aria-checked="appearanceMode === mode.id"
          @click="chooseAppearanceMode(mode.id)"
        >
          {{ (t as any)[mode.labelKey] }}
        </button>
      </div>
      <p class="field-hint">
        {{
          appearanceMode === 'system'
            ? t.appearanceModeSystemHint
            : resolvedIsDark
              ? t.appearanceModeDarkHint
              : t.appearanceModeLightHint
        }}
      </p>
      <div class="language-row" aria-label="Language">
        <button
          type="button"
          class="btn"
          :class="{ ink: locale === 'vi' }"
          @click="chooseLocale('vi')"
        >
          {{ t.vietnamese }}
        </button>
        <button
          type="button"
          class="btn"
          :class="{ ink: locale === 'en' }"
          @click="chooseLocale('en')"
        >
          {{ t.english }}
        </button>
      </div>
      <p class="field-hint">{{ t.languageHint }}</p>

      <div class="theme-entry-divider"></div>

      <!-- Theme Entry Tile -->
      <RouterLink to="/settings/theme" class="theme-tile-link">
        <div class="theme-tile-info">
          <div class="theme-tile-header">
            <span class="theme-tile-label">{{ t.themeTitle }}</span>
            <span class="theme-current-pill">
              <span class="theme-dot" :style="{ background: currentThemeDef.swatchGradient }"></span>
              {{ (t as any)[currentThemeDef.nameKey] || currentThemeDef.nameDefault }}
            </span>
          </div>
          <span class="theme-tile-hint">{{ t.themeColorPalette }}</span>
        </div>
        <div class="theme-tile-action">
          <span class="theme-action-text">{{ t.seeAll }}</span>
          <Icon name="arrow-right" :size="15" />
        </div>
      </RouterLink>

      <!-- Quick Swatches Grid (balanced 6 columns, no clipping) -->
      <div class="quick-swatches-grid" role="radiogroup" aria-label="Quick themes">
        <button
          v-for="th in THEMES.slice(0, 6)"
          :key="th.id"
          type="button"
          class="quick-swatch"
          :class="{ active: currentColorTheme === th.id }"
          :style="{ background: th.swatchGradient }"
          :title="(t as any)[th.nameKey] || th.nameDefault"
          :aria-label="(t as any)[th.nameKey] || th.nameDefault"
          @click="applyColorTheme(th.id)"
        >
          <span v-if="currentColorTheme === th.id" class="quick-swatch-check" aria-hidden="true">
            <Icon name="check" :size="10" />
          </span>
        </button>
      </div>
    </section>

    <section class="card section">
      <h2 class="section-title">{{ t.trash }}</h2>
      <label class="toggle-row">
        <input v-model="trashAutoDeleteEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">{{ t.autoDeleteTrash }}</span>
      </label>
      <p class="field-hint">{{ t.trashHint }}</p>
      <label class="field">
        <span>{{ t.retentionDays }}</span>
        <input v-model.number="trashRetentionDays" type="number" min="1" inputmode="numeric" />
      </label>
      <button class="btn save-btn" type="button" :disabled="savingSettings" @click="saveSettings">
        {{ savingSettings ? t.saving : t.saveTrashSettings }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">{{ t.mediaPreviews }}</h2>
      <p class="field-hint section-hint">
        {{ t.mediaPreviewsHint }}
      </p>
      <label class="toggle-row">
        <input v-model="imageThumbnailsEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">{{ t.showImageThumbnails }}</span>
      </label>
      <label class="toggle-row">
        <input v-model="videoThumbnailsEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">{{ t.showVideoPreviews }}</span>
      </label>
      <button class="btn" type="button" :disabled="savingSettings" @click="saveSettings">
        {{ savingSettings ? t.saving : t.savePreviewSettings }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">Shared links</h2>
      <p v-if="linksLoading" class="muted">Loading…</p>
      <p v-else-if="shareLinks.length === 0" class="muted">No active share links.</p>
      <ul v-else class="link-list">
        <li v-for="link in shareLinks" :key="link.id" class="link-row">
          <Icon name="share" :size="18" class="row-icon" />
          <span class="link-name">{{ link.fileName }}</span>
          <span class="link-meta muted">
            {{ link.expiresAt ? `Expires ${new Date(link.expiresAt).toLocaleDateString()}` : 'Never expires' }}
            · {{ formatBytes(link.sizeBytes ?? 0) }}
          </span>
          <button type="button" class="btn icon-only" aria-label="Copy link" @click="copyLink(link)">
            <Icon name="file" :size="18" />
          </button>
          <button type="button" class="btn icon-only danger-text" aria-label="Revoke link" @click="revokeFromSettings(link)">
            <Icon name="trash" :size="18" />
          </button>
        </li>
      </ul>
    </section>

    <section class="card section">
      <h2 class="section-title">Activity</h2>
      <p v-if="activityLoading" class="muted">Loading…</p>
      <p v-else-if="activityError" class="muted">
        Could not load activity.
        <button type="button" class="btn retry-btn" @click="loadActivity">Retry</button>
      </p>
      <p v-else-if="activityEvents.length === 0" class="muted">No recent activity.</p>
      <ul v-else class="activity-list">
        <li v-for="ev in activityEvents" :key="ev.id" class="activity-row">
          <span class="activity-icon" aria-hidden="true">
            <Icon :name="activityIcon(ev.type)" :size="18" />
          </span>
          <span class="activity-text">
            <span class="activity-action">{{ activityLabel(ev.type) }}</span>
            <span v-if="ev.targetName" class="activity-target">{{ ev.targetName }}</span>
          </span>
          <time class="activity-time muted" :datetime="ev.createdAt">{{ relativeTime(ev.createdAt) }}</time>
        </li>
      </ul>
      <button
        v-if="activityNextBefore && !activityLoading"
        class="btn block load-more"
        type="button"
        :disabled="activityLoadingMore"
        @click="loadMoreActivity"
      >
        {{ activityLoadingMore ? 'Loading…' : 'Load more' }}
      </button>
    </section>

    <section class="card section">
      <button class="btn danger block" type="button" @click="logout">Log out</button>
    </section>
  </div>
</template>

<style scoped>
.section {
  margin-bottom: var(--space-md);
}

.link-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.link-row {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.link-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.link-meta {
  font-size: 0.75rem;
  white-space: nowrap;
}

.activity-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.activity-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xs) 0;
}

.activity-row + .activity-row {
  border-top: 1px solid var(--hairline);
}

.activity-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: var(--radius-md);
  background: var(--hairline-soft);
  color: var(--muted);
}

.activity-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.activity-action {
  font-weight: 500;
}

.activity-target {
  font-size: 0.8125rem;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.activity-time {
  font-size: 0.75rem;
  white-space: nowrap;
}

.load-more {
  margin-top: var(--space-sm);
}

.retry-btn {
  padding: 0;
  min-height: auto;
  border: none;
  background: none;
  color: var(--accent);
  text-decoration: underline;
}

.danger-text {
  color: var(--danger);
}

.section-hint {
  margin: calc(var(--space-xs) * -1) 0 var(--space-sm);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: var(--touch-min);
  cursor: pointer;
}

.switch-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-track {
  position: relative;
  flex-shrink: 0;
  width: 44px;
  height: 24px;
  border-radius: var(--radius-pill);
  background: var(--hairline);
  transition: background-color var(--motion-press) var(--ease-standard);
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
  transition: transform var(--motion-press) var(--ease-standard);
}

.switch-input:checked + .switch-track {
  background: var(--accent);
}

.switch-input:checked + .switch-track::after {
  transform: translateX(20px);
}

.switch-input:focus-visible + .switch-track {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.toggle-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink);
}

.language-row,
.appearance-mode-row {
  display: flex;
  gap: var(--space-xs);
  margin-top: var(--space-sm);
}

.appearance-mode-label {
  margin-bottom: 0;
}

.appearance-mode-btn {
  flex: 1;
  min-width: 0;
}

.save-btn {
  width: 100%;
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .save-btn {
    width: auto;
  }

  .desktop-only {
    display: block;
  }
}

.settings-page {
  padding-bottom: var(--space-xl);
}

.profile-header-card {
  margin-bottom: var(--space-md);
  padding: var(--space-md);
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-xl);
}

.profile-header-main {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.profile-avatar-box {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-pill);
  background: linear-gradient(135deg, #0284c7 0%, #0d9488 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.2);
}

.profile-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-avatar-initials {
  font-size: 1.125rem;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 0.5px;
}

.profile-meta {
  flex: 1;
  min-width: 0;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.profile-name {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.profile-email {
  margin: 2px 0 0;
  font-size: 0.8125rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-arrow-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-pill);
  background: var(--surface-soft);
  color: var(--muted);
  flex-shrink: 0;
  text-decoration: none;
  transition: transform var(--duration-short) var(--ease-standard), background-color var(--duration-short) var(--ease-standard);
}

.profile-arrow-link:hover {
  color: var(--ink);
  background: var(--hairline-soft);
}

/* Storage card */
.storage-card {
  margin-bottom: var(--space-md);
  padding: var(--space-md);
  border-radius: var(--radius-xl);
  background: var(--surface);
  border: 1px solid var(--hairline);
}

.storage-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-sm);
}

.storage-title-wrap {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.storage-cloud-icon {
  color: var(--accent);
}

.storage-title {
  font-weight: 700;
  font-size: 0.9375rem;
  color: var(--ink);
}

.storage-stats-text {
  font-size: 0.8125rem;
  color: var(--muted);
}

.storage-stats-text strong {
  color: var(--ink);
}

.storage-bar-track {
  width: 100%;
  height: 8px;
  border-radius: var(--radius-pill);
  background: var(--hairline);
  overflow: hidden;
  margin-bottom: var(--space-md);
}

.storage-bar-fill {
  height: 100%;
  border-radius: var(--radius-pill);
  background: var(--accent);
  transition: width var(--duration-medium) var(--ease-standard);
}

.storage-bar-fill.warning {
  background: var(--warning);
}

.storage-bar-fill.danger {
  background: var(--danger);
}

.storage-status {
  display: flex;
  align-items: flex-start;
  gap: var(--space-xs);
  margin: 0 0 var(--space-md);
  font-size: 0.8125rem;
  line-height: 1.35;
}

.storage-status.warning {
  color: var(--warning);
}

.storage-status.full {
  color: var(--danger);
}

.storage-status-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.storage-action-link {
  flex-shrink: 0;
  align-self: center;
  padding: 4px var(--space-sm);
  border-radius: var(--radius-pill);
  background: var(--danger-soft, rgba(239, 68, 68, 0.12));
  color: var(--danger);
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
}

.storage-retry-btn {
  margin-bottom: var(--space-md);
}

.storage-shortcuts-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-xs);
  padding-top: var(--space-xs);
  border-top: 1px solid var(--hairline);
}

.shortcut-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  text-decoration: none;
  color: var(--ink);
  padding: var(--space-xs) 0;
  border-radius: var(--radius-lg);
  transition: background-color var(--duration-short) var(--ease-standard);
}

.shortcut-item:hover {
  background: var(--surface-soft);
}

.shortcut-icon-box {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.files-icon {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
}

.shared-icon {
  background: rgba(13, 148, 136, 0.12);
  color: #0d9488;
}

.trash-icon {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}

.fav-icon {
  background: rgba(59, 130, 246, 0.12);
  color: #2563eb;
}

.shortcut-label {
  font-size: 0.75rem;
  font-weight: 500;
  text-align: center;
  white-space: nowrap;
}

.theme-entry-divider {
  height: 1px;
  background: var(--hairline);
  margin: var(--space-md) 0 var(--space-sm);
}

.theme-tile-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
  text-decoration: none;
  color: inherit;
  margin-bottom: var(--space-xs);
  transition: opacity var(--duration-short) var(--ease-standard);
}

.theme-tile-link:hover {
  opacity: 0.8;
}

.theme-tile-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-tile-header {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.theme-tile-label {
  font-weight: 600;
  font-size: 0.9375rem;
  color: var(--ink);
}

.theme-current-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--muted);
  background: var(--surface-soft);
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  border: 1px solid var(--hairline);
}

.theme-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.theme-tile-hint {
  font-size: 0.75rem;
  color: var(--muted);
}

.theme-tile-action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--accent);
  font-size: 0.8125rem;
  font-weight: 600;
}

.quick-swatches-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: var(--space-xs);
  padding: 4px 2px;
}

.quick-swatch {
  position: relative;
  aspect-ratio: 1 / 1;
  width: 100%;
  max-width: 44px;
  margin: 0 auto;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  padding: 0;
  transition: transform var(--duration-short) var(--ease-standard);
}

.quick-swatch:active {
  transform: scale(0.92);
}

.quick-swatch.active {
  box-shadow: 0 0 0 2px var(--canvas), 0 0 0 4px var(--accent);
}

.quick-swatch-check {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #ffffff;
  color: #111827;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
}

@media (prefers-reduced-motion: reduce) {
  .switch-track,
  .switch-track::after {
    transition: none;
  }
}
</style>
