<script setup lang="ts">
import { computed, inject, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, formatBytes, uploadToPresigned } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import UploadFab from '@/components/UploadFab.vue'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import FolderPickerSheet from '@/components/FolderPickerSheet.vue'
import LoadingSkeletonFiles from '@/components/LoadingSkeletonFiles.vue'
import { mimeIcon, mimeLabel } from '@/lib/mimeIcon'
import type { Browser, DownloadURL, SearchResult, UploadSession } from '@/api/types'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const reloadStorage = inject<() => Promise<void>>('reloadStorage')

const browser = ref<Browser | null>(null)
const searchResults = ref<SearchResult | null>(null)
const loading = ref(false)
const searchLoading = ref(false)
const error = ref('')
const uploadProgress = ref<number | null>(null)
const searchQuery = ref('')
const newFolderName = ref('')
const folderSheetOpen = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const pickerOpen = ref(false)
const pickerMode = ref<'file' | 'folder' | null>(null)
const pickerTargetId = ref<string | null>(null)
const pickerTitle = ref('Move to')

const folderId = computed(() => {
  const raw = route.query.folderId
  return typeof raw === 'string' && raw ? raw : null
})

const pickerExcludeFolderId = computed(() => (pickerMode.value === 'folder' ? pickerTargetId.value : null))

const isEmpty = computed(() => {
  if (searchResults.value) return false
  const b = browser.value
  if (!b) return false
  return b.folders.length === 0 && b.files.length === 0
})

async function loadBrowser() {
  error.value = ''
  searchResults.value = null
  loading.value = true
  try {
    const q = folderId.value ? `?folderId=${folderId.value}` : ''
    browser.value = await api<Browser>(`/browser${q}`)
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load files')
  } finally {
    loading.value = false
  }
}

async function runSearch() {
  if (!searchQuery.value.trim()) {
    searchResults.value = null
    searchLoading.value = false
    return
  }
  error.value = ''
  searchLoading.value = true
  try {
    searchResults.value = await api<SearchResult>(`/search?q=${encodeURIComponent(searchQuery.value.trim())}`)
  } catch (e) {
    error.value = formatApiError(e, 'Search failed')
  } finally {
    searchLoading.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(searchQuery, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(runSearch, 300)
})

async function openFolder(id: string | null) {
  await router.push(id ? { path: '/files', query: { folderId: id } } : { path: '/files' })
}

async function createFolder() {
  if (!newFolderName.value.trim()) return
  error.value = ''
  try {
    await api('/folders', {
      method: 'POST',
      body: JSON.stringify({ name: newFolderName.value.trim(), parentId: folderId.value }),
    })
    newFolderName.value = ''
    folderSheetOpen.value = false
    ui.showToast('Folder created')
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to create folder')
  }
}

function triggerUpload() {
  fileInputRef.value?.click()
}

async function onUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  error.value = ''
  uploadProgress.value = 0
  try {
    const session = await api<UploadSession>('/files/upload-sessions', {
      method: 'POST',
      body: JSON.stringify({
        name: file.name,
        size: file.size,
        contentType: file.type || 'application/octet-stream',
        folderId: folderId.value,
      }),
    })
    await uploadToPresigned(session.uploadUrl, file, file.type || 'application/octet-stream', (p) => {
      uploadProgress.value = p
    })
    await api(`/files/${session.fileId}/complete`, { method: 'POST', body: '{}' })
    ui.showToast('Upload complete')
    await loadBrowser()
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, 'Upload failed')
  } finally {
    uploadProgress.value = null
  }
}

async function downloadFile(id: string) {
  try {
    const out = await api<DownloadURL>(`/files/${id}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, 'Download failed')
  }
}

async function renameFile(id: string, current: string) {
  const name = await ui.prompt({ title: 'Rename file', label: 'Name', initialValue: current })
  if (!name || name === current) return
  error.value = ''
  try {
    await api(`/files/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) })
    ui.showToast('File renamed')
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, 'Rename failed')
  }
}

async function renameFolder(id: string, current: string) {
  const name = await ui.prompt({ title: 'Rename folder', label: 'Name', initialValue: current })
  if (!name || name === current) return
  error.value = ''
  try {
    await api(`/folders/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) })
    ui.showToast('Folder renamed')
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, 'Rename failed')
  }
}

function openMoveFile(id: string) {
  pickerTargetId.value = id
  pickerMode.value = 'file'
  pickerTitle.value = 'Move file'
  pickerOpen.value = true
}

function openMoveFolder(id: string) {
  pickerTargetId.value = id
  pickerMode.value = 'folder'
  pickerTitle.value = 'Move folder'
  pickerOpen.value = true
}

async function onPickerSelect(targetFolderId: string | null) {
  pickerOpen.value = false
  const id = pickerTargetId.value
  const mode = pickerMode.value
  pickerTargetId.value = null
  pickerMode.value = null
  if (!id || !mode) return

  error.value = ''
  try {
    if (mode === 'file') {
      await api(`/files/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ folderId: targetFolderId }),
      })
      ui.showToast('File moved')
    } else {
      await api(`/folders/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ parentId: targetFolderId }),
      })
      ui.showToast('Folder moved')
    }
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, 'Move failed')
  }
}

async function deleteFile(id: string) {
  const ok = await ui.confirm({
    title: 'Move to trash?',
    message: 'You can restore this file from Trash later.',
    confirmLabel: 'Move to trash',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/files/${id}`, { method: 'DELETE' })
    ui.showToast('Moved to trash')
    await loadBrowser()
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
  }
}

async function deleteFolder(id: string) {
  const ok = await ui.confirm({
    title: 'Move folder to trash?',
    message: 'The folder must be empty. You can restore it from Trash later.',
    confirmLabel: 'Move to trash',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/folders/${id}`, { method: 'DELETE' })
    ui.showToast('Moved to trash')
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
  }
}

async function openFolderActions(folder: { id: string; name: string }) {
  const action = await ui.openActionSheet(folder.name, [
    { id: 'open', label: 'Open' },
    { id: 'rename', label: 'Rename' },
    { id: 'move', label: 'Move' },
    { id: 'delete', label: 'Move to trash', danger: true },
  ])
  if (action === 'open') await openFolder(folder.id)
  if (action === 'rename') await renameFolder(folder.id, folder.name)
  if (action === 'move') openMoveFolder(folder.id)
  if (action === 'delete') await deleteFolder(folder.id)
}

async function openFileActions(file: { id: string; name: string }) {
  const action = await ui.openActionSheet(file.name, [
    { id: 'download', label: 'Download' },
    { id: 'rename', label: 'Rename' },
    { id: 'move', label: 'Move' },
    { id: 'delete', label: 'Move to trash', danger: true },
  ])
  if (action === 'download') await downloadFile(file.id)
  if (action === 'rename') await renameFile(file.id, file.name)
  if (action === 'move') openMoveFile(file.id)
  if (action === 'delete') await deleteFile(file.id)
}

watch(() => route.query.folderId, loadBrowser, { immediate: true })
</script>

<template>
  <div class="files-page">
    <h1 class="page-title desktop-only">My Files</h1>

    <nav class="breadcrumb" aria-label="Folder path">
      <button
        v-if="folderId"
        type="button"
        class="btn ghost mobile-back"
        @click="openFolder(browser?.folder?.parentId ?? null)"
      >
        ← Back
      </button>
      <span class="breadcrumb-trail desktop-only">
        <a href="#" @click.prevent="openFolder(null)">Root</a>
        <template v-for="item in browser?.breadcrumb ?? []" :key="item.id">
          <span aria-hidden="true">/</span>
          <a href="#" @click.prevent="openFolder(item.id)">{{ item.name }}</a>
        </template>
        <template v-if="browser?.folder">
          <span aria-hidden="true">/</span>
          <span class="current">{{ browser.folder.name }}</span>
        </template>
      </span>
      <span v-if="browser?.folder" class="mobile-current">{{ browser.folder.name }}</span>
      <span v-else-if="!folderId" class="mobile-current">Root</span>
    </nav>

    <div class="toolbar toolbar-sticky">
      <input
        v-model="searchQuery"
        class="search-input"
        type="search"
        placeholder="Search by name"
        enterkeyhint="search"
      />
      <button class="btn accent desktop-only" type="button" @click="triggerUpload">Upload</button>
      <button class="btn desktop-only" type="button" @click="folderSheetOpen = true">New folder</button>
      <button class="btn mobile-only" type="button" aria-label="New folder" @click="folderSheetOpen = true">
        + Folder
      </button>
      <button
        v-if="searchResults"
        class="btn ghost"
        type="button"
        @click="searchResults = null; searchQuery = ''"
      >
        Clear
      </button>
    </div>

    <p v-if="uploadProgress !== null" class="muted upload-status">
      Uploading… {{ Math.round(uploadProgress * 100) }}%
    </p>
    <p v-if="error" class="error">{{ error }}</p>
    <LoadingSkeletonFiles v-if="loading && !searchResults" mode="browse" />
    <LoadingSkeletonFiles v-else-if="searchLoading" mode="search" />

    <section v-if="!loading && searchResults" class="list">
      <h2 class="section-title">Search results</h2>
      <div v-for="folder in searchResults.folders" :key="folder.id" class="row">
        <span class="name"><Icon name="folder" :size="18" class="row-icon" />{{ folder.name }}</span>
        <button class="btn icon-only" type="button" aria-label="Open folder" @click="openFolder(folder.id)">
          →
        </button>
      </div>
      <div v-for="file in searchResults.files" :key="file.id" class="row">
        <span class="name"><Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />{{ file.name }}</span>
        <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
        <button class="btn icon-only" type="button" aria-label="File actions" @click="openFileActions(file)">
          ⋯
        </button>
      </div>
      <EmptyState
        v-if="searchResults.folders.length === 0 && searchResults.files.length === 0"
        title="No results"
        description="Try a different search term."
        icon="file"
      />
    </section>

    <section v-else-if="!loading" class="list">
      <div
        v-for="folder in browser?.folders ?? []"
        :key="folder.id"
        class="row tappable"
        @click="openFolder(folder.id)"
      >
        <span class="name"><Icon name="folder" :size="18" class="row-icon" />{{ folder.name }}</span>
        <button
          class="btn icon-only"
          type="button"
          aria-label="Folder actions"
          @click.stop="openFolderActions(folder)"
        >
          ⋯
        </button>
      </div>
      <div v-for="file in browser?.files ?? []" :key="file.id" class="row">
        <span class="name"><Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />{{ file.name }}</span>
        <span class="meta desktop-only">{{ mimeLabel(file.mimeType) }} · {{ formatBytes(file.sizeBytes) }}</span>
        <button class="btn icon-only" type="button" aria-label="File actions" @click="openFileActions(file)">
          ⋯
        </button>
      </div>
      <EmptyState
        v-if="isEmpty"
        title="No files here"
        description="Upload a file or create a folder to get started."
        action-label="Upload file"
        icon="folder"
        @action="triggerUpload"
      />
    </section>

    <input ref="fileInputRef" type="file" class="sr-only" @change="onUploadChange" />

    <UploadFab label="Upload file" :disabled="uploadProgress !== null" @click="triggerUpload" />

    <BottomSheet :open="folderSheetOpen" title="New folder" @close="folderSheetOpen = false">
      <label class="field">
        <span>Folder name</span>
        <input v-model="newFolderName" type="text" @keyup.enter="createFolder" />
      </label>
      <button type="button" class="btn block ink" @click="createFolder">Create folder</button>
    </BottomSheet>

    <FolderPickerSheet
      :open="pickerOpen"
      :title="pickerTitle"
      :exclude-folder-id="pickerExcludeFolderId"
      confirm-label="Move here"
      @select="onPickerSelect"
      @close="pickerOpen = false"
    />
  </div>
</template>

<style scoped>
.files-page {
  position: relative;
}

.row-icon {
  flex-shrink: 0;
  color: var(--muted);
  vertical-align: -0.2em;
  margin-right: var(--space-xs);
}

.upload-status {
  margin-bottom: var(--space-sm);
}

.toolbar-sticky {
  position: sticky;
  top: 0;
  z-index: 10;
  padding: var(--space-xs) 0 var(--space-sm);
  margin-bottom: var(--space-md);
  background: var(--surface-soft);
}

.breadcrumb {
  align-items: center;
}

.mobile-back {
  min-height: auto;
  padding: 0.25rem 0.5rem;
  margin-right: var(--space-xs);
}

.mobile-current {
  font-weight: 600;
  color: var(--ink);
}

.breadcrumb-trail {
  display: none;
}

.mobile-only {
  display: inline-flex;
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .mobile-back,
  .mobile-current,
  .mobile-only {
    display: none;
  }

  .breadcrumb-trail {
    display: inline;
  }

  .desktop-only {
    display: inline-flex;
  }

  .page-title.desktop-only {
    display: block;
  }

  .toolbar-sticky {
    background: transparent;
    padding-top: 0;
  }
}
</style>
