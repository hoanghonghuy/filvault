<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, formatBytes, setTokens } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import Icon from '@/components/AppIcon.vue'
import type { ActivityEvent, ActivityEventType, ShareLinkInfo, User } from '@/api/types'
import { setLocale, useI18n, type Locale } from '@/lib/i18n'

const auth = useAuthStore()
const ui = useUiStore()
const { locale, t } = useI18n()

const displayName = ref(auth.user?.displayName ?? '')
const imageThumbnailsEnabled = ref(auth.user?.imageThumbnailsEnabled ?? true)
const videoThumbnailsEnabled = ref(auth.user?.videoThumbnailsEnabled ?? true)
const trashAutoDeleteEnabled = ref(auth.user?.trashAutoDeleteEnabled ?? false)
const trashRetentionDays = ref(auth.user?.trashRetentionDays ?? 30)
const currentPassword = ref('')
const newPassword = ref('')
const error = ref('')
const savingProfile = ref(false)
const changingPassword = ref(false)
const darkModeEnabled = ref(document.documentElement.dataset.theme === 'dark')

function toggleDarkMode() {
  darkModeEnabled.value = !darkModeEnabled.value
  if (darkModeEnabled.value) {
    document.documentElement.dataset.theme = 'dark'
    localStorage.setItem('filvault.theme', 'dark')
    return
  }
  delete document.documentElement.dataset.theme
  localStorage.setItem('filvault.theme', 'light')
}

function chooseLocale(next: Locale) {
  setLocale(next)
}

async function saveProfile() {
  error.value = ''
  savingProfile.value = true
  try {
    await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({
        displayName: displayName.value,
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
    savingProfile.value = false
  }
}

async function changePassword() {
  error.value = ''
  changingPassword.value = true
  try {
    const tokens = await api<{ accessToken: string; refreshToken: string }>('/users/me/password', {
      method: 'POST',
      body: JSON.stringify({
        currentPassword: currentPassword.value,
        newPassword: newPassword.value,
      }),
    })
    setTokens(tokens.accessToken, tokens.refreshToken)
    await auth.loadMe()
    currentPassword.value = ''
    newPassword.value = ''
    ui.showToast('Password updated')
  } catch (e) {
    error.value = formatApiError(e, 'Password change failed')
  } finally {
    changingPassword.value = false
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
  loadShareLinks()
  loadActivity()
})
</script>

<template>
  <div>
    <h1 class="page-title desktop-only">{{ t.settingsTitle }}</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <section class="card section">
      <h2 class="section-title">{{ t.appearance }}</h2>
      <label class="toggle-row">
        <input :checked="darkModeEnabled" type="checkbox" class="switch-input" @change="toggleDarkMode" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">{{ t.darkMode }}</span>
      </label>
      <p class="field-hint">{{ t.darkModeHint }}</p>
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
      <p class="field-hint">filvault.locale · {{ t.languageHint }}</p>
    </section>

    <section class="card section">
      <h2 class="section-title">{{ t.profile }}</h2>
      <p class="muted">{{ auth.user?.email }}</p>
      <label class="field">
        <span>{{ t.displayName }}</span>
        <input v-model="displayName" autocomplete="nickname" />
      </label>
      <button class="btn ink save-btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? t.saving : t.saveProfile }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">Trash</h2>
      <label class="toggle-row">
        <input v-model="trashAutoDeleteEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">Auto-delete trash</span>
      </label>
      <p class="field-hint">Permanently delete items older than the retention period.</p>
      <label class="field">
        <span>Retention days</span>
        <input v-model.number="trashRetentionDays" type="number" min="1" inputmode="numeric" />
      </label>
      <button class="btn save-btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? 'Saving…' : 'Save trash settings' }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">Media previews</h2>
      <p class="field-hint section-hint">
        Turn previews off to reduce object-storage bandwidth while browsing Photos.
      </p>
      <label class="toggle-row">
        <input v-model="imageThumbnailsEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">Show image thumbnails</span>
      </label>
      <label class="toggle-row">
        <input v-model="videoThumbnailsEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">Show video previews</span>
      </label>
      <button class="btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? 'Saving…' : 'Save preview settings' }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">Password</h2>
      <label class="field">
        <span>Current password</span>
        <input v-model="currentPassword" type="password" autocomplete="current-password" />
      </label>
      <label class="field">
        <span>New password</span>
        <input v-model="newPassword" type="password" minlength="8" autocomplete="new-password" />
      </label>
      <button class="btn ink save-btn" type="button" :disabled="changingPassword" @click="changePassword">
        {{ changingPassword ? 'Changing…' : 'Change password' }}
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

.language-row {
  display: flex;
  gap: var(--space-xs);
  margin-top: var(--space-sm);
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

@media (prefers-reduced-motion: reduce) {
  .switch-track,
  .switch-track::after {
    transition: none;
  }
}
</style>
