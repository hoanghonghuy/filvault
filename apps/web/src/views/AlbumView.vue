<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import PhotoThumb from '@/components/PhotoThumb.vue'
import Icon from '@/components/AppIcon.vue'
import PhotoMediaSheet from '@/components/PhotoMediaSheet.vue'
import MediaPickerSheet from '@/components/MediaPickerSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import LoadingSkeletonAlbum from '@/components/LoadingSkeletonAlbum.vue'
import { cellDelay } from '@/lib/motion'
import type { AlbumDetail, DownloadURL, TimelineItem } from '@/api/types'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()

const album = ref<AlbumDetail | null>(null)
const error = ref('')
const loading = ref(false)
const pickerOpen = ref(false)
const mediaOpen = ref(false)
const mediaItem = ref<TimelineItem | null>(null)

const albumId = computed(() => String(route.params.id))
const itemIds = computed(() => album.value?.items.map((i) => i.id) ?? [])

async function load() {
  loading.value = true
  error.value = ''
  try {
    album.value = await api<AlbumDetail>(`/photos/albums/${albumId.value}`)
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load album')
  } finally {
    loading.value = false
  }
}

async function openAlbumMenu() {
  if (!album.value) return
  const action = await ui.openActionSheet(album.value.name, [
    { id: 'add', label: 'Add photos' },
    { id: 'rename', label: 'Rename album' },
    { id: 'delete', label: 'Delete album', danger: true },
  ])
  if (action === 'add') pickerOpen.value = true
  if (action === 'rename') await renameAlbum()
  if (action === 'delete') await deleteAlbum()
}

async function renameAlbum() {
  if (!album.value) return
  const name = await ui.prompt({
    title: 'Rename album',
    label: 'Name',
    initialValue: album.value.name,
  })
  if (!name || name === album.value.name) return
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}`, {
      method: 'PATCH',
      body: JSON.stringify({ name }),
    })
    ui.showToast('Album renamed')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Rename failed')
  }
}

async function deleteAlbum() {
  if (!album.value) return
  const ok = await ui.confirm({
    title: 'Delete album?',
    message: `"${album.value.name}" will be removed. Your files stay in My Files.`,
    confirmLabel: 'Delete album',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}`, { method: 'DELETE' })
    ui.showToast('Album deleted')
    await router.push('/photos')
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
  }
}

async function addItem(fileId: string) {
  pickerOpen.value = false
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}/items`, {
      method: 'POST',
      body: JSON.stringify({ fileIds: [fileId] }),
    })
    ui.showToast('Added to album')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to add item')
  }
}

async function openItemActions(item: TimelineItem) {
  const action = await ui.openActionSheet(item.name, [
    { id: 'view', label: 'View' },
    { id: 'download', label: 'Download' },
    { id: 'remove', label: 'Remove from album', danger: true },
  ])
  if (action === 'view') {
    mediaItem.value = item
    mediaOpen.value = true
  }
  if (action === 'download') {
    mediaItem.value = item
    await downloadMedia()
  }
  if (action === 'remove') await removeItem(item.id, item.name)
}

async function removeItem(fileId: string, name: string) {
  const ok = await ui.confirm({
    title: 'Remove from album?',
    message: `"${name}" will be removed from this album only.`,
    confirmLabel: 'Remove',
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}/items/${fileId}`, { method: 'DELETE' })
    ui.showToast('Removed from album')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Remove failed')
  }
}

async function viewMedia() {
  if (!mediaItem.value) return
  error.value = ''
  try {
    const out = await api<DownloadURL>(`/files/${mediaItem.value.id}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
    mediaOpen.value = false
  } catch (e) {
    error.value = formatApiError(e, 'View failed')
  }
}

async function downloadMedia() {
  await viewMedia()
}

watch(() => route.params.id, load, { immediate: true })
</script>

<template>
  <div>
    <div class="album-header">
      <button type="button" class="btn ghost mobile-back" @click="router.push('/photos')">
        ← Photos
      </button>
      <h1 class="album-title">{{ album?.name ?? 'Album' }}</h1>
      <button type="button" class="btn icon-only" aria-label="Album menu" @click="openAlbumMenu">
        <Icon name="more" :size="18" />
      </button>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonAlbum v-if="loading" />
    <div v-else>
      <div v-if="album?.items.length" class="grid photos">
        <PhotoThumb
          v-for="(item, index) in album.items"
          :key="item.id"
          :mime-type="item.mimeType"
          :name="item.name"
          :thumbnail-url="item.thumbnailUrl"
          class="appear"
          :style="{ animationDelay: cellDelay(index) }"
          @click="openItemActions(item)"
        />
      </div>

      <EmptyState
        v-else
        title="Album is empty"
        description="Add photos or videos from your library."
        action-label="Add photos"
        icon="photos"
        @action="pickerOpen = true"
      />
    </div>

    <PhotoMediaSheet
      :open="mediaOpen"
      :name="mediaItem?.name ?? ''"
      @view="viewMedia"
      @download="downloadMedia"
      @close="mediaOpen = false"
    />

    <MediaPickerSheet
      :open="pickerOpen"
      :exclude-ids="itemIds"
      @select="addItem"
      @close="pickerOpen = false"
    />
  </div>
</template>

<style scoped>
.album-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
}

.album-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-back {
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
}

@media (min-width: 768px) {
  .mobile-back {
    display: none;
  }

  .album-title {
    font-size: 1.5rem;
    letter-spacing: -0.02em;
  }
}
</style>
