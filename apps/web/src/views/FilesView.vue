<script setup lang="ts">
import { computed, inject, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, formatBytes, uploadToPresigned } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import UploadFab from '@/components/UploadFab.vue'
import UploadProgress from '@/components/UploadProgress.vue'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import FolderPickerSheet from '@/components/FolderPickerSheet.vue'
import LoadingSkeletonFiles from '@/components/LoadingSkeletonFiles.vue'
import { mimeIcon, mimeLabel, resolveContentType } from '@/lib/mimeIcon'
import type { Browser, DownloadURL, SearchFilters, SearchResult, UploadSession } from '@/api/types'
import SearchFilterSheet from '@/components/SearchFilterSheet.vue'

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
const filterSheetOpen = ref(false)
const filters = ref<SearchFilters>({ type: 'all', sort: 'relevance', order: 'desc' })
const newFolderName = ref('')
const folderSheetOpen = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const folderInputRef = ref<HTMLInputElement | null>(null)
const dragDepth = ref(0)
const isDragging = computed(() => dragDepth.value > 0)

function onDragEnter(event: DragEvent) {
  if (!event.dataTransfer?.types.includes('Files')) return
  dragDepth.value += 1
}

function onDragOver(event: DragEvent) {
  if (!event.dataTransfer?.types.includes('Files')) return
  event.preventDefault()
  event.dataTransfer.dropEffect = 'copy'
}

function onDragLeave() {
  dragDepth.value = Math.max(0, dragDepth.value - 1)
}

function onDrop(event: DragEvent) {
  dragDepth.value = 0
  const files = Array.from(event.dataTransfer?.files ?? [])
  if (files.length === 0) return
  event.preventDefault()
  void uploadFiles(files)
}

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

const hasActiveFilters = computed(
  () =>
    filters.value.type !== 'all' ||
    filters.value.sort !== 'relevance' ||
    filters.value.order !== 'desc' ||
    Boolean(filters.value.folderId) ||
    Boolean(filters.value.from) ||
    Boolean(filters.value.to),
)

const TYPE_LABELS: Record<SearchFilters['type'], string> = {
  all: 'All',
  image: 'Images',
  video: 'Videos',
  document: 'Documents',
  archive: 'Archives',
  folder: 'Folders',
}

const SORT_LABELS: Record<SearchFilters['sort'], string> = {
  relevance: 'Best match',
  name: 'Name',
  date: 'Date',
  size: 'Size',
}

const activeFilterChips = computed(() => {
  const f = filters.value
  const chips: Array<{ key: keyof SearchFilters; label: string }> = []
  if (f.type !== 'all') chips.push({ key: 'type', label: TYPE_LABELS[f.type] })
  if (f.folderId) chips.push({ key: 'folderId', label: 'This folder' })
  if (f.from) chips.push({ key: 'from', label: `From ${f.from}` })
  if (f.to) chips.push({ key: 'to', label: `To ${f.to}` })
  if (f.sort !== 'relevance') {
    chips.push({ key: 'sort', label: `Sort: ${SORT_LABELS[f.sort]}` })
  }
  if (f.sort !== 'relevance' && f.order !== 'desc') {
    chips.push({ key: 'order', label: 'Ascending' })
  }
  return chips
})

function buildSearchQuery(): string {
  const params = new URLSearchParams()
  params.set('q', searchQuery.value.trim())
  const f = filters.value
  if (f.type !== 'all') params.set('type', f.type)
  if (f.folderId) params.set('folderId', f.folderId)
  if (f.from) params.set('from', f.from)
  if (f.to) params.set('to', f.to)
  if (f.sort !== 'relevance') params.set('sort', f.sort)
  if (f.order !== 'desc') params.set('order', f.order)
  return params.toString()
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
    searchResults.value = await api<SearchResult>(`/search?${buildSearchQuery()}`)
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

watch(filters, runSearch)

function onFiltersApply(applied: SearchFilters) {
  filterSheetOpen.value = false
  filters.value = applied
  if (searchTimer) clearTimeout(searchTimer)
  void runSearch()
}

function clearFilters() {
  filters.value = { type: 'all', sort: 'relevance', order: 'desc' }
}

function removeFilter(key: keyof SearchFilters) {
  const next = { ...filters.value }
  if (key === 'type') next.type = 'all'
  else if (key === 'sort') next.sort = 'relevance'
  else if (key === 'order') next.order = 'desc'
  else delete next[key]
  filters.value = next
}

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

function triggerFolderUpload() {
  folderInputRef.value?.click()
}

async function uploadOneFile(file: File, targetFolderId: string | null) {
  const contentType = resolveContentType(file)
  if (!contentType) {
    ui.showToast(`Skipped "${file.name}" (unsupported type)`, 'info')
    return
  }
  const session = await api<UploadSession>('/files/upload-sessions', {
    method: 'POST',
    body: JSON.stringify({
      name: file.name,
      size: file.size,
      contentType,
      folderId: targetFolderId,
    }),
  })
  await uploadToPresigned(session.uploadUrl, file, contentType)
  await api(`/files/${session.fileId}/complete`, { method: 'POST', body: '{}' })
}

async function uploadFiles(files: File[]) {
  if (files.length === 0) return
  error.value = ''
  uploadProgress.value = 0
  let done = 0
  try {
    for (const file of files) {
      await uploadOneFile(file, folderId.value)
      done += 1
      uploadProgress.value = done / files.length
    }
    ui.showToast(files.length === 1 ? 'Upload complete' : `${files.length} files uploaded`)
    await loadBrowser()
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, 'Upload failed')
  } finally {
    uploadProgress.value = null
  }
}

async function onUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''
  await uploadFiles(files)
}

async function onFolderUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''
  if (files.length === 0) return
  error.value = ''
  uploadProgress.value = 0
  let done = 0
  try {
    const folderCache = new Map<string, string>()
    for (const file of files) {
      const relPath = (file as File & { webkitRelativePath?: string }).webkitRelativePath ?? file.name
      const parts = relPath.split('/')
      parts.pop()
      let parentId = folderId.value
      if (parts.length > 0) {
        parentId = await ensureFolderPath(parts, parentId, folderCache)
      }
      await uploadOneFile(file, parentId)
      done += 1
      uploadProgress.value = done / files.length
    }
    ui.showToast(files.length === 1 ? 'Upload complete' : `${files.length} files uploaded`)
    await loadBrowser()
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, 'Upload failed')
  } finally {
    uploadProgress.value = null
  }
}

async function ensureFolderPath(
  parts: string[],
  rootParentId: string | null,
  cache: Map<string, string>,
): Promise<string | null> {
  let parentId = rootParentId
  let currentPath = ''
  for (const part of parts) {
    currentPath = currentPath ? `${currentPath}/${part}` : part
    const cached = cache.get(currentPath)
    if (cached) {
      parentId = cached
      continue
    }
    const created = await api<{ id: string }>('/folders/get-or-create', {
      method: 'POST',
      body: JSON.stringify({ name: part, parentId }),
    })
    parentId = created.id
    cache.set(currentPath, parentId)
  }
  return parentId
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

/** Optimistic UI: drop the row immediately so TransitionGroup animates the removal. */
function removeOptimistic(id: string) {
  const b = browser.value
  if (b) {
    b.folders = b.folders.filter((folder) => folder.id !== id)
    b.files = b.files.filter((file) => file.id !== id)
  }
  const s = searchResults.value
  if (s) {
    s.folders = s.folders.filter((folder) => folder.id !== id)
    s.files = s.files.filter((file) => file.id !== id)
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
  removeOptimistic(id)
  try {
    await api(`/files/${id}`, { method: 'DELETE' })
    ui.showToast('Moved to trash')
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
    await loadBrowser()
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
  removeOptimistic(id)
  try {
    await api(`/folders/${id}`, { method: 'DELETE' })
    ui.showToast('Moved to trash')
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
    await loadBrowser()
  }
}

async function openFolderActions(folder: { id: string; name: string }) {
  const action = await ui.openActionSheet(folder.name, [
    { id: 'open', label: 'Open', icon: 'arrow-right' },
    { id: 'rename', label: 'Rename', icon: 'pencil' },
    { id: 'move', label: 'Move', icon: 'move' },
    { id: 'delete', label: 'Move to trash', icon: 'trash', danger: true },
  ])
  if (action === 'open') await openFolder(folder.id)
  if (action === 'rename') await renameFolder(folder.id, folder.name)
  if (action === 'move') openMoveFolder(folder.id)
  if (action === 'delete') await deleteFolder(folder.id)
}

async function openFileActions(file: { id: string; name: string }) {
  const action = await ui.openActionSheet(file.name, [
    { id: 'download', label: 'Download', icon: 'download' },
    { id: 'rename', label: 'Rename', icon: 'pencil' },
    { id: 'move', label: 'Move', icon: 'move' },
    { id: 'delete', label: 'Move to trash', icon: 'trash', danger: true },
  ])
  if (action === 'download') await downloadFile(file.id)
  if (action === 'rename') await renameFile(file.id, file.name)
  if (action === 'move') openMoveFile(file.id)
  if (action === 'delete') await deleteFile(file.id)
}

watch(() => route.query.folderId, loadBrowser, { immediate: true })
</script>

<template>
  <div
    class="files-page"
    @dragenter="onDragEnter"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <Transition name="page">
      <div v-if="isDragging" class="drop-overlay" aria-hidden="true">
        <div class="drop-card">
          <span class="drop-icon">⬆</span>
          <p class="drop-label">Drop files to upload</p>
        </div>
      </div>
    </Transition>
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
        aria-label="Search files"
        enterkeyhint="search"
      />
      <button
        class="btn icon-only filter-toggle"
        :class="{ 'filter-active': hasActiveFilters }"
        type="button"
        aria-label="Search filters"
        @click="filterSheetOpen = true"
      >
        <Icon name="filter" :size="18" />
      </button>
      <button class="btn accent desktop-only" type="button" @click="triggerUpload">Upload</button>
      <button class="btn desktop-only" type="button" @click="triggerFolderUpload">Upload folder</button>
      <button class="btn desktop-only" type="button" @click="folderSheetOpen = true">New folder</button>
      <button class="btn mobile-only" type="button" aria-label="New folder" @click="folderSheetOpen = true">
        New folder
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

    <UploadProgress :progress="uploadProgress" />
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonFiles v-if="loading && !searchResults" mode="browse" />
    <LoadingSkeletonFiles v-else-if="searchLoading" mode="search" />

    <TransitionGroup v-if="!loading && searchResults" name="row" tag="section" class="list">
      <h2 key="search-title" class="section-title">
        {{ searchResults.folders.length + searchResults.files.length }} results
        <template v-if="hasActiveFilters"> · filtered</template>
      </h2>
      <div v-if="hasActiveFilters" key="filter-chips" class="filter-chips">
        <span v-for="chip in activeFilterChips" :key="chip.key" class="filter-chip">
          {{ chip.label }}
          <button type="button" class="chip-remove" aria-label="Remove filter" @click="removeFilter(chip.key)">
            ×
          </button>
        </span>
        <button type="button" class="chip-clear" @click="clearFilters">Clear all</button>
      </div>
      <div
        v-for="folder in searchResults.folders"
        :key="folder.id"
        class="row tappable"
        @click="openFolder(folder.id)"
      >
        <span class="name"><Icon name="folder" :size="18" class="row-icon" />{{ folder.name }}</span>
        <button class="btn icon-only" type="button" aria-label="Open folder" @click.stop="openFolder(folder.id)">
          <Icon name="arrow-right" :size="18" />
        </button>
      </div>
      <div
        v-for="file in searchResults.files"
        :key="file.id"
        class="row tappable"
        @click="openFileActions(file)"
      >
        <span class="name"><Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />{{ file.name }}</span>
        <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
        <button class="btn icon-only" type="button" aria-label="File actions" @click.stop="openFileActions(file)">
          <Icon name="more" :size="18" />
        </button>
      </div>
      <EmptyState
        v-if="searchResults.folders.length === 0 && searchResults.files.length === 0"
        key="search-empty"
        :title="hasActiveFilters ? 'No results match your filters' : 'No results'"
        :description="hasActiveFilters ? 'Try removing some filters.' : 'Try a different search term.'"
        icon="file"
      >
        <button v-if="hasActiveFilters" type="button" class="btn" @click="clearFilters">Clear filters</button>
      </EmptyState>
    </TransitionGroup>

    <TransitionGroup v-else-if="!loading" name="row" tag="section" class="list">
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
          <Icon name="more" :size="18" />
        </button>
      </div>
      <div v-for="file in browser?.files ?? []" :key="file.id" class="row tappable" @click="openFileActions(file)">
        <span class="name"><Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />{{ file.name }}</span>
        <span class="meta desktop-only">{{ mimeLabel(file.mimeType) }} · {{ formatBytes(file.sizeBytes) }}</span>
        <button class="btn icon-only" type="button" aria-label="File actions" @click.stop="openFileActions(file)">
          <Icon name="more" :size="18" />
        </button>
      </div>
      <EmptyState
        v-if="isEmpty"
        key="browse-empty"
        title="No files here"
        description="Upload a file or create a folder to get started."
        action-label="Upload file"
        icon="folder"
        @action="triggerUpload"
      />
    </TransitionGroup>

    <input ref="fileInputRef" type="file" class="sr-only" multiple @change="onUploadChange" />
    <input ref="folderInputRef" type="file" class="sr-only" webkitdirectory @change="onFolderUploadChange" />

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

    <SearchFilterSheet
      :open="filterSheetOpen"
      :filters="filters"
      @apply="onFiltersApply"
      @close="filterSheetOpen = false"
    />
  </div>
</template>

<style scoped>
.files-page {
  position: relative;
  padding-bottom: calc(var(--fab-size) + var(--space-md));
}

.drop-overlay {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-md);
  background: var(--overlay);
}

.drop-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xl) var(--space-xl);
  border: 2px dashed var(--accent);
  border-radius: var(--radius-xl);
  background: var(--canvas);
  color: var(--ink);
}

.drop-icon {
  font-size: 32px;
  line-height: 1;
  color: var(--accent);
}

.drop-label {
  margin: 0;
  font-weight: 600;
}

.row-icon {
  flex-shrink: 0;
  color: var(--muted);
  vertical-align: -0.2em;
  margin-right: var(--space-xs);
}

.toolbar-sticky {
  position: sticky;
  top: 0;
  z-index: 10;
  padding: var(--space-xs) 0 var(--space-sm);
  margin-bottom: var(--space-md);
  background: var(--surface-soft);
}

.toolbar-sticky .search-input {
  flex: 1 1 100%;
}

.filter-toggle {
  border: 1px solid var(--hairline);
}

.filter-toggle.filter-active {
  border-color: var(--accent);
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
}

.filter-chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-xs);
  margin-bottom: var(--space-sm);
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xxs);
  min-height: 28px;
  padding: 0 var(--space-xs);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  font-size: 0.8125rem;
  color: var(--ink);
}

.chip-remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--muted);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
}

.chip-remove:hover {
  background: var(--surface-soft);
  color: var(--ink);
}

.chip-clear {
  min-height: 28px;
  padding: 0 var(--space-xs);
  border: none;
  background: transparent;
  color: var(--accent);
  font-weight: 600;
  cursor: pointer;
}

.breadcrumb {
  align-items: center;
}

.mobile-back {
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
  margin-right: var(--space-xs);
}

.mobile-current {
  font-weight: 600;
  color: var(--ink);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  .files-page {
    padding-bottom: 0;
  }

  .toolbar-sticky .search-input {
    flex: 1 1 auto;
  }

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
