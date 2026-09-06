<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import { mimeIcon } from '@/lib/mimeIcon'
import { useI18n } from '@/lib/i18n'
import type {
  DownloadURL,
  IncomingShare,
  SharedBrowser,
} from '@/api/types'

const ui = useUiStore()
const { t } = useI18n()

const loading = ref(false)
const error = ref('')
const shares = ref<IncomingShare[]>([])
// Folder currently browsed (one level deep per spec 09 §5.5).
const browsing = ref<SharedBrowser | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const out = await api<{ shares: IncomingShare[] }>('/shares/with-me')
    shares.value = out.shares
  } catch (e) {
    error.value = formatApiError(e, 'Could not load shared items')
  } finally {
    loading.value = false
  }
}

function relativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime()
  const minutes = Math.round(diffMs / 60000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.round(hours / 24)
  if (days < 30) return `${days}d ago`
  return new Date(iso).toLocaleDateString()
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
    ui.showToast(`Downloading "${name}"`, 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Could not download file')
  }
}

function onRowClick(share: IncomingShare) {
  if (browsing.value) return
  if (share.resourceType === 'folder') void browseFolder(share)
  else void downloadFile(share.resourceId, share.resourceName)
}

onMounted(load)
</script>

<template>
  <div class="shared-page">
    <h1 class="page-title desktop-only">{{ t.sharedWithMe }}</h1>

    <Transition name="page">
      <section v-if="browsing" class="browse" aria-label="Shared folder contents">
        <button type="button" class="back-btn" @click="browsing = null">
          <Icon name="restore" :size="16" />
          {{ t.allSharedItems }}
        </button>
        <h2 class="folder-name">{{ browsing.folder?.name }}</h2>
        <ul v-if="browsing.folders.length" class="rows">
          <li v-for="f in browsing.folders" :key="'d-' + f.id" class="row static">
            <span class="row-icon"><Icon name="folder" :size="18" /></span>
            <span class="row-name">{{ f.name }}</span>
            <span class="row-meta">{{ t.folder }}</span>
          </li>
        </ul>
        <ul v-if="browsing.files.length" class="rows">
          <li v-for="f in browsing.files" :key="'f-' + f.id">
            <button type="button" class="row tappable" @click="downloadFile(f.id, f.name)">
              <span class="row-icon"><Icon :name="mimeIcon(f.mimeType)" :size="18" /></span>
              <span class="row-name">{{ f.name }}</span>
              <span class="row-meta">{{ formatBytes(f.sizeBytes) }}</span>
            </button>
          </li>
        </ul>
        <p v-if="!browsing.folders.length && !browsing.files.length" class="empty-inline">
          {{ t.folderEmpty }}
        </p>
      </section>
    </Transition>

    <template v-if="!browsing">
      <p v-if="error" class="error" role="alert">{{ error }} <button type="button" class="retry-btn" @click="load">{{ t.retry }}</button></p>

      <div v-if="loading" class="list" aria-busy="true" aria-live="polite">
        <div v-for="i in 3" :key="i" class="skeleton sk-row" />
      </div>

      <EmptyState
        v-else-if="!shares.length"
        icon="users"
        :title="t.nothingShared"
        :description="t.nothingSharedDesc"
      />

      <ul v-else class="rows" :aria-label="t.sharedWithMe">
        <li v-for="share in shares" :key="share.id">
          <button type="button" class="row tappable" @click="onRowClick(share)">
            <span class="row-icon">
              <Icon
                :name="share.resourceType === 'folder' ? 'folder' : mimeIcon('')"
                :size="18"
              />
            </span>
            <span class="row-copy">
              <span class="row-name">{{ share.resourceName }}</span>
              <span class="row-sub">by {{ share.owner.displayName }} · {{ relativeTime(share.createdAt) }}</span>
            </span>
            <span class="row-meta">
              {{ share.resourceType === 'folder' ? t.open : t.download }}
            </span>
          </button>
        </li>
      </ul>
    </template>
  </div>
</template>

<style scoped>
.shared-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xxs);
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
  margin-left: calc(-1 * var(--space-sm));
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  color: var(--accent);
}

.back-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.folder-name {
  margin: 0 0 var(--space-xs);
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--ink);
}

.rows {
  display: flex;
  flex-direction: column;
  gap: var(--space-xxs);
  margin: 0;
  padding: 0;
  list-style: none;
}

.row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  min-height: var(--touch-min);
  padding: var(--space-xs) var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface);
  color: var(--ink);
  text-align: left;
  transition:
    background-color var(--motion-press) var(--ease-standard),
    border-color var(--motion-press) var(--ease-standard);
}

.row.static {
  cursor: default;
}

.row.tappable:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

@media (hover: hover) {
  .row.tappable:hover {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
}

.row-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  color: var(--muted);
}

.row-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.row-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  font-weight: 600;
}

.row-sub {
  font-size: 12px;
  color: var(--muted);
}

.row-meta {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 500;
  color: var(--muted);
}

.empty-inline {
  margin: 0;
  font-size: 14px;
  color: var(--muted);
}

.error {
  margin: 0;
  font-size: 14px;
  color: var(--danger);
}

.retry-btn {
  margin-left: var(--space-xs);
  font-weight: 600;
  color: var(--accent);
}

.sk-row {
  height: var(--touch-min);
  border-radius: var(--radius-md);
}
</style>
