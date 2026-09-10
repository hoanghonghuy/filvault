<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import { mimeIcon } from '@/lib/mimeIcon'
import { useI18n } from '@/lib/i18n'
import type { DownloadURL, IncomingShare, SharedBrowser } from '@/api/types'

interface ShareLinkInfo {
  token: string
  fileId: string
  fileName: string
  url?: string
  expiresAt: string | null
  createdAt: string
}

const ui = useUiStore()
const { t } = useI18n()

const shareTab = ref<'my-shares' | 'with-me'>('my-shares')
const loading = ref(false)
const error = ref('')
const shares = ref<IncomingShare[]>([])
const shareLinks = ref<ShareLinkInfo[]>([])
const browsing = ref<SharedBrowser | null>(null)
const viewMode = ref<'list' | 'grid'>('list')

function toggleViewMode() {
  viewMode.value = viewMode.value === 'list' ? 'grid' : 'list'
}

async function loadIncomingShares() {
  try {
    const out = await api<{ shares: IncomingShare[] }>('/shares/with-me')
    shares.value = out.shares
  } catch (e) {
    error.value = formatApiError(e, 'Could not load shared items')
  }
}

async function loadMyShareLinks() {
  try {
    const out = await api<{ links: ShareLinkInfo[] }>('/share-links')
    shareLinks.value = out.links
  } catch (e) {
    error.value = formatApiError(e, 'Could not load your share links')
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    await Promise.all([loadIncomingShares(), loadMyShareLinks()])
  } finally {
    loading.value = false
  }
}

function formatDate(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function formatExpiry(expiresAt: string | null): string {
  if (!expiresAt) return t.value.activeForever || 'Có hiệu lực vĩnh viễn'
  const exp = new Date(expiresAt)
  if (exp.getTime() < Date.now()) return 'Đã hết hạn'
  return `Hết hạn ${formatDate(expiresAt)}`
}

function getFileTypeColor(name?: string): string {
  if (!name) return '#0084ff'
  const ext = name.split('.').pop()?.toLowerCase() ?? ''
  if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext)) return '#8b5cf6'
  if (['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) return '#ec4899'
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return '#f97316'
  if (['mp3', 'wav', 'ogg', 'flac'].includes(ext)) return '#06b6d4'
  if (['pdf'].includes(ext)) return '#ef4444'
  if (['xls', 'xlsx', 'csv'].includes(ext)) return '#10b981'
  return '#0084ff'
}

async function browseFolder(share: IncomingShare) {
  error.value = ''
  try {
    browsing.value = await api<SharedBrowser>(`/shared/folders/${share.resourceId}`)
  } catch (e) {
    error.value = formatApiError(e, 'Could not open folder')
  }
}

async function downloadFile(fileId: string, name: string) {
  error.value = ''
  try {
    const out = await api<DownloadURL>(`/shared/files/${fileId}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
    ui.showToast(`Downloading \"${name}\"`, 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Could not download file')
  }
}

function onIncomingRowClick(share: IncomingShare) {
  if (browsing.value) return
  if (share.resourceType === 'folder') void browseFolder(share)
  else void downloadFile(share.resourceId, share.resourceName)
}

async function copyLink(link: ShareLinkInfo) {
  try {
    const fullUrl = window.location.origin + (link.url ?? `/s/${link.token}`)
    await navigator.clipboard.writeText(fullUrl)
    ui.showToast('Link copied')
  } catch {
    ui.showToast('Copy failed', 'info')
  }
}

async function revokeLink(link: ShareLinkInfo) {
  const ok = await ui.confirm({
    title: 'Revoke link?',
    message: `\"${link.fileName}\" will no longer be shared publicly.`,
    confirmLabel: 'Revoke link',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/files/${link.fileId}/share`, { method: 'DELETE' })
    ui.showToast('Link revoked')
    await loadMyShareLinks()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to revoke link')
  }
}

async function openMyShareActions(link: ShareLinkInfo) {
  const fullUrl = window.location.origin + (link.url ?? `/s/${link.token}`)
  const action = await ui.openActionSheet(link.fileName, [
    { id: 'copy', label: 'Copy link', icon: 'copy' },
    { id: 'open', label: 'Open link', icon: 'external-link' },
    { id: 'revoke', label: 'Revoke link', icon: 'trash', danger: true },
  ])
  if (action === 'copy') await copyLink(link)
  if (action === 'open') window.open(fullUrl, '_blank', 'noopener')
  if (action === 'revoke') await revokeLink(link)
}

onMounted(load)
</script>

<template>
  <div class="shared-page">
    <h1 class="page-title desktop-only">{{ t.navShared || 'Chia sẻ' }}</h1>

    <div class="tabs-header">
      <div class="tabs-pill-list" role="tablist" aria-label="Shares views">
        <button
          type="button"
          role="tab"
          class="tab-pill"
          :class="{ active: shareTab === 'my-shares' }"
          :aria-selected="shareTab === 'my-shares'"
          @click="shareTab = 'my-shares'"
        >
          {{ t.myShares || 'Chia sẻ của tôi' }}
          <span v-if="shareLinks.length" class="tab-count-badge">{{ shareLinks.length }}</span>
        </button>
        <button
          type="button"
          role="tab"
          class="tab-pill"
          :class="{ active: shareTab === 'with-me' }"
          :aria-selected="shareTab === 'with-me'"
          @click="shareTab = 'with-me'"
        >
          {{ t.sharedWithMeTab || 'Được chia sẻ' }}
          <span v-if="shares.length" class="tab-count-badge">{{ shares.length }}</span>
        </button>
      </div>
    </div>

    <div v-if="!browsing" class="shares-sub-bar">
      <button
        type="button"
        class="sub-icon-btn"
        :title="viewMode === 'list' ? t.viewGrid : t.viewList"
        :aria-label="viewMode === 'list' ? t.viewGrid : t.viewList"
        :aria-pressed="viewMode === 'grid'"
        @click="toggleViewMode"
      >
        <svg
          v-if="viewMode === 'list'"
          class="view-mode-glyph"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <rect x="4" y="4" width="6" height="6" rx="1" />
          <rect x="14" y="4" width="6" height="6" rx="1" />
          <rect x="4" y="14" width="6" height="6" rx="1" />
          <rect x="14" y="14" width="6" height="6" rx="1" />
        </svg>
        <svg v-else class="view-mode-glyph" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M8 6h12M8 12h12M8 18h12" />
          <circle cx="4" cy="6" r="1" />
          <circle cx="4" cy="12" r="1" />
          <circle cx="4" cy="18" r="1" />
        </svg>
      </button>
    </div>

    <Transition name="page">
      <section v-if="browsing" class="browse" aria-label="Shared folder contents">
        <button type="button" class="back-btn" @click="browsing = null">
          <Icon name="arrow-left" :size="18" />
          {{ t.allSharedItems || 'Tất cả mục chia sẻ' }}
        </button>
        <h2 class="folder-name">{{ browsing.folder?.name }}</h2>
        <div class="shares-container" :class="{ 'grid-mode': viewMode === 'grid' }">
          <div
            v-for="f in browsing.folders"
            :key="'d-' + f.id"
            class="share-item-card static"
          >
            <div class="share-icon-badge folder-badge">
              <Icon name="folder" :size="22" />
            </div>
            <div class="share-item-info">
              <span class="share-item-title">{{ f.name }}</span>
              <span class="share-item-sub">{{ t.folder }}</span>
            </div>
          </div>
          <div
            v-for="f in browsing.files"
            :key="'f-' + f.id"
            class="share-item-card tappable"
            @click="downloadFile(f.id, f.name)"
          >
            <div
              class="share-icon-badge"
              :style="{
                background: `color-mix(in srgb, ${getFileTypeColor(f.name)} 14%, transparent)`,
                color: getFileTypeColor(f.name),
              }"
            >
              <Icon :name="mimeIcon(f.mimeType)" :size="20" />
            </div>
            <div class="share-item-info">
              <span class="share-item-title">{{ f.name }}</span>
              <span class="share-item-sub">{{ formatBytes(f.sizeBytes) }}</span>
            </div>
            <button class="share-action-btn" type="button" aria-label="Download">
              <Icon name="download" :size="18" />
            </button>
          </div>
        </div>
        <p v-if="!browsing.folders.length && !browsing.files.length" class="empty-inline">
          {{ t.folderEmpty }}
        </p>
      </section>
    </Transition>

    <template v-if="!browsing">
      <p v-if="error" class="error" role="alert">
        {{ error }}
        <button type="button" class="retry-btn" @click="load">{{ t.retry }}</button>
      </p>

      <div v-if="loading" class="shares-container">
        <div v-for="i in 4" :key="i" class="skeleton sk-row" />
      </div>

      <div v-else-if="shareTab === 'my-shares'">
        <EmptyState
          v-if="!shareLinks.length"
          icon="share"
          title="Chưa có liên kết chia sẻ nào"
          description="Khi bạn tạo liên kết công khai cho tệp, chúng sẽ xuất hiện tại đây."
        />
        <div v-else class="shares-container" :class="{ 'grid-mode': viewMode === 'grid' }">
          <div
            v-for="link in shareLinks"
            :key="link.token"
            class="share-item-card tappable"
            @click="openMyShareActions(link)"
          >
            <div
              class="share-icon-badge"
              :style="{
                background: `color-mix(in srgb, ${getFileTypeColor(link.fileName)} 14%, transparent)`,
                color: getFileTypeColor(link.fileName),
              }"
            >
              <Icon name="file" :size="20" />
            </div>
            <div class="share-item-info">
              <span class="share-item-title">{{ link.fileName }}</span>
              <span class="share-item-sub">{{ formatDate(link.createdAt) }} · {{ formatExpiry(link.expiresAt) }}</span>
            </div>
            <button
              class="share-action-btn"
              type="button"
              aria-label="Share actions"
              @click.stop="openMyShareActions(link)"
            >
              <Icon name="more" :size="18" />
            </button>
          </div>
        </div>
      </div>

      <div v-else-if="shareTab === 'with-me'">
        <EmptyState
          v-if="!shares.length"
          icon="users"
          :title="t.nothingShared"
          :description="t.nothingSharedDesc"
        />
        <div v-else class="shares-container" :class="{ 'grid-mode': viewMode === 'grid' }">
          <div
            v-for="share in shares"
            :key="share.id"
            class="share-item-card tappable"
            @click="onIncomingRowClick(share)"
          >
            <div
              class="share-icon-badge"
              :class="{ 'folder-badge': share.resourceType === 'folder' }"
              :style="
                share.resourceType === 'folder'
                  ? {}
                  : {
                      background: `color-mix(in srgb, ${getFileTypeColor(share.resourceName)} 14%, transparent)`,
                      color: getFileTypeColor(share.resourceName),
                    }
              "
            >
              <Icon :name="share.resourceType === 'folder' ? 'folder' : 'file'" :size="20" />
            </div>
            <div class="share-item-info">
              <span class="share-item-title">{{ share.resourceName }}</span>
              <span class="share-item-sub">
                {{ share.owner.displayName }} · {{ formatDate(share.createdAt) }}
              </span>
            </div>
            <button
              class="share-action-btn"
              type="button"
              :aria-label="share.resourceType === 'folder' ? t.open : t.download"
              @click.stop="onIncomingRowClick(share)"
            >
              <Icon :name="share.resourceType === 'folder' ? 'chevron-right' : 'download'" :size="18" />
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.shared-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.tabs-header {
  border-bottom: 1px solid var(--hairline);
  padding-bottom: 4px;
}

.tabs-pill-list {
  display: flex;
  align-items: center;
  gap: 16px;
}

.tab-pill {
  background: transparent;
  border: none;
  padding: 6px 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--muted);
  cursor: pointer;
  position: relative;
  transition: color 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.tab-pill.active {
  color: var(--ink);
}

.tab-pill.active::after {
  content: '';
  position: absolute;
  bottom: -5px;
  left: 0;
  right: 0;
  height: 3px;
  border-radius: 3px;
  background: var(--accent);
}

.tab-count-badge {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 9999px;
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

.shares-sub-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-height: var(--touch-min);
}

.sub-icon-btn {
  background: transparent;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-sm, 8px);
  width: var(--touch-min);
  height: var(--touch-min);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  cursor: pointer;
}

.sub-icon-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.view-mode-glyph {
  width: 20px;
  height: 20px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.shares-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.shares-container.grid-mode {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

@media (min-width: 640px) {
  .shares-container.grid-mode {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .shares-container.grid-mode {
    grid-template-columns: repeat(4, 1fr);
  }
}

.share-item-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-md, 12px);
  background: var(--canvas);
  border: 1px solid transparent;
  transition: background-color 0.15s ease, border-color 0.15s ease;
  user-select: none;
  cursor: pointer;
}

.share-item-card:hover {
  background: var(--surface-card, rgba(255, 255, 255, 0.03));
}

.share-item-card:active {
  background: var(--accent-soft);
  border-color: var(--accent);
}

.share-item-card.static {
  cursor: default;
}

.share-icon-badge {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.folder-badge {
  background: rgba(245, 158, 11, 0.14);
  color: #f59e0b;
}

.share-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.share-item-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.share-item-sub {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-variant-numeric: tabular-nums;
}

.share-action-btn {
  background: transparent;
  border: none;
  color: var(--muted);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.share-action-btn:active {
  background: var(--surface-card);
  color: var(--ink);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: var(--accent);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  padding: 4px 0;
  margin-bottom: 8px;
}

.folder-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
}

.empty-inline {
  margin: var(--space-md) 0;
  font-size: 14px;
  color: var(--muted);
}

.error {
  font-size: 14px;
  color: var(--danger);
}

.retry-btn {
  margin-left: 6px;
  background: none;
  border: none;
  color: var(--accent);
  font-weight: 600;
  cursor: pointer;
}

.sk-row {
  height: 52px;
  border-radius: var(--radius-md);
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: block;
  }
}
</style>
