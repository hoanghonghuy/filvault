<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import LoadingSkeletonTrash from '@/components/LoadingSkeletonTrash.vue'
import { mimeIcon } from '@/lib/mimeIcon'
import { runTrashPurge, type TrashPurgeTarget } from '@/lib/trashPurge'
import { trashRetentionNotice } from '@/lib/trashRetention'
import { useI18n } from '@/lib/i18n'
import type { TrashList } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()
const { t, locale } = useI18n()
const trash = ref<TrashList | null>(null)
const loading = ref(false)
const error = ref('')
const emptyingTrash = ref(false)
const purgeCompleted = ref(0)
const purgeTotal = ref(0)

const retentionNotice = computed(() => trashRetentionNotice(auth.user, locale.value))

const trashBatchCopy = computed(() =>
  locale.value === 'vi'
    ? {
        deleting: (completed: number, total: number) => `Đang xóa ${completed}/${total}…`,
        partial: (deleted: number, failed: number) =>
          `Đã xóa vĩnh viễn ${deleted} mục; ${failed} mục không thể xóa.`,
        failed: 'Không thể dọn sạch thùng rác',
      }
    : {
        deleting: (completed: number, total: number) => `Deleting ${completed}/${total}…`,
        partial: (deleted: number, failed: number) =>
          `Permanently deleted ${deleted} items; ${failed} could not be deleted.`,
        failed: 'Could not empty trash',
      },
)

const isEmpty = computed(() => {
  if (!trash.value) return false
  return trash.value.folders.length === 0 && trash.value.files.length === 0
})

const emptyTrashLabel = computed(() => {
  if (!emptyingTrash.value) return t.value.emptyTrash || 'Dọn sạch thùng rác'
  return trashBatchCopy.value.deleting(purgeCompleted.value, purgeTotal.value)
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
  if (emptyingTrash.value) return

  const ok = await ui.confirm({
    title: t.value.emptyTrash || 'Dọn sạch thùng rác?',
    message: 'Tất cả tệp và thư mục trong thùng rác sẽ bị xóa vĩnh viễn.',
    confirmLabel: t.value.emptyTrash || 'Dọn sạch',
    danger: true,
  })
  if (!ok || emptyingTrash.value) return

  const targets: TrashPurgeTarget[] = [
    ...(trash.value?.folders ?? []).map((folder) => ({ type: 'folders' as const, id: folder.id })),
    ...(trash.value?.files ?? []).map((file) => ({ type: 'files' as const, id: file.id })),
  ]
  if (targets.length === 0) return

  emptyingTrash.value = true
  purgeCompleted.value = 0
  purgeTotal.value = targets.length
  error.value = ''

  try {
    const result = await runTrashPurge(
      targets,
      (target) => api(`/trash/${target.type}/${target.id}`, { method: 'DELETE' }),
      (completed) => {
        purgeCompleted.value = completed
      },
    )

    await load()
    if (result.failed > 0) {
      error.value = trashBatchCopy.value.partial(result.deleted, result.failed)
    } else {
      ui.showToast(t.value.deleteForever)
    }
  } catch (e) {
    await load()
    error.value = formatApiError(e, trashBatchCopy.value.failed)
  } finally {
    emptyingTrash.value = false
    purgeCompleted.value = 0
    purgeTotal.value = 0
  }
}

onMounted(load)
</script>

<template>
  <div class="trash-page">
    <h1 class="page-title desktop-only">{{ t.trashTitle }}</h1>

    <div class="trash-notice-banner">
      <Icon name="info" :size="18" class="notice-icon" />
      <span>{{ retentionNotice }}</span>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonTrash v-if="loading && !emptyingTrash" />
    <div v-else>
      <EmptyState
        v-if="isEmpty"
        :title="t.trashEmpty"
        :description="t.trashEmptyDesc"
        icon="trash"
      />

      <template v-else>
        <div class="trash-toolbar">
          <span class="trash-count-text">{{ totalCount }} {{ t.results }}</span>
          <button
            type="button"
            class="empty-trash-btn"
            :disabled="emptyingTrash"
            :aria-busy="emptyingTrash ? 'true' : undefined"
            @click="emptyAllTrash"
          >
            <Icon name="trash" :size="15" />
            {{ emptyTrashLabel }}
          </button>
        </div>

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
                :title="t.restore"
                :aria-label="`${t.restore}: ${folder.name}`"
                :disabled="emptyingTrash"
                @click.stop="restoreFolder(folder.id)"
              >
                <Icon name="restore" :size="18" />
              </button>
              <button
                class="trash-action-btn danger"
                type="button"
                :title="t.deleteForever"
                :aria-label="`${t.deleteForever}: ${folder.name}`"
                :disabled="emptyingTrash"
                @click.stop="permanentDelete('folders', folder.id, folder.name)"
              >
                <Icon name="trash" :size="18" />
              </button>
            </div>
          </div>
        </TransitionGroup>

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
                :title="t.restore"
                :aria-label="`${t.restore}: ${file.name}`"
                :disabled="emptyingTrash"
                @click.stop="restoreFile(file.id)"
              >
                <Icon name="restore" :size="18" />
              </button>
              <button
                class="trash-action-btn danger"
                type="button"
                :title="t.deleteForever"
                :aria-label="`${t.deleteForever}: ${file.name}`"
                :disabled="emptyingTrash"
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
  line-height: 1.4;
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
  gap: var(--space-sm);
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
  justify-content: center;
  min-height: var(--touch-min);
  gap: 6px;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: var(--radius-pill, 9999px);
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease, opacity 0.15s ease;
}

.empty-trash-btn:disabled,
.trash-action-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.empty-trash-btn:active:not(:disabled) {
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
  width: var(--touch-min);
  height: var(--touch-min);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.trash-action-btn:active:not(:disabled) {
  background: var(--surface-card);
  color: var(--ink);
}

.trash-action-btn.danger {
  color: #ef4444;
}

.trash-action-btn.danger:active:not(:disabled) {
  background: rgba(239, 68, 68, 0.12);
}

.files-section {
  margin-top: var(--space-md);
}

.desktop-only {
  display: none;
}

@media (max-width: 479px) {
  .trash-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .empty-trash-btn {
    width: 100%;
  }
}

@media (min-width: 768px) {
  .desktop-only {
    display: block;
  }
}
</style>
