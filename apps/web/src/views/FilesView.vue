<script setup lang="ts">
import { computed, inject, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useFileUploadQueue } from '@/lib/useFileUploadQueue'
import { useUiStore } from '@/stores/ui'
import UploadFab from '@/components/UploadFab.vue'
import UploadProgress from '@/components/UploadProgress.vue'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import FolderPickerSheet from '@/components/FolderPickerSheet.vue'
import ShareSheet from '@/components/ShareSheet.vue'
import ShareUserSheet from '@/components/ShareUserSheet.vue'
import LoadingSkeletonFiles from '@/components/LoadingSkeletonFiles.vue'
import BatchActionBar from '@/components/BatchActionBar.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import { usePullToRefresh } from '@/lib/usePullToRefresh'
import { useLongPress } from '@/lib/useLongPress'
import { mimeIcon } from '@/lib/mimeIcon'
import { mimeCategoryColor } from '@/lib/mimeColors'
import type {
  Browser,
  DownloadURL,
  FavoriteFile,
  SearchFilters,
  SearchResult,
  ShareLinkInfo,
  ShareLinkTTL,
} from '@/api/types'
import SearchFilterSheet from '@/components/SearchFilterSheet.vue'
import { useI18n } from '@/lib/i18n'
import { vaultMoveCopy } from '@/lib/vaultMoveCopy'
import { filesOperationsCopy } from '@/lib/filesOperationsCopy'
import {
  buildFilesRouteQuery,
  buildSearchApiQueryString,
  getActiveFilterChipKeys,
  getEmptySearchCause,
  hasActiveSearchFilters,
  parseSearchFiltersFromRoute,
  parseSearchQueryFromRoute,
  removeSearchFilter,
  resetSearchFilters,
  routeQueriesEqual,
  type FilesSearchChipKey,
} from '@/lib/filesSearchState'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const { locale, t } = useI18n()
const vaultCopy = computed(() => vaultMoveCopy(locale.value))
const operationsCopy = computed(() => filesOperationsCopy(locale.value))
const reloadStorage = inject<() => Promise<void>>('reloadStorage')

const browser = ref<Browser | null>(null)
const searchResults = ref<SearchResult | null>(null)
const loading = ref(false)
const searchLoading = ref(false)
const error = ref('')
const searchQuery = ref(parseSearchQueryFromRoute(route.query))
const {
  aggregateProgress: uploadAggregateProgress,
  progressItems: uploadProgressItems,
  enqueueFiles: enqueueUploadFiles,
  retryUpload,
  cancelUpload,
  dismissFailedUpload,
  dismissUploadPanel,
} = useFileUploadQueue({
  getRootFolderId: () => folderId.value,
  onBatchSettled: async () => {
    await loadBrowser()
    await reloadStorage?.()
  },
  showToast: (message, tone) => ui.showToast(message, tone),
})
const filterSheetOpen = ref(false)
const filters = ref<SearchFilters>(parseSearchFiltersFromRoute(route.query))
let syncingFromRoute = false
const newFolderName = ref('')
const folderSheetOpen = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const folderInputRef = ref<HTMLInputElement | null>(null)
const dragDepth = ref(0)
const isDragging = computed(() => dragDepth.value > 0)
const segment = ref<'all' | 'favorites'>(route.query.view === 'favorites' ? 'favorites' : 'all')
const favorites = ref<FavoriteFile[] | null>(null)
const favoritesLoading = ref(false)

const filesPageRef = ref<HTMLElement | null>(null)
const { pullDistance, isRefreshing, attachListeners } = usePullToRefresh(filesPageRef, {
  onRefresh: loadBrowser,
})

// Multi-select state
const isSelecting = ref(false)
const selectedFileIds = ref<Set<string>>(new Set())
const selectedFolderIds = ref<Set<string>>(new Set())
const totalSelectedCount = computed(() => selectedFileIds.value.size + selectedFolderIds.value.size)
const selectedFileCount = computed(() => selectedFileIds.value.size)
const selectedFolderCount = computed(() => selectedFolderIds.value.size)
const totalItemsCount = computed(
  () => (browser.value?.folders.length ?? 0) + (browser.value?.files.length ?? 0),
)

function startSelection(type: 'file' | 'folder', id: string) {
  isSelecting.value = true
  if (type === 'file') selectedFileIds.value.add(id)
  else selectedFolderIds.value.add(id)
}

function toggleSelectItem(type: 'file' | 'folder', id: string) {
  const set = type === 'file' ? selectedFileIds.value : selectedFolderIds.value
  if (set.has(id)) {
    set.delete(id)
    if (totalSelectedCount.value === 0) {
      isSelecting.value = false
    }
  } else {
    set.add(id)
  }
}

function selectAllItems() {
  if (browser.value) {
    selectedFileIds.value = new Set(browser.value.files.map((f) => f.id))
    selectedFolderIds.value = new Set(browser.value.folders.map((f) => f.id))
  }
}

function clearSelection() {
  selectedFileIds.value.clear()
  selectedFolderIds.value.clear()
  isSelecting.value = false
}

function enterSelectionMode() {
  isSelecting.value = true
}

function toggleSelectionMode() {
  if (isSelecting.value) clearSelection()
  else enterSelectionMode()
}

function onItemClick(event: MouseEvent, type: 'file' | 'folder', id: string) {
  if (shouldIgnoreClick) return
  const withModifier = event.ctrlKey || event.metaKey
  if (isSelecting.value || withModifier) {
    event.preventDefault()
    if (!isSelecting.value) startSelection(type, id)
    else toggleSelectItem(type, id)
    return
  }
  if (type === 'folder') void openFolder(id)
}

function onFileItemClick(event: MouseEvent, file: { id: string; name: string; mimeType?: string }) {
  if (shouldIgnoreClick) return
  const withModifier = event.ctrlKey || event.metaKey
  if (isSelecting.value || withModifier) {
    event.preventDefault()
    if (!isSelecting.value) startSelection('file', file.id)
    else toggleSelectItem('file', file.id)
    return
  }
  void openFileActions(file)
}

function onItemKeydown(event: KeyboardEvent, type: 'file' | 'folder', id: string) {
  if (event.key === ' ' || event.key === 'Spacebar') {
    event.preventDefault()
    if (!isSelecting.value) startSelection(type, id)
    else toggleSelectItem(type, id)
    return
  }
  if (event.key === 'Enter') {
    if (isSelecting.value) {
      event.preventDefault()
      toggleSelectItem(type, id)
      return
    }
    if (type === 'folder') {
      event.preventDefault()
      void openFolder(id)
    }
  }
  if (event.key === 'Escape' && isSelecting.value) {
    event.preventDefault()
    clearSelection()
  }
}

function onFileItemKeydown(
  event: KeyboardEvent,
  file: { id: string; name: string; mimeType?: string },
) {
  if (event.key === ' ' || event.key === 'Spacebar') {
    event.preventDefault()
    if (!isSelecting.value) startSelection('file', file.id)
    else toggleSelectItem('file', file.id)
    return
  }
  if (event.key === 'Enter') {
    if (isSelecting.value) {
      event.preventDefault()
      toggleSelectItem('file', file.id)
      return
    }
    event.preventDefault()
    void openFileActions(file)
  }
  if (event.key === 'Escape' && isSelecting.value) {
    event.preventDefault()
    clearSelection()
  }
}

// Long-press detection
const {
  start: startLongPress,
  move: moveLongPress,
  end: endLongPress,
  cancel: cancelLongPress,
  shouldIgnoreClick,
} = useLongPress({
  onLongPress: (payload) => {
    if (payload && typeof payload === 'object' && 'type' in payload && 'id' in payload) {
      const p = payload as { type: 'file' | 'folder'; id: string }
      startSelection(p.type, p.id)
    }
  },
})

// Mobile breadcrumb sheet
const breadcrumbSheetOpen = ref(false)

const sortBy = ref<'updatedAt' | 'name' | 'size'>('updatedAt')
const sortOrder = ref<'asc' | 'desc'>('desc')
const viewMode = ref<'list' | 'grid'>(
  (localStorage.getItem('filvault.filesViewMode') as 'list' | 'grid') || 'list',
)

function toggleViewMode() {
  viewMode.value = viewMode.value === 'list' ? 'grid' : 'list'
  localStorage.setItem('filvault.filesViewMode', viewMode.value)
}

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

const currentSortLabel = computed(() => {
  if (sortBy.value === 'name') return t.value.sortByName
  if (sortBy.value === 'size') return t.value.sortBySize
  return t.value.sortByTime
})

async function openSortMenu() {
  const choice = await ui.openActionSheet(t.value.sortFilesAria, [
    { id: 'updatedAt', label: t.value.sortByTime, icon: 'sort' },
    { id: 'name', label: t.value.sortByName, icon: 'doc' },
    { id: 'size', label: t.value.sortBySize, icon: 'archive' },
  ])
  if (choice) {
    if (sortBy.value === choice) {
      sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      sortBy.value = choice as 'updatedAt' | 'name' | 'size'
      sortOrder.value = choice === 'name' ? 'asc' : 'desc'
    }
  }
}

const sortedFolders = computed(() => {
  if (!browser.value?.folders) return []
  const list = [...browser.value.folders]
  list.sort((a, b) => {
    if (sortBy.value === 'name') {
      const cmp = a.name.localeCompare(b.name)
      return sortOrder.value === 'asc' ? cmp : -cmp
    }
    const cmp = (a.updatedAt || '').localeCompare(b.updatedAt || '')
    return sortOrder.value === 'asc' ? cmp : -cmp
  })
  return list
})

const sortedFiles = computed(() => {
  if (!browser.value?.files) return []
  const list = [...browser.value.files]
  list.sort((a, b) => {
    if (sortBy.value === 'name') {
      const cmp = a.name.localeCompare(b.name)
      return sortOrder.value === 'asc' ? cmp : -cmp
    }
    if (sortBy.value === 'size') {
      const cmp = (a.sizeBytes || 0) - (b.sizeBytes || 0)
      return sortOrder.value === 'asc' ? cmp : -cmp
    }
    const cmp = (a.updatedAt || '').localeCompare(b.updatedAt || '')
    return sortOrder.value === 'asc' ? cmp : -cmp
  })
  return list
})

async function onFabClick() {
  const action = await ui.openActionSheet(t.value.upload, [
    { id: 'file', label: t.value.upload, icon: 'upload' },
    { id: 'folder-upload', label: t.value.uploadFolder, icon: 'folder' },
    { id: 'new-folder', label: t.value.newFolder, icon: 'plus' },
  ])
  if (action === 'file') triggerUpload()
  else if (action === 'folder-upload') triggerFolderUpload()
  else if (action === 'new-folder') folderSheetOpen.value = true
}

// In-app preview
const previewOpen = ref(false)
const previewFile = ref<{ id: string; name: string; mimeType: string; url: string } | null>(null)

async function previewMediaFile(file: { id: string; name: string; mimeType: string }) {
  try {
    const out = await api<DownloadURL>(`/files/${file.id}/download`)
    previewFile.value = {
      id: file.id,
      name: file.name,
      mimeType: file.mimeType,
      url: out.downloadUrl,
    }
    previewOpen.value = true
  } catch (e) {
    error.value = formatApiError(e, 'Could not preview file')
  }
}

onMounted(() => {
  if (filesPageRef.value) attachListeners(filesPageRef.value)
  if (searchQuery.value.trim()) void runSearch()
})

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
const pickerTitle = computed(() => {
  if (pickerTargetId.value === 'batch')
    return operationsCopy.value.moveBatchTitle(totalSelectedCount.value)
  if (pickerMode.value === 'folder') return operationsCopy.value.moveFolderTitle
  return operationsCopy.value.moveFileTitle
})

const folderId = computed(() => {
  const raw = route.query.folderId
  return typeof raw === 'string' && raw ? raw : null
})

const pickerExcludeFolderId = computed(() =>
  pickerMode.value === 'folder' ? pickerTargetId.value : null,
)

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

const hasActiveFilters = computed(() => hasActiveSearchFilters(filters.value))

function filterChipLabel(key: FilesSearchChipKey): string {
  const f = filters.value
  const sortLabels: Record<SearchFilters['sort'], string> = {
    relevance: t.value.searchSortRelevance,
    name: t.value.searchSortName,
    date: t.value.searchSortDate,
    size: t.value.searchSortSize,
  }
  const typeLabels: Record<SearchFilters['type'], string> = {
    all: t.value.all,
    image: t.value.filterTypeImage,
    video: t.value.filterTypeVideo,
    document: t.value.filterTypeDocument,
    archive: t.value.filterTypeArchive,
    folder: t.value.filterTypeFolder,
  }
  switch (key) {
    case 'type':
      return typeLabels[f.type]
    case 'folderId':
      return t.value.filterThisFolder
    case 'from':
      return t.value.filterChipFrom.replace('{date}', f.from ?? '')
    case 'to':
      return t.value.filterChipTo.replace('{date}', f.to ?? '')
    case 'sort':
      return t.value.filterChipSort.replace('{sort}', sortLabels[f.sort])
    case 'order':
      return t.value.filterChipOrderAsc
    default:
      return ''
  }
}

const activeFilterChips = computed(() =>
  getActiveFilterChipKeys(filters.value).map((key) => ({
    key,
    label: filterChipLabel(key),
  })),
)

const emptySearchCause = computed(() => {
  if (!searchResults.value) return null
  if (searchResults.value.folders.length > 0 || searchResults.value.files.length > 0) return null
  return getEmptySearchCause(searchQuery.value, filters.value)
})

const emptySearchTitle = computed(() => {
  const cause = emptySearchCause.value
  if (cause === 'both') return t.value.noResultsQueryAndFilters
  if (cause === 'filters') return t.value.noResultsFiltered
  return t.value.noResults
})

const emptySearchDesc = computed(() => {
  const cause = emptySearchCause.value
  if (cause === 'both') return t.value.noResultsQueryAndFiltersDesc
  if (cause === 'filters') return t.value.noResultsFilteredDesc
  return t.value.noResultsDesc
})

function syncRouteFromSearchState() {
  if (syncingFromRoute) return
  const nextQuery = buildFilesRouteQuery(route.query, searchQuery.value, filters.value)
  if (!routeQueriesEqual(route.query, nextQuery)) {
    void router.replace({ path: route.path, query: nextQuery })
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
    searchResults.value = await api<SearchResult>(
      `/search?${buildSearchApiQueryString(searchQuery.value, filters.value)}`,
    )
  } catch (e) {
    error.value = formatApiError(e, 'Search failed')
  } finally {
    searchLoading.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(searchQuery, () => {
  if (syncingFromRoute) return
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    syncRouteFromSearchState()
    void runSearch()
  }, 300)
})

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})

watch(filters, () => {
  if (syncingFromRoute) return
  syncRouteFromSearchState()
  void runSearch()
})

watch(
  () => route.query,
  (query) => {
    syncingFromRoute = true
    const nextQ = parseSearchQueryFromRoute(query)
    const nextFilters = parseSearchFiltersFromRoute(query)
    const qChanged = searchQuery.value !== nextQ
    const filtersChanged = JSON.stringify(filters.value) !== JSON.stringify(nextFilters)
    if (qChanged) searchQuery.value = nextQ
    if (filtersChanged) filters.value = nextFilters
    syncingFromRoute = false
    if (qChanged || filtersChanged) {
      if (nextQ.trim()) void runSearch()
      else searchResults.value = null
    }
  },
)

function onFiltersApply(applied: SearchFilters) {
  filterSheetOpen.value = false
  filters.value = applied
  if (searchTimer) clearTimeout(searchTimer)
  void runSearch()
}

function clearFilters() {
  filters.value = resetSearchFilters()
}

function clearSearch() {
  searchQuery.value = ''
  searchResults.value = null
  syncRouteFromSearchState()
}

function removeFilter(key: FilesSearchChipKey) {
  filters.value = removeSearchFilter(filters.value, key)
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

function uploadFiles(files: File[]) {
  if (files.length === 0) return
  error.value = ''
  enqueueUploadFiles(files)
}

async function onUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''
  await uploadFiles(files)
}

function onFolderUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''
  if (files.length === 0) return
  error.value = ''
  enqueueUploadFiles(files, true)
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
  const name = await ui.prompt({
    title: operationsCopy.value.renameFileTitle,
    label: operationsCopy.value.nameLabel,
    initialValue: current,
  })
  if (!name || name === current) return
  error.value = ''
  try {
    await api(`/files/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) })
    ui.showToast(operationsCopy.value.fileRenamed)
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, operationsCopy.value.renameFailed)
  }
}

async function renameFolder(id: string, current: string) {
  const name = await ui.prompt({
    title: operationsCopy.value.renameFolderTitle,
    label: operationsCopy.value.nameLabel,
    initialValue: current,
  })
  if (!name || name === current) return
  error.value = ''
  try {
    await api(`/folders/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) })
    ui.showToast(operationsCopy.value.folderRenamed)
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, operationsCopy.value.renameFailed)
  }
}

function openMoveFile(id: string) {
  pickerTargetId.value = id
  pickerMode.value = 'file'
  pickerOpen.value = true
}

function openMoveFolder(id: string) {
  pickerTargetId.value = id
  pickerMode.value = 'folder'
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
    if (id === 'batch') {
      for (const fId of selectedFileIds.value) {
        await api(`/files/${fId}`, {
          method: 'PATCH',
          body: JSON.stringify({ folderId: targetFolderId }),
        })
      }
      for (const dId of selectedFolderIds.value) {
        await api(`/folders/${dId}`, {
          method: 'PATCH',
          body: JSON.stringify({ parentId: targetFolderId }),
        })
      }
      ui.showToast(operationsCopy.value.movedBatch(totalSelectedCount.value))
      clearSelection()
      await loadBrowser()
      return
    }

    if (mode === 'file') {
      await api(`/files/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ folderId: targetFolderId }),
      })
      ui.showToast(operationsCopy.value.fileMoved)
    } else {
      await api(`/folders/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ parentId: targetFolderId }),
      })
      ui.showToast(operationsCopy.value.folderMoved)
    }
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, operationsCopy.value.moveFailed)
  }
}

async function batchDelete() {
  const count = totalSelectedCount.value
  if (count === 0) return
  const ok = await ui.confirm({
    title: operationsCopy.value.trashBatchTitle(count),
    message: operationsCopy.value.trashBatchMessage,
    confirmLabel: t.value.moveToTrash,
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    for (const id of selectedFileIds.value) {
      removeOptimistic(id)
      await api(`/files/${id}`, { method: 'DELETE' })
    }
    for (const id of selectedFolderIds.value) {
      removeOptimistic(id)
      await api(`/folders/${id}`, { method: 'DELETE' })
    }
    ui.showToast(operationsCopy.value.movedBatchToTrash(count))
    clearSelection()
    await loadBrowser()
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, operationsCopy.value.trashFailed)
    await loadBrowser()
  }
}

async function batchFavorite() {
  const fileIds = Array.from(selectedFileIds.value)
  if (fileIds.length === 0) {
    ui.showToast('No files selected to favorite', 'info')
    return
  }
  error.value = ''
  try {
    for (const id of fileIds) {
      await api(`/files/${id}/favorite`, { method: 'PUT' })
    }
    ui.showToast(`Added ${fileIds.length} files to favorites`)
    clearSelection()
    await loadFavorites()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to favorite items')
  }
}

function batchMove() {
  if (totalSelectedCount.value === 0) return
  pickerTargetId.value = 'batch'
  pickerMode.value = 'file'
  pickerOpen.value = true
}

async function batchDownload() {
  const fileIds = Array.from(selectedFileIds.value)
  if (fileIds.length === 0) return
  for (const id of fileIds) {
    await downloadFile(id)
  }
  clearSelection()
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
    title: operationsCopy.value.trashFileTitle,
    message: operationsCopy.value.trashFileMessage,
    confirmLabel: t.value.moveToTrash,
    danger: true,
  })
  if (!ok) return
  error.value = ''
  removeOptimistic(id)
  try {
    await api(`/files/${id}`, { method: 'DELETE' })
    ui.showToast(operationsCopy.value.movedToTrash)
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, operationsCopy.value.trashFailed)
    await loadBrowser()
  }
}

async function deleteFolder(id: string) {
  const ok = await ui.confirm({
    title: operationsCopy.value.trashFolderTitle,
    message: operationsCopy.value.trashFolderMessage,
    confirmLabel: t.value.moveToTrash,
    danger: true,
  })
  if (!ok) return
  error.value = ''
  removeOptimistic(id)
  try {
    await api(`/folders/${id}`, { method: 'DELETE' })
    ui.showToast(operationsCopy.value.movedToTrash)
  } catch (e) {
    error.value = formatApiError(e, operationsCopy.value.trashFailed)
    await loadBrowser()
  }
}

async function openFolderActions(folder: { id: string; name: string }) {
  const action = await ui.openActionSheet(folder.name, [
    { id: 'open', label: t.value.open, icon: 'arrow-right' },
    { id: 'rename', label: t.value.rename, icon: 'pencil' },
    { id: 'move', label: t.value.move, icon: 'move' },
    { id: 'delete', label: t.value.moveToTrash, icon: 'trash', danger: true },
  ])
  if (action === 'open') await openFolder(folder.id)
  if (action === 'rename') await renameFolder(folder.id, folder.name)
  if (action === 'move') openMoveFolder(folder.id)
  if (action === 'delete') await deleteFolder(folder.id)
}

async function loadFavorites() {
  error.value = ''
  favoritesLoading.value = true
  try {
    const out = await api<{ files: FavoriteFile[] }>('/files/favorites')
    favorites.value = out.files
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load favorites')
  } finally {
    favoritesLoading.value = false
  }
}

watch(segment, (next) => {
  if (next === 'favorites' && favorites.value === null) void loadFavorites()
})

watch(
  () => route.query.view,
  (view) => {
    const next = view === 'favorites' ? 'favorites' : 'all'
    if (segment.value !== next) segment.value = next
  },
)

const isFavorited = (id: string) => favorites.value?.some((f) => f.id === id) ?? false

async function toggleFavorite(fileId: string, name: string) {
  const wasFavorited = isFavorited(fileId)
  error.value = ''
  try {
    if (wasFavorited) {
      await api(`/files/${fileId}/favorite`, { method: 'DELETE' })
      ui.showToast(`Removed "${name}" from favorites`)
    } else {
      await api(`/files/${fileId}/favorite`, { method: 'PUT' })
      ui.showToast(`Added "${name}" to favorites`)
    }
    await loadFavorites()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to update favorite')
  }
}

async function openFileActions(file: { id: string; name: string; mimeType?: string }) {
  const favorited = isFavorited(file.id)
  const canPreview = Boolean(
    file.mimeType &&
    (file.mimeType.startsWith('image/') ||
      file.mimeType.startsWith('video/') ||
      file.mimeType === 'application/pdf' ||
      file.mimeType.startsWith('audio/')),
  )
  const action = await ui.openActionSheet(file.name, [
    ...(canPreview ? [{ id: 'preview', label: t.value.preview, icon: 'eye' }] : []),
    { id: 'download', label: t.value.download, icon: 'download' },
    {
      id: 'favorite',
      label: favorited ? t.value.removeFromFavorites : t.value.addToFavorites,
      icon: favorited ? 'star-filled' : 'star',
    },
    { id: 'share', label: t.value.shareLink, icon: 'share' },
    { id: 'share-user', label: t.value.shareWithUser, icon: 'users' },
    { id: 'rename', label: t.value.rename, icon: 'pencil' },
    { id: 'move', label: t.value.move, icon: 'move' },
    { id: 'vault', label: t.value.vaultMoveToVault, icon: 'lock' },
    { id: 'delete', label: t.value.moveToTrash, icon: 'trash', danger: true },
  ])
  if (action === 'preview')
    await previewMediaFile({ id: file.id, name: file.name, mimeType: file.mimeType ?? '' })
  if (action === 'download') await downloadFile(file.id)
  if (action === 'favorite') await toggleFavorite(file.id, file.name)
  if (action === 'share') {
    shareFileName.value = file.name
    await openShareSheet(file.id)
  }
  if (action === 'share-user') {
    shareUserFileName.value = file.name
    shareUserTargetId.value = file.id
    shareUserOpen.value = true
  }
  if (action === 'rename') await renameFile(file.id, file.name)
  if (action === 'move') openMoveFile(file.id)
  if (action === 'vault') await moveFileToVault(file.id, file.name)
  if (action === 'delete') await deleteFile(file.id)
}

async function moveFileToVault(id: string, name: string) {
  const ok = await ui.confirm({
    title: vaultCopy.value.confirmSingleTitle,
    message: vaultCopy.value.confirmSingleMessage(name),
    confirmLabel: vaultCopy.value.confirmSingleLabel,
  })
  if (!ok) return
  error.value = ''
  removeOptimistic(id)
  try {
    await api('/vault/items', {
      method: 'POST',
      body: JSON.stringify({ fileIds: [id] }),
    })
    ui.showToast(vaultCopy.value.moveInSuccess, 'success')
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, vaultCopy.value.moveInFailed)
    await loadBrowser()
  }
}

async function batchMoveToVault() {
  const fileIds = Array.from(selectedFileIds.value)
  if (fileIds.length === 0) return
  const ok = await ui.confirm({
    title: vaultCopy.value.confirmBatchTitle(fileIds.length),
    message: vaultCopy.value.confirmBatchMessage(fileIds.length),
    confirmLabel: vaultCopy.value.confirmBatchLabel,
  })
  if (!ok) return
  error.value = ''
  for (const id of fileIds) {
    removeOptimistic(id)
  }
  try {
    await api('/vault/items', {
      method: 'POST',
      body: JSON.stringify({ fileIds }),
    })
    ui.showToast(vaultCopy.value.moveInBatchSuccess(fileIds.length), 'success')
    clearSelection()
    await loadBrowser()
  } catch (e) {
    error.value = formatApiError(e, vaultCopy.value.moveInFailed)
    await loadBrowser()
  }
}

const shareSheetOpen = ref(false)
const shareFileName = ref('')
const shareTargetId = ref<string | null>(null)
const existingLink = ref<{ url: string; expiresAt: string | null; createdAt: string } | null>(null)

/** The owner API does not expose the token again; reuse the create response
 *  within this session so Copy works without a second request. */
const sessionLinks = new Map<string, { url: string; expiresAt: string | null; createdAt: string }>()

async function openShareSheet(fileId: string) {
  shareTargetId.value = fileId
  const cached = sessionLinks.get(fileId)
  if (cached) {
    existingLink.value = cached
    shareFileName.value = shareFileName.value || ''
  }
  // Ask the list endpoint for expiry info; url/token only exist right after creation.
  try {
    const out = await api<{ links: ShareLinkInfo[] }>('/share-links')
    const link = out.links.find((l) => l.fileId === fileId)
    existingLink.value = cached ?? (link ? sessionLinkFromList(link) : null)
  } catch {
    if (!cached) existingLink.value = null
  }
  shareSheetOpen.value = true
}

function sessionLinkFromList(link: ShareLinkInfo) {
  return {
    url: link.url ?? '',
    expiresAt: link.expiresAt,
    createdAt: link.createdAt,
  }
}

async function createShareLink(ttl: ShareLinkTTL | null) {
  const fileId = shareTargetId.value
  if (!fileId) return
  error.value = ''
  try {
    const created = await api<ShareLinkInfo>(`/files/${fileId}/share`, {
      method: 'POST',
      body: JSON.stringify({ expiresIn: ttl }),
    })
    const entry = {
      url: created.url ?? '',
      expiresAt: created.expiresAt,
      createdAt: created.createdAt,
    }
    sessionLinks.set(fileId, entry)
    existingLink.value = entry
    ui.showToast('Share link created')
  } catch (e) {
    error.value = formatApiError(e, 'Could not create link')
  }
}

async function copyShareLink(url: string) {
  try {
    await navigator.clipboard.writeText(window.location.origin + url)
    ui.showToast('Link copied')
  } catch {
    ui.showToast('Copy failed', 'info')
  }
}

async function revokeShareLink() {
  const fileId = shareTargetId.value
  if (!fileId) return
  const ok = await ui.confirm({
    title: 'Revoke link?',
    message: 'Anyone who has this link will lose access immediately.',
    confirmLabel: 'Revoke link',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/files/${fileId}/share`, { method: 'DELETE' })
    sessionLinks.delete(fileId)
    existingLink.value = null
    ui.showToast('Link revoked')
  } catch (e) {
    error.value = formatApiError(e, 'Could not revoke link')
  }
}

const shareUserOpen = ref(false)
const shareUserFileName = ref('')
const shareUserTargetId = ref<string | null>(null)

function onUserShared(result: { invited: boolean; email: string }) {
  const name = shareUserFileName.value
  if (result.invited) {
    ui.showToast(`Invitation sent to ${result.email}`)
  } else {
    ui.showToast(`Shared "${name}" with ${result.email}`, 'success')
  }
}

watch(() => route.query.folderId, loadBrowser, { immediate: true })
</script>

<template>
  <div
    ref="filesPageRef"
    class="files-page"
    @dragenter="onDragEnter"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <div
      v-if="pullDistance > 0 || isRefreshing"
      class="pull-refresh-bar"
      :style="{ height: `${pullDistance}px` }"
    >
      <span class="pull-icon" :class="{ spin: isRefreshing }">{{ isRefreshing ? '↻' : '↓' }}</span>
    </div>
    <Transition name="page">
      <div v-if="isDragging" class="drop-overlay" aria-hidden="true">
        <div class="drop-card">
          <span class="drop-icon">⬆</span>
          <p class="drop-label">{{ t.dropFilesToUpload }}</p>
        </div>
      </div>
    </Transition>
    <!-- TeraBox Search Bar -->
    <div class="files-search-wrap">
      <Icon name="search" :size="18" class="files-search-icon" />
      <input
        v-model="searchQuery"
        type="search"
        class="files-search-input"
        :placeholder="t.searchInFilvault || 'Tìm kiếm trong Filvault…'"
        :aria-label="t.searchByName"
        enterkeyhint="search"
      />
      <button
        v-if="searchQuery"
        type="button"
        class="files-search-clear"
        :aria-label="t.clearSearchAria"
        @click="clearSearch"
      >
        <Icon name="close" :size="16" />
      </button>
    </div>

    <!-- TeraBox Segmented Tabs -->
    <div class="tabs-header">
      <div class="tabs-pill-list" role="tablist" aria-label="File views">
        <button
          type="button"
          role="tab"
          class="tab-pill"
          :class="{ active: segment === 'all' }"
          :aria-selected="segment === 'all'"
          @click="segment = 'all'"
        >
          {{ t.all }}
        </button>
        <button
          type="button"
          role="tab"
          class="tab-pill"
          :class="{ active: segment === 'favorites' }"
          :aria-selected="segment === 'favorites'"
          @click="segment = 'favorites'"
        >
          {{ t.favorites }}
        </button>
      </div>
    </div>

    <!-- Breadcrumb / Folder Top Bar -->
    <div v-if="segment === 'all' && folderId" class="folder-header-bar">
      <button
        type="button"
        class="folder-back-btn"
        @click="openFolder(browser?.folder?.parentId ?? null)"
      >
        <Icon name="arrow-left" :size="20" />
        <span class="folder-header-title">{{ browser?.folder?.name || t.navFiles }}</span>
      </button>
      <button
        type="button"
        class="folder-hierarchy-btn mobile-only"
        aria-label="View hierarchy"
        @click="breadcrumbSheetOpen = true"
      >
        <Icon name="more" :size="18" />
      </button>
      <span class="breadcrumb-trail desktop-only">
        <a href="#" @click.prevent="openFolder(null)">{{ t.root }}</a>
        <template v-for="item in browser?.breadcrumb ?? []" :key="item.id">
          <span aria-hidden="true">/</span>
          <a href="#" @click.prevent="openFolder(item.id)">{{ item.name }}</a>
        </template>
        <template v-if="browser?.folder">
          <span aria-hidden="true">/</span>
          <span class="current">{{ browser.folder.name }}</span>
        </template>
      </span>
    </div>

    <!-- Sub-Toolbar (Sort + View Mode + Actions) -->
    <div v-if="segment === 'all'" class="files-sub-bar">
      <button
        type="button"
        class="sub-sort-btn"
        :aria-label="t.sortFilesAria"
        @click="openSortMenu"
      >
        <Icon name="sort" :size="16" />
        <span>{{ currentSortLabel }}</span>
        <Icon name="chevron-down" :size="12" />
      </button>
      <div class="sub-actions">
        <button
          type="button"
          class="sub-icon-btn"
          :class="{ 'filter-active': hasActiveFilters }"
          :aria-label="t.searchFiltersAria"
          @click="filterSheetOpen = true"
        >
          <Icon name="sliders" :size="18" />
        </button>
        <button
          type="button"
          class="sub-icon-btn"
          :class="{ 'filter-active': isSelecting }"
          :title="t.selectItems"
          :aria-label="t.selectItems"
          :aria-pressed="isSelecting"
          @click="toggleSelectionMode"
        >
          <Icon name="check" :size="18" />
        </button>
        <button
          type="button"
          class="sub-icon-btn"
          :title="viewMode === 'list' ? t.viewGrid : t.viewList"
          @click="toggleViewMode"
        >
          <Icon :name="viewMode === 'list' ? 'palette' : 'file'" :size="18" />
        </button>
        <button class="btn desktop-only" type="button" @click="folderSheetOpen = true">
          {{ t.newFolder }}
        </button>
        <button class="btn accent desktop-only" type="button" @click="triggerUpload">
          {{ t.upload }}
        </button>
        <button class="btn desktop-only" type="button" @click="triggerFolderUpload">
          {{ t.uploadFolder }}
        </button>
      </div>
    </div>

    <UploadProgress
      :aggregate-progress="uploadAggregateProgress"
      :items="uploadProgressItems"
      :label="t.upload"
      @retry="retryUpload"
      @cancel="cancelUpload"
      @dismiss-failed="dismissFailedUpload"
      @dismiss="dismissUploadPanel"
    />
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonFiles v-if="segment === 'all' && loading && !searchResults" mode="browse" />
    <LoadingSkeletonFiles v-else-if="searchLoading" mode="search" />

    <TransitionGroup
      v-if="segment === 'all' && !loading && searchResults"
      name="row"
      tag="section"
      class="list"
    >
      <h2 key="search-title" class="section-title">
        {{ searchResults.folders.length + searchResults.files.length }} {{ t.results }}
        <template v-if="hasActiveFilters"> · {{ t.filtered }}<!-- · filtered --></template>
      </h2>
      <div v-if="hasActiveFilters" key="filter-chips" class="filter-chips">
        <span v-for="chip in activeFilterChips" :key="chip.key" class="filter-chip">
          {{ chip.label }}
          <button
            type="button"
            class="chip-remove"
            :aria-label="t.removeFilterAria"
            @click="removeFilter(chip.key)"
          >
            ×
          </button>
        </span>
        <button type="button" class="chip-clear" @click="clearFilters">{{ t.clearFilters }}</button>
      </div>
      <div
        v-for="folder in searchResults.folders"
        :key="folder.id"
        class="row tappable"
        @click="openFolder(folder.id)"
      >
        <span class="name"
          ><Icon name="folder" :size="18" class="row-icon" />{{ folder.name }}</span
        >
        <button
          class="btn icon-only"
          type="button"
          aria-label="Open folder"
          @click.stop="openFolder(folder.id)"
        >
          <Icon name="arrow-right" :size="18" />
        </button>
      </div>
      <div
        v-for="file in searchResults.files"
        :key="file.id"
        class="row tappable"
        @click="openFileActions(file)"
      >
        <span class="name"
          ><Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />{{ file.name }}</span
        >
        <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
        <button
          class="btn icon-only"
          type="button"
          aria-label="File actions"
          @click.stop="openFileActions(file)"
        >
          <Icon name="more" :size="18" />
        </button>
      </div>
      <EmptyState
        v-if="searchResults.folders.length === 0 && searchResults.files.length === 0"
        key="search-empty"
        :title="emptySearchTitle"
        :description="emptySearchDesc"
        icon="file"
      >
        <div class="empty-search-actions">
          <button
            v-if="emptySearchCause === 'query' || emptySearchCause === 'both'"
            type="button"
            class="btn"
            @click="clearSearch"
          >
            {{ t.clearSearch }}
          </button>
          <button
            v-if="emptySearchCause === 'filters' || emptySearchCause === 'both'"
            type="button"
            class="btn"
            @click="clearFilters"
          >
            {{ t.clearFilters }}
          </button>
        </div>
      </EmptyState>
    </TransitionGroup>

    <!-- TeraBox Root Vault Shortcut -->
    <RouterLink
      v-if="!folderId && segment === 'all' && !searchResults && !loading"
      to="/vault"
      class="vault-entry-card tappable"
      :aria-label="t.personalVault"
    >
      <div class="vault-icon-box">
        <Icon name="lock" :size="20" aria-hidden="true" />
      </div>
      <div class="vault-info">
        <span class="vault-title">{{ t.personalVault }}</span>
        <span class="vault-hint">{{ t.vaultSubtitle }}</span>
      </div>
      <Icon name="chevron-right" :size="16" class="vault-arrow" aria-hidden="true" />
    </RouterLink>

    <TransitionGroup
      v-if="segment === 'all' && !loading"
      name="row"
      tag="section"
      class="files-container"
      :class="{ 'grid-mode': viewMode === 'grid' }"
    >
      <!-- Folders -->
      <div
        v-for="folder in sortedFolders"
        :key="folder.id"
        class="file-item-card tappable"
        :class="{ selected: selectedFolderIds.has(folder.id) }"
        role="button"
        tabindex="0"
        :aria-pressed="isSelecting ? selectedFolderIds.has(folder.id) : undefined"
        :aria-label="folder.name"
        @touchstart.passive="startLongPress($event, { type: 'folder', id: folder.id })"
        @touchmove.passive="moveLongPress"
        @touchend="endLongPress"
        @touchcancel="cancelLongPress"
        @click="onItemClick($event, 'folder', folder.id)"
        @keydown="onItemKeydown($event, 'folder', folder.id)"
      >
        <span
          v-if="isSelecting"
          class="checkbox-indicator"
          :class="{ checked: selectedFolderIds.has(folder.id) }"
        >
          <Icon v-if="selectedFolderIds.has(folder.id)" name="check" :size="14" />
        </span>
        <div class="file-icon-badge folder-badge">
          <Icon name="folder" :size="22" />
        </div>
        <div class="file-item-info">
          <span class="file-item-title">{{ folder.name }}</span>
          <span class="file-item-sub">{{ formatItemDate(folder.updatedAt) }}</span>
        </div>
        <button
          v-if="!isSelecting"
          class="file-item-more-btn"
          type="button"
          aria-label="Folder actions"
          @click.stop="openFolderActions(folder)"
        >
          <Icon name="more" :size="18" />
        </button>
      </div>

      <!-- Files -->
      <div
        v-for="file in sortedFiles"
        :key="file.id"
        class="file-item-card tappable"
        :class="{ selected: selectedFileIds.has(file.id) }"
        role="button"
        tabindex="0"
        :aria-pressed="isSelecting ? selectedFileIds.has(file.id) : undefined"
        :aria-label="file.name"
        @touchstart.passive="startLongPress($event, { type: 'file', id: file.id })"
        @touchmove.passive="moveLongPress"
        @touchend="endLongPress"
        @touchcancel="cancelLongPress"
        @click="onFileItemClick($event, file)"
        @keydown="onFileItemKeydown($event, file)"
      >
        <span
          v-if="isSelecting"
          class="checkbox-indicator"
          :class="{ checked: selectedFileIds.has(file.id) }"
        >
          <Icon v-if="selectedFileIds.has(file.id)" name="check" :size="14" />
        </span>
        <div
          class="file-icon-badge"
          :style="{
            background: `color-mix(in srgb, ${mimeCategoryColor(file.mimeType, file.name)} 14%, transparent)`,
            color: mimeCategoryColor(file.mimeType, file.name),
          }"
        >
          <Icon :name="isFavorited(file.id) ? 'star-filled' : mimeIcon(file.mimeType)" :size="20" />
        </div>
        <div class="file-item-info">
          <span class="file-item-title">{{ file.name }}</span>
          <span class="file-item-sub"
            >{{ formatItemDate(file.updatedAt) }} · {{ formatBytes(file.sizeBytes) }}</span
          >
        </div>
        <button
          v-if="!isSelecting"
          class="file-item-more-btn"
          type="button"
          aria-label="File actions"
          @click.stop="openFileActions(file)"
        >
          <Icon name="more" :size="18" />
        </button>
      </div>
      <EmptyState
        v-if="isEmpty"
        key="browse-empty"
        :title="t.noFilesHere"
        :description="t.noFilesHereDesc"
        :action-label="t.uploadFile"
        icon="folder"
        @action="triggerUpload"
      />
    </TransitionGroup>

    <TransitionGroup
      v-else-if="segment === 'favorites'"
      name="row"
      tag="section"
      class="files-container"
      :class="{ 'grid-mode': viewMode === 'grid' }"
    >
      <LoadingSkeletonFiles v-if="favoritesLoading" key="fav-skeleton" mode="browse" />
      <template v-else>
        <div
          v-for="file in favorites ?? []"
          :key="file.id"
          class="file-item-card tappable"
          @click="openFileActions(file)"
        >
          <div
            class="file-icon-badge"
            :style="{
              background: `color-mix(in srgb, ${mimeCategoryColor(file.mimeType, file.name)} 14%, transparent)`,
              color: mimeCategoryColor(file.mimeType, file.name),
            }"
          >
            <Icon name="star-filled" :size="20" style="color: #f59e0b" />
          </div>
          <div class="file-item-info">
            <span class="file-item-title">{{ file.name }}</span>
            <span class="file-item-sub"
              >{{ formatItemDate(file.updatedAt) }} · {{ formatBytes(file.sizeBytes) }}</span
            >
          </div>
          <button
            class="file-item-more-btn"
            type="button"
            aria-label="File actions"
            @click.stop="openFileActions(file)"
          >
            <Icon name="more" :size="18" />
          </button>
        </div>
        <EmptyState
          v-if="(favorites ?? []).length === 0"
          key="favorites-empty"
          :title="t.noFavorites"
          :description="t.noFavoritesDesc"
          icon="star"
        />
      </template>
    </TransitionGroup>

    <input ref="fileInputRef" type="file" class="sr-only" multiple @change="onUploadChange" />
    <input
      ref="folderInputRef"
      type="file"
      class="sr-only"
      webkitdirectory
      @change="onFolderUploadChange"
    />

    <UploadFab v-if="segment === 'all' && !isSelecting" :label="t.uploadFile" @click="onFabClick" />

    <BatchActionBar
      v-if="isSelecting"
      :selected-count="totalSelectedCount"
      :selected-file-count="selectedFileCount"
      :selected-folder-count="selectedFolderCount"
      :total-count="totalItemsCount"
      @close="clearSelection"
      @select-all="selectAllItems"
      @clear-selection="clearSelection"
      @delete="batchDelete"
      @move="batchMove"
      @favorite="batchFavorite"
      @vault="batchMoveToVault"
      @download="batchDownload"
    />

    <BottomSheet :open="folderSheetOpen" :title="t.newFolder" @close="folderSheetOpen = false">
      <label class="field">
        <span>{{ t.folderName }}</span>
        <input v-model="newFolderName" type="text" @keyup.enter="createFolder" />
      </label>
      <button type="button" class="btn block ink" @click="createFolder">
        {{ t.createFolder }}
      </button>
    </BottomSheet>

    <BottomSheet
      :open="breadcrumbSheetOpen"
      :title="t.location"
      @close="breadcrumbSheetOpen = false"
    >
      <div class="breadcrumb-sheet-list">
        <button
          type="button"
          class="breadcrumb-sheet-item"
          :class="{ active: !folderId }"
          @click="openFolder(null); breadcrumbSheetOpen = false"
        >
          <Icon name="folder" :size="20" />
          <span>{{ t.root }}</span>
        </button>
        <button
          v-for="item in browser?.breadcrumb ?? []"
          :key="item.id"
          type="button"
          class="breadcrumb-sheet-item"
          @click="openFolder(item.id); breadcrumbSheetOpen = false"
        >
          <Icon name="folder" :size="20" />
          <span>{{ item.name }}</span>
        </button>
        <div v-if="browser?.folder" class="breadcrumb-sheet-item current">
          <Icon name="folder" :size="20" />
          <span
            ><strong>{{ browser.folder.name }}</strong> ({{ t.current }})</span
          >
        </div>
      </div>
    </BottomSheet>

    <FolderPickerSheet
      :open="pickerOpen"
      :title="pickerTitle"
      :exclude-folder-id="pickerExcludeFolderId"
      :confirm-label="t.moveHere"
      @select="onPickerSelect"
      @close="pickerOpen = false"
    />

    <SearchFilterSheet
      :open="filterSheetOpen"
      :filters="filters"
      @apply="onFiltersApply"
      @close="filterSheetOpen = false"
    />

    <ShareSheet
      :open="shareSheetOpen"
      :name="shareFileName"
      :existing="existingLink"
      @create="createShareLink"
      @copy="copyShareLink"
      @revoke="revokeShareLink"
      @close="shareSheetOpen = false"
    />

    <ShareUserSheet
      :open="shareUserOpen"
      :name="shareUserFileName"
      :resource-id="shareUserTargetId"
      @shared="onUserShared"
      @close="shareUserOpen = false"
    />

    <MediaLightbox
      :open="previewOpen"
      :name="previewFile?.name ?? ''"
      :mime-type="previewFile?.mimeType ?? ''"
      :url="previewFile?.url ?? ''"
      @download="previewFile ? downloadFile(previewFile.id) : undefined"
      @close="previewOpen = false"
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

.favorite-icon {
  color: var(--accent);
}

.segment-tabs {
  display: inline-flex;
  gap: var(--space-xxs);
  padding: 3px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  margin-bottom: var(--space-sm);
}

.segment-tab {
  min-height: 32px;
  padding: 0 var(--space-md);
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--muted);
  font-weight: 600;
  cursor: pointer;
  transition:
    background var(--duration-short) var(--ease-standard),
    color var(--duration-short) var(--ease-standard);
}

.segment-tab.active {
  background: var(--accent);
  color: var(--on-accent, #fff);
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

.empty-search-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-xs);
  justify-content: center;
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

.pull-refresh-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  color: var(--accent);
  transition: height var(--duration-short) var(--ease-standard);
}

.pull-icon {
  font-size: 1.25rem;
  line-height: 1;
  transition: transform var(--duration-short) var(--ease-standard);
}

.pull-icon.spin {
  animation: spin 800ms linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.row.selected {
  background: var(--accent-soft);
}

.checkbox-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-right: var(--space-xs);
  border: 1.5px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  flex-shrink: 0;
  color: #fff;
  transition:
    background var(--duration-short) var(--ease-standard),
    border-color var(--duration-short) var(--ease-standard);
}

.checkbox-indicator.checked {
  background: var(--accent);
  border-color: var(--accent);
}

.mobile-folder-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xxs);
  padding: var(--space-xxs) var(--space-xs);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  color: var(--ink);
  cursor: pointer;
  max-width: 200px;
}

.mobile-breadcrumb-more {
  color: var(--muted);
}

.breadcrumb-sheet-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xxs);
  padding: var(--space-xs) 0;
}

.breadcrumb-sheet-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--ink);
  font-size: 0.9375rem;
  text-align: left;
  cursor: pointer;
}

.breadcrumb-sheet-item:hover {
  background: var(--surface-soft);
}

.breadcrumb-sheet-item.active,
.breadcrumb-sheet-item.current {
  color: var(--accent);
  font-weight: 600;
}

/* TeraBox Search Bar */
.files-search-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--surface-card, rgba(255, 255, 255, 0.06));
  border: 1px solid var(--hairline);
  border-radius: 9999px;
  padding: 8px 16px;
  margin-bottom: var(--space-md);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.files-search-wrap:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft, rgba(0, 132, 255, 0.15));
}

.files-search-icon {
  color: var(--muted);
  flex-shrink: 0;
}

.files-search-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--ink);
  font-size: 14px;
  outline: none;
  min-width: 0;
}

.files-search-input::placeholder {
  color: var(--muted);
}

.files-search-clear {
  border: none;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* TeraBox Tabs */
.tabs-header {
  margin-bottom: var(--space-sm);
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

/* Folder Header / Breadcrumbs */
.folder-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 0;
  margin-bottom: var(--space-xs);
}

.folder-back-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: transparent;
  border: none;
  color: var(--ink);
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  padding: 0;
}

.folder-header-title {
  max-width: 240px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-hierarchy-btn {
  background: transparent;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Sub-toolbar */
.files-sub-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
  margin-bottom: var(--space-sm);
  font-size: 13px;
}

.sub-sort-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: var(--muted);
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 0;
}

.sub-sort-btn:active {
  color: var(--ink);
}

.sub-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sub-icon-btn {
  background: transparent;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-sm, 8px);
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  cursor: pointer;
}

.sub-icon-btn.filter-active {
  color: var(--accent);
  border-color: var(--accent);
  background: var(--accent-soft);
}

/* Root Vault Card */
.vault-entry-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: var(--radius-lg, 16px);
  background: var(--surface-card, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--hairline);
  margin-bottom: 12px;
  cursor: pointer;
  text-decoration: none;
  color: inherit;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease;
}

.vault-entry-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.vault-entry-card:active {
  background: var(--accent-soft);
  border-color: var(--accent);
}

.vault-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.vault-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.vault-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink);
}

.vault-hint {
  font-size: 12px;
  color: var(--muted);
}

.vault-arrow {
  color: var(--muted);
}

/* Files Container (List & Grid Modes) */
.files-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.files-container.grid-mode {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

@media (min-width: 640px) {
  .files-container.grid-mode {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .files-container.grid-mode {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* File Item Card */
.file-item-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-md, 12px);
  background: var(--canvas);
  border: 1px solid transparent;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease;
  user-select: none;
}

.file-item-card:hover {
  background: var(--surface-card, rgba(255, 255, 255, 0.03));
}

.file-item-card:active {
  background: var(--accent-soft);
  border-color: var(--accent);
}

.file-item-card.selected {
  background: var(--accent-soft);
  border-color: var(--accent);
}

.file-item-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.file-icon-badge {
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

.file-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.file-item-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-item-sub {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-variant-numeric: tabular-nums;
}

.file-item-more-btn {
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

.file-item-more-btn:active {
  background: var(--surface-card);
  color: var(--ink);
}

/* Grid mode adjustments */
.files-container.grid-mode .file-item-card {
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 16px 12px;
  border: 1px solid var(--hairline);
  position: relative;
}

.files-container.grid-mode .file-icon-badge {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  margin-bottom: 6px;
}

.files-container.grid-mode .file-item-info {
  align-items: center;
  width: 100%;
}

.files-container.grid-mode .file-item-more-btn {
  position: absolute;
  top: 6px;
  right: 6px;
}
</style>
