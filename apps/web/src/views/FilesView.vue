<script setup lang="ts">
import { computed, inject, onBeforeUnmount, onMounted, ref, watch } from 'vue'
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
import ShareSheet from '@/components/ShareSheet.vue'
import ShareUserSheet from '@/components/ShareUserSheet.vue'
import LoadingSkeletonFiles from '@/components/LoadingSkeletonFiles.vue'
import BatchActionBar from '@/components/BatchActionBar.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import { usePullToRefresh } from '@/lib/usePullToRefresh'
import { useLongPress } from '@/lib/useLongPress'
import { mimeIcon, mimeLabel, resolveContentType } from '@/lib/mimeIcon'
import type {
  Browser,
  DownloadURL,
  FavoriteFile,
  SearchFilters,
  SearchResult,
  ShareLinkInfo,
  ShareLinkTTL,
  UploadSession,
} from '@/api/types'
import SearchFilterSheet from '@/components/SearchFilterSheet.vue'
import { useI18n } from '@/lib/i18n'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const { t } = useI18n()
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
const totalItemsCount = computed(() => (browser.value?.folders.length ?? 0) + (browser.value?.files.length ?? 0))

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

// Long-press detection
const { start: startLongPress, move: moveLongPress, end: endLongPress, cancel: cancelLongPress, shouldIgnoreClick } = useLongPress({
  onLongPress: (payload) => {
    if (payload && typeof payload === 'object' && 'type' in payload && 'id' in payload) {
      const p = payload as { type: 'file' | 'folder'; id: string }
      startSelection(p.type, p.id)
    }
  },
})

// Mobile breadcrumb sheet
const breadcrumbSheetOpen = ref(false)

// In-app preview
const previewOpen = ref(false)
const previewFile = ref<{ id: string; name: string; mimeType: string; url: string } | null>(null)

async function previewMediaFile(file: { id: string; name: string; mimeType: string }) {
  try {
    const out = await api<DownloadURL>(`/files/${file.id}/download`)
    previewFile.value = { id: file.id, name: file.name, mimeType: file.mimeType, url: out.downloadUrl }
    previewOpen.value = true
  } catch (e) {
    error.value = formatApiError(e, 'Could not preview file')
  }
}

onMounted(() => {
  if (filesPageRef.value) attachListeners(filesPageRef.value)
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

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
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
      ui.showToast(`Moved ${totalSelectedCount.value} items`)
      clearSelection()
      await loadBrowser()
      return
    }

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

async function batchDelete() {
  const count = totalSelectedCount.value
  if (count === 0) return
  const ok = await ui.confirm({
    title: `Move ${count} items to trash?`,
    message: 'You can restore them from Trash later.',
    confirmLabel: 'Move to trash',
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
    ui.showToast(`Moved ${count} items to trash`)
    clearSelection()
    await loadBrowser()
    await reloadStorage?.()
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
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
  pickerTitle.value = `Move ${totalSelectedCount.value} items`
  pickerOpen.value = true
}

async function batchDownload() {
  for (const id of selectedFileIds.value) {
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
    ...(canPreview ? [{ id: 'preview', label: 'Preview', icon: 'eye' }] : []),
    { id: 'download', label: 'Download', icon: 'download' },
    {
      id: 'favorite',
      label: favorited ? 'Remove from favorites' : 'Add to favorites',
      icon: favorited ? 'star-filled' : 'star',
    },
    { id: 'share', label: 'Share link', icon: 'share' },
    { id: 'share-user', label: 'Share with user', icon: 'users' },
    { id: 'rename', label: 'Rename', icon: 'pencil' },
    { id: 'move', label: 'Move', icon: 'move' },
    { id: 'delete', label: 'Move to trash', icon: 'trash', danger: true },
  ])
  if (action === 'preview') await previewMediaFile({ id: file.id, name: file.name, mimeType: file.mimeType ?? '' })
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
  if (action === 'delete') await deleteFile(file.id)
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
    const entry = { url: created.url ?? '', expiresAt: created.expiresAt, createdAt: created.createdAt }
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

async function shareWithUser(email: string) {
  const fileId = shareUserTargetId.value
  if (!fileId) return
  const name = shareUserFileName.value
  error.value = ''
  try {
    const res = await api<{ invited?: boolean }>(`/shares`, {
      method: 'POST',
      body: JSON.stringify({ resourceType: 'file', resourceId: fileId, email }),
    })
    shareUserOpen.value = false
    if (res.invited) {
      ui.showToast(`Invitation sent to ${email}`)
    } else {
      ui.showToast(`Shared "${name}" with ${email}`, 'success')
    }
  } catch (e) {
    error.value = formatApiError(e, 'Could not share')
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
    <div v-if="pullDistance > 0 || isRefreshing" class="pull-refresh-bar" :style="{ height: `${pullDistance}px` }">
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
    <h1 class="page-title desktop-only">{{ t.navFiles }}</h1>

    <div class="segment-tabs" role="tablist" aria-label="File views">
      <button
        type="button"
        role="tab"
        class="segment-tab"
        :class="{ active: segment === 'all' }"
        :aria-selected="segment === 'all'"
        @click="segment = 'all'"
      >
        {{ t.all }}
      </button>
      <button
        type="button"
        role="tab"
        class="segment-tab"
        :class="{ active: segment === 'favorites' }"
        :aria-selected="segment === 'favorites'"
        @click="segment = 'favorites'"
      >
        {{ t.favorites }}
      </button>
    </div>

    <nav v-if="segment === 'all'" class="breadcrumb" aria-label="Folder path">
      <button
        v-if="folderId"
        type="button"
        class="btn ghost mobile-back"
        @click="openFolder(browser?.folder?.parentId ?? null)"
      >
        ← {{ t.back }}
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
      <button
        type="button"
        class="mobile-folder-btn mobile-only"
        aria-label="View folder hierarchy"
        @click="breadcrumbSheetOpen = true"
      >
        <span v-if="browser?.folder" class="mobile-current">{{ browser.folder.name }}</span>
        <span v-else-if="!folderId" class="mobile-current">{{ t.root }}</span>
        <Icon name="more" :size="14" class="mobile-breadcrumb-more" />
      </button>
    </nav>

    <div v-if="segment === 'all'" class="toolbar toolbar-sticky">
      <input
        v-model="searchQuery"
        class="search-input"
        type="search"
        :placeholder="t.searchByName"
        :aria-label="t.searchByName"
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
      <button class="btn accent desktop-only" type="button" @click="triggerUpload">{{ t.upload }}</button>
      <button class="btn desktop-only" type="button" @click="triggerFolderUpload">{{ t.uploadFolder }}</button>
      <button class="btn desktop-only" type="button" @click="folderSheetOpen = true">{{ t.newFolder }}</button>
      <button class="btn mobile-only" type="button" :aria-label="t.newFolder" @click="folderSheetOpen = true">
        {{ t.newFolder }}
      </button>
      <button
        v-if="searchResults"
        class="btn ghost"
        type="button"
        @click="searchResults = null; searchQuery = ''"
      >
        {{ t.clear }}
      </button>
    </div>

    <UploadProgress :progress="uploadProgress" />
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonFiles v-if="segment === 'all' && loading && !searchResults" mode="browse" />
    <LoadingSkeletonFiles v-else-if="searchLoading" mode="search" />

    <TransitionGroup v-if="segment === 'all' && !loading && searchResults" name="row" tag="section" class="list">
      <h2 key="search-title" class="section-title">
        {{ searchResults.folders.length + searchResults.files.length }} {{ t.results }}
        <template v-if="hasActiveFilters"> · {{ t.filtered }}<!-- · filtered --></template>
      </h2>
      <div v-if="hasActiveFilters" key="filter-chips" class="filter-chips">
        <span v-for="chip in activeFilterChips" :key="chip.key" class="filter-chip">
          {{ chip.label }}
          <button type="button" class="chip-remove" aria-label="Remove filter" @click="removeFilter(chip.key)">
            ×
          </button>
        </span>
        <button type="button" class="chip-clear" @click="clearFilters">{{ t.clearAll }}<!-- Clear all --></button>
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
        :title="hasActiveFilters ? t.noResultsFiltered : t.noResults"
        :description="hasActiveFilters ? t.noResultsFilteredDesc : t.noResultsDesc"
        icon="file"
      >
        <!-- No results match your filters -->
        <button v-if="hasActiveFilters" type="button" class="btn" @click="clearFilters">{{ t.clearFilters }}</button>
      </EmptyState>
    </TransitionGroup>

    <TransitionGroup
      v-else-if="segment === 'all' && !loading"
      name="row"
      tag="section"
      class="list"
    >
      <div
        v-for="folder in browser?.folders ?? []"
        :key="folder.id"
        class="row tappable"
        :class="{ selected: selectedFolderIds.has(folder.id) }"
        @touchstart.passive="startLongPress($event, { type: 'folder', id: folder.id })"
        @touchmove.passive="moveLongPress"
        @touchend="endLongPress"
        @touchcancel="cancelLongPress"
        @click="shouldIgnoreClick ? undefined : (isSelecting ? toggleSelectItem('folder', folder.id) : openFolder(folder.id))"
      >
        <span v-if="isSelecting" class="checkbox-indicator" :class="{ checked: selectedFolderIds.has(folder.id) }">
          <Icon v-if="selectedFolderIds.has(folder.id)" name="check" :size="14" />
        </span>
        <span class="name"><Icon name="folder" :size="18" class="row-icon" />{{ folder.name }}</span>
        <button
          v-if="!isSelecting"
          class="btn icon-only"
          type="button"
          aria-label="Folder actions"
          @click.stop="openFolderActions(folder)"
        >
          <Icon name="more" :size="18" />
        </button>
      </div>
      <div
        v-for="file in browser?.files ?? []"
        :key="file.id"
        class="row tappable"
        :class="{ selected: selectedFileIds.has(file.id) }"
        @touchstart.passive="startLongPress($event, { type: 'file', id: file.id })"
        @touchmove.passive="moveLongPress"
        @touchend="endLongPress"
        @touchcancel="cancelLongPress"
        @click="shouldIgnoreClick ? undefined : (isSelecting ? toggleSelectItem('file', file.id) : openFileActions(file))"
      >
        <span v-if="isSelecting" class="checkbox-indicator" :class="{ checked: selectedFileIds.has(file.id) }">
          <Icon v-if="selectedFileIds.has(file.id)" name="check" :size="14" />
        </span>
        <span class="name">
          <Icon
            :name="isFavorited(file.id) ? 'star-filled' : mimeIcon(file.mimeType)"
            :size="18"
            class="row-icon favorite-icon"
          />{{ file.name }}
        </span>
        <span class="meta desktop-only">{{ mimeLabel(file.mimeType) }} · {{ formatBytes(file.sizeBytes) }}</span>
        <button
          v-if="!isSelecting"
          class="btn icon-only"
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

    <TransitionGroup v-else-if="segment === 'favorites'" name="row" tag="section" class="list">
      <LoadingSkeletonFiles v-if="favoritesLoading" key="fav-skeleton" mode="browse" />
      <template v-else>
        <div
          v-for="file in favorites ?? []"
          :key="file.id"
          class="row tappable"
          @click="openFileActions(file)"
        >
          <span class="name">
            <Icon name="star-filled" :size="18" class="row-icon favorite-icon" />{{ file.name }}
          </span>
          <span class="meta desktop-only">{{ formatBytes(file.sizeBytes) }}</span>
          <button class="btn icon-only" type="button" aria-label="File actions" @click.stop="openFileActions(file)">
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
    <input ref="folderInputRef" type="file" class="sr-only" webkitdirectory @change="onFolderUploadChange" />

    <UploadFab
      v-if="segment === 'all' && !isSelecting"
      :label="t.uploadFile"
      :disabled="uploadProgress !== null"
      @click="triggerUpload"
    />

    <BatchActionBar
      v-if="isSelecting"
      :selected-count="totalSelectedCount"
      :total-count="totalItemsCount"
      @close="clearSelection"
      @select-all="selectAllItems"
      @clear-selection="clearSelection"
      @delete="batchDelete"
      @move="batchMove"
      @favorite="batchFavorite"
      @download="batchDownload"
    />

    <BottomSheet :open="folderSheetOpen" :title="t.newFolder" @close="folderSheetOpen = false">
      <label class="field">
        <span>{{ t.folderName }}</span>
        <input v-model="newFolderName" type="text" @keyup.enter="createFolder" />
      </label>
      <button type="button" class="btn block ink" @click="createFolder">{{ t.createFolder }}</button>
    </BottomSheet>

    <BottomSheet :open="breadcrumbSheetOpen" :title="t.location" @close="breadcrumbSheetOpen = false">
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
          <span><strong>{{ browser.folder.name }}</strong> ({{ t.current }})</span>
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
      @share="shareWithUser"
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
  transition: background var(--duration-short) var(--ease-standard), color var(--duration-short) var(--ease-standard);
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
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
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
  transition: background var(--duration-short) var(--ease-standard), border-color var(--duration-short) var(--ease-standard);
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
</style>
