<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import type { TrashList } from '@/api/types'

const ui = useUiStore()
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

async function restoreFile(id: string) {
  try {
    await api(`/files/${id}/restore`, { method: 'POST', body: '{}' })
    ui.showToast('File restored')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Restore failed')
  }
}

async function restoreFolder(id: string) {
  try {
    await api(`/folders/${id}/restore`, { method: 'POST', body: '{}' })
    ui.showToast('Folder restored')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Restore failed')
  }
}

async function permanentDelete(type: 'files' | 'folders', id: string, name: string) {
  const ok = await ui.confirm({
    title: 'Delete forever?',
    message: `"${name}" will be permanently deleted. This cannot be undone.`,
    confirmLabel: 'Delete forever',
    danger: true,
  })
  if (!ok) return
  try {
    await api(`/trash/${type}/${id}`, { method: 'DELETE' })
    ui.showToast('Deleted permanently')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
  }
}

async function openFolderActions(folder: { id: string; name: string }) {
  const action = await ui.openActionSheet(folder.name, [
    { id: 'restore', label: 'Restore' },
    { id: 'delete', label: 'Delete forever', danger: true },
  ])
  if (action === 'restore') await restoreFolder(folder.id)
  if (action === 'delete') await permanentDelete('folders', folder.id, folder.name)
}

async function openFileActions(file: { id: string; name: string }) {
  const action = await ui.openActionSheet(file.name, [
    { id: 'restore', label: 'Restore' },
    { id: 'delete', label: 'Delete forever', danger: true },
  ])
  if (action === 'restore') await restoreFile(file.id)
  if (action === 'delete') await permanentDelete('files', file.id, file.name)
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="page-title desktop-only">Trash</h1>
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="loading" class="muted">Loading…</p>

    <EmptyState
      v-if="isEmpty && !loading"
      title="Trash is empty"
      description="Deleted files and folders will appear here."
    />

    <section v-if="trash?.folders.length" class="list">
      <h2 class="section-title">Folders</h2>
      <div v-for="folder in trash.folders" :key="folder.id" class="row">
        <span class="name"><span class="icon-folder" />{{ folder.name }}</span>
        <button class="btn icon-only" type="button" aria-label="Folder actions" @click="openFolderActions(folder)">
          ⋯
        </button>
      </div>
    </section>

    <section v-if="trash?.files.length" class="list" style="margin-top: 1rem">
      <h2 class="section-title">Files</h2>
      <div v-for="file in trash.files" :key="file.id" class="row">
        <span class="name">{{ file.name }}</span>
        <span class="meta desktop-only">{{ file.mimeType }} · {{ formatBytes(file.sizeBytes ?? 0) }}</span>
        <button class="btn icon-only" type="button" aria-label="File actions" @click="openFileActions(file)">
          ⋯
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: inline;
  }

  .page-title.desktop-only {
    display: block;
  }
}
</style>
