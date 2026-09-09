<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import LoadingSkeletonTrash from '@/components/LoadingSkeletonTrash.vue'
import { mimeIcon, mimeLabel } from '@/lib/mimeIcon'
import { useI18n } from '@/lib/i18n'
import type { TrashList } from '@/api/types'

const ui = useUiStore()
const { t } = useI18n()
const trash = ref<TrashList | null>(null)
const loading = ref(false)
const error = ref('')

const isEmpty = computed(() => {
  if (!trash.value) return false
  return trash.value.folders.length === 0 && trash.value.files.length === 0
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    trash.value = await api<TrashList>('/trash')
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load trash')
  } finally {
    loading.value = false
  }
}

/** Optimistic UI: drop the row immediately so TransitionGroup animates the removal. */
function removeFromLocal(id: string) {
  const t = trash.value
  if (!t) return
  t.folders = t.folders.filter((folder) => folder.id !== id)
  t.files = t.files.filter((file) => file.id !== id)
}

async function restoreFile(id: string) {
  removeFromLocal(id)
  try {
    await api(`/files/${id}/restore`, { method: 'POST', body: '{}' })
    ui.showToast(t.value.fileRestored)
  } catch (e) {
    error.value = formatApiError(e, 'Restore failed')
    await load()
  }
}

async function restoreFolder(id: string) {
  removeFromLocal(id)
  try {
    await api(`/folders/${id}/restore`, { method: 'POST', body: '{}' })
    ui.showToast(t.value.folderRestored)
  } catch (e) {
    error.value = formatApiError(e, 'Restore failed')
    await load()
  }
}

async function permanentDelete(type: 'files' | 'folders', id: string, name: string) {
  const ok = await ui.confirm({
    title: `${t.value.deleteForever}?`,
    message: `"${name}" will be permanently deleted. This cannot be undone.`,
    confirmLabel: t.value.deleteForever,
    danger: true,
  })
  if (!ok) return
  removeFromLocal(id)
  try {
    await api(`/trash/${type}/${id}`, { method: 'DELETE' })
    ui.showToast(t.value.deleteForever)
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
    await load()
  }
}

async function openFolderActions(folder: { id: string; name: string }) {
  const action = await ui.openActionSheet(folder.name, [
    { id: 'restore', label: 'Restore', icon: 'restore' },
    { id: 'delete', label: 'Delete forever', icon: 'trash', danger: true },
  ])
  if (action === 'restore') await restoreFolder(folder.id)
  if (action === 'delete') await permanentDelete('folders', folder.id, folder.name)
}

async function openFileActions(file: { id: string; name: string }) {
  const action = await ui.openActionSheet(file.name, [
    { id: 'restore', label: 'Restore', icon: 'restore' },
    { id: 'delete', label: 'Delete forever', icon: 'trash', danger: true },
  ])
  if (action === 'restore') await restoreFile(file.id)
  if (action === 'delete') await permanentDelete('files', file.id, file.name)
}

const totalCount = computed(() => (trash.value?.folders.length ?? 0) + (trash.value?.files.length ?? 0))

function formatItemDate(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function getFileTypeColor(mimeType?: string): string {
  if (!mimeType) return '#64748b'
  if (mimeType.startsWith('image/')) return '#8b5cf6'
  if (mimeType.startsWith('video/')) return '#ec4899'
  if (mimeType.startsWith('audio/')) return '#06b6d4'
  if (mimeType.includes('pdf')) return '#ef4444'
  if (mimeType.includes('zip') || mimeType.includes('tar') || mimeType.includes('rar') || mimeType.includes('7z')) return '#f97316'
  return '#0084ff'
}

async function emptyAllTrash() {
  const ok = await ui.confirm({
    title: t.value.emptyTrash || 'Dọn sạch thùng rác?',
    message: 'Tất cả tệp và thư mục trong thùng rác sẽ bị xóa vĩnh viễn.',
    confirmLabel: t.value.emptyTrash || 'Dọn sạch',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    for (const f of (trash.value?.folders ?? [])) {
      await api(`/trash/folders/${f.id}`, { method: 'DELETE' })
    }
    for (const f of (trash.value?.files ?? [])) {
      await api(`/trash/files/${f.id}`, { method: 'DELETE' })
    }
    ui.showToast(t.value.deleteForever)
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to empty trash')
    await load()
  }
}

onMounted(load)
</script>

<template>
  <div class="trash-page">
    <h1 class="page-title desktop-only">{{ t.trashTitle }}</h1>

    <!-- TeraBox Trash Retention Banner -->
    <div class="trash-notice-banner">
      <Icon name="info" :size="18" class="notice-icon" />
      <span>{{ t.trashRetentionNotice || 'Tệp trong thùng rác sẽ tự động xóa sau 30 ngày.' }}</span>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonTrash v-if="loading" />
    <div v-else>
      <EmptyState
        v-if="isEmpty"
        :title="t.trashEmpty"
        :description="t.trashEmptyDesc"
        icon="trash"
      />

      <template v-else>
        <!-- Action bar -->
        <div class="trash-toolbar">
          <span class="trash-count-text">{{ totalCount }} {{ t.results }}</span>
          <button type="button" class="empty-trash-btn" @click="emptyAllTrash">
            <Icon name="trash" :size="15" />
            {{ t.emptyTrash || 'Dọn sạch thùng rác' }}
          </button>
        </div>

        <!-- Folders -->
        <TransitionGroup v-if="trash?.folders.length" name="row" tag="section" class="trash-list">
          <h2 key="folders-title" class="section-title">{{ t.folders }}</h2>
          <div
            v-for="folder in trash.folders"
            :key="folder.id"
            class="trash-item-card tappable"
            @click="openFolderActions(folder)"
          >
            <div class="trash-icon-badge folder-badge">
              <Icon name="folder" :size="22" />
            </div>
            <div class="trash-item-info">
              <span class="trash-item-title">{{ folder.name }}</span>
              <span class="trash-item-sub">{{ formatItemDate(folder.deletedAt) }}</span>
            </div>
            <div class="trash-actions">
              <button
                class="trash-action-btn"
                type="button"
                title="Restore"
                @click.stop="restoreFolder(folder.id)"
              >
                <Icon name="restore" :size="18" />
              </button>
              <button
                class="trash-action-btn danger"
                type="button"
                title="Delete forever"
                @click.stop="permanentDelete('folders', folder.id, folder.name)"
              >
                <Icon name="trash" :size="18" />
              </button>
            </div>
          </div>
        </TransitionGroup>

        <!-- Files -->
        <TransitionGroup v-if="trash?.files.length" name="row" tag="section" class="trash-list files-section">
          <h2 key="files-title" class="section-title">{{ t.files }}</h2>
          <div
            v-for="file in trash.files"
            :key="file.id"
            class="trash-item-card tappable"
            @click="openFileActions(file)"
          >
            <div
              class="trash-icon-badge"
              :style="{
                background: `color-mix(in srgb, ${getFileTypeColor(file.mimeType)} 14%, transparent)`,
                color: getFileTypeColor(file.mimeType)
              }"
            >
              <Icon :name="mimeIcon(file.mimeType ?? '')" :size="20" />
            </div>
            <div class="trash-item-info">
              <span class="trash-item-title">{{ file.name }}</span>
              <span class="trash-item-sub">{{ formatItemDate(file.deletedAt) }} · {{ formatBytes(file.sizeBytes ?? 0) }}</span>
            </div>
            <div class="trash-actions">
              <button
                class="trash-action-btn"
                type="button"
                title="Restore"
                @click.stop="restoreFile(file.id)"
              >
                <Icon name="restore" :size="18" />
              </button>
              <button
                class="trash-action-btn danger"
                type="button"
                title="Delete forever"
                @click.stop="permanentDelete('files', file.id, file.name)"
              >
                <Icon name="trash" :size="18" />
              </button>
            </div>
          </div>
        </TransitionGroup>
      </template>
    </div>
  </div>
</template>

<style scoped>
.trash-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.trash-notice-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: var(--radius-md, 12px);
  background: var(--accent-soft, rgba(0, 132, 255, 0.08));
  color: var(--ink);
  font-size: 13px;
  border: 1px solid var(--hairline);
}

.notice-icon {
  color: var(--accent, #0084ff);
  flex-shrink: 0;
}

.trash-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 0;
}

.trash-count-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--muted);
}

.empty-trash-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: var(--radius-pill, 9999px);
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease;
}

.empty-trash-btn:active {
  background: rgba(239, 68, 68, 0.2);
}

.trash-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.trash-item-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-md, 12px);
  background: var(--canvas);
  border: 1px solid transparent;
  transition: background-color 0.15s ease, border-color 0.15s ease;
  user-select: none;
}

.trash-item-card:hover {
  background: var(--surface-card, rgba(255, 255, 255, 0.03));
}

.trash-item-card:active {
  background: var(--surface-card, rgba(255, 255, 255, 0.06));
}

.trash-icon-badge {
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

.trash-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.trash-item-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.trash-item-sub {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-variant-numeric: tabular-nums;
}

.trash-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.trash-action-btn {
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

.trash-action-btn:active {
  background: var(--surface-card);
  color: var(--ink);
}

.trash-action-btn.danger {
  color: #ef4444;
}

.trash-action-btn.danger:active {
  background: rgba(239, 68, 68, 0.12);
}

.files-section {
  margin-top: var(--space-md);
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
