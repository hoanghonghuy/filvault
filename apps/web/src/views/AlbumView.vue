<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import PhotoThumb from '@/components/PhotoThumb.vue'
import Icon from '@/components/AppIcon.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import MediaPickerSheet from '@/components/MediaPickerSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import LoadingSkeletonAlbum from '@/components/LoadingSkeletonAlbum.vue'
import { cellDelay } from '@/lib/motion'
import { useI18n } from '@/lib/i18n'
import { albumRuntimeCopy } from '@/lib/albumCopy'
import type { AlbumDetail, DownloadURL, TimelineItem } from '@/api/types'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const { t, locale } = useI18n()
const copy = computed(() => albumRuntimeCopy(locale.value))

const album = ref<AlbumDetail | null>(null)
const error = ref('')
const loading = ref(false)
const pickerOpen = ref(false)
const mediaItem = ref<TimelineItem | null>(null)
const lightboxOpen = ref(false)
const lightboxUrl = ref('')

const albumId = computed(() => (Array.isArray(route.params.id) ? route.params.id[0] : (route.params.id ?? '')))
const itemIds = computed(() => album.value?.items.map((i) => i.id) ?? [])

async function load() {
  loading.value = true
  error.value = ''
  try {
    album.value = await api<AlbumDetail>(`/photos/albums/${albumId.value}`)
  } catch (e) {
    error.value = formatApiError(e, copy.value.loadFailed)
  } finally {
    loading.value = false
  }
}

async function openAlbumMenu() {
  if (!album.value) return
  const action = await ui.openActionSheet(album.value.name, [
    { id: 'add', label: copy.value.addPhotos, icon: 'plus' },
    { id: 'rename', label: copy.value.renameAlbum, icon: 'pencil' },
    { id: 'delete', label: copy.value.deleteAlbum, icon: 'trash', danger: true },
  ])
  if (!action) return
  if (action === 'add') pickerOpen.value = true
  if (action === 'rename') await renameAlbum()
  if (action === 'delete') await deleteAlbum()
}

async function renameAlbum() {
  if (!album.value) return
  const name = await ui.prompt({
    title: t.value.renameAlbum,
    label: t.value.displayName,
    initialValue: album.value.name,
  })
  if (!name || name === album.value.name) return
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}`, {
      method: 'PATCH',
      body: JSON.stringify({ name }),
    })
    ui.showToast(t.value.renameAlbum)
    await load()
  } catch (e) {
    error.value = formatApiError(e, copy.value.renameFailed)
  }
}

async function deleteAlbum() {
  if (!album.value) return
  const ok = await ui.confirm({
    title: `${t.value.deleteAlbum}?`,
    message: `"${album.value.name}" ${t.value.deleteAlbumConfirm}`,
    confirmLabel: t.value.deleteAlbum,
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}`, { method: 'DELETE' })
    ui.showToast(t.value.deleteAlbum)
    await router.push('/photos')
  } catch (e) {
    error.value = formatApiError(e, copy.value.deleteFailed)
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
    ui.showToast(t.value.addPhotos)
    await load()
  } catch (e) {
    error.value = formatApiError(e, copy.value.addItemFailed)
  }
}

async function openItemActions(item: TimelineItem) {
  const isCover = album.value?.coverFileId === item.id
  const action = await ui.openActionSheet(item.name, [
    { id: 'download', label: copy.value.download, icon: 'download' },
    isCover
      ? { id: 'unset-cover', label: copy.value.removeCover, icon: 'restore' }
      : { id: 'set-cover', label: copy.value.setCover, icon: 'image' },
    { id: 'remove', label: copy.value.removeFromAlbum, icon: 'trash', danger: true },
  ])
  if (!action) return
  if (action === 'download') {
    mediaItem.value = item
    await downloadMedia()
  }
  if (action === 'set-cover') await setCover(item)
  if (action === 'unset-cover') await removeCover()
  if (action === 'remove') await removeItem(item.id, item.name)
}

async function setCover(item: TimelineItem) {
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}/cover`, {
      method: 'POST',
      body: JSON.stringify({ fileId: item.id }),
    })
    ui.showToast(copy.value.coverUpdated)
    await load()
  } catch (e) {
    error.value = formatApiError(e, copy.value.setCoverFailed)
  }
}

async function removeCover() {
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}/cover`, { method: 'DELETE' })
    ui.showToast(copy.value.coverReset)
    await load()
  } catch (e) {
    error.value = formatApiError(e, copy.value.removeCoverFailed)
  }
}

async function removeItem(fileId: string, name: string) {
  const ok = await ui.confirm({
    title: `${t.value.removeFromAlbum}?`,
    message: copy.value.removeConfirmation(name),
    confirmLabel: t.value.remove,
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/photos/albums/${albumId.value}/items/${fileId}`, { method: 'DELETE' })
    ui.showToast(t.value.removeFromAlbum)
    await load()
  } catch (e) {
    error.value = formatApiError(e, copy.value.removeFailed)
  }
}

async function openLightbox(item: TimelineItem) {
  error.value = ''
  mediaItem.value = item
  try {
    const out = await api<DownloadURL>(`/files/${item.id}/download`)
    lightboxUrl.value = out.downloadUrl
    lightboxOpen.value = true
  } catch (e) {
    error.value = formatApiError(e, copy.value.viewFailed)
  }
}

const allAlbumItems = computed(() => album.value?.items ?? [])
const currentLightboxIndex = computed(() =>
  mediaItem.value ? allAlbumItems.value.findIndex((i) => i.id === mediaItem.value?.id) : -1,
)
const hasNextMedia = computed(
  () => currentLightboxIndex.value !== -1 && currentLightboxIndex.value < allAlbumItems.value.length - 1,
)
const hasPrevMedia = computed(() => currentLightboxIndex.value > 0)

async function nextMedia() {
  if (!hasNextMedia.value) return
  const nextItem = allAlbumItems.value[currentLightboxIndex.value + 1]
  if (nextItem) await openLightbox(nextItem)
}

async function prevMedia() {
  if (!hasPrevMedia.value) return
  const prevItem = allAlbumItems.value[currentLightboxIndex.value - 1]
  if (prevItem) await openLightbox(prevItem)
}

async function downloadMedia() {
  if (!mediaItem.value) return
  error.value = ''
  try {
    const out = await api<DownloadURL>(`/files/${mediaItem.value.id}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, copy.value.downloadFailed)
  }
}

watch(() => route.params.id, load, { immediate: true })

onBeforeUnmount(() => {
  pickerOpen.value = false
})
</script>

<template>
  <div>
    <div class="album-header">
      <RouterLink to="/photos" class="btn ghost album-back" :aria-label="`${t.back}: ${t.photosTitle}`">
        <span aria-hidden="true">←</span>
        <span>{{ t.photosTitle }}</span>
      </RouterLink>
      <h1 class="album-title">{{ album?.name ?? t.album }}</h1>
      <button type="button" class="btn icon-only album-menu-btn" :aria-label="t.albumMenu" @click="openAlbumMenu">
        <Icon name="more" :size="18" />
      </button>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonAlbum v-if="loading" />
    <div v-else>
      <div v-if="album?.items.length" class="grid photos">
        <div
          v-for="(item, index) in album.items"
          :key="item.id"
          class="album-media-cell appear"
          :style="{ animationDelay: cellDelay(index) }"
        >
          <PhotoThumb
            :mime-type="item.mimeType"
            :name="item.name"
            :thumbnail-url="item.thumbnailUrl"
            @click="openLightbox(item)"
            @contextmenu.prevent="openItemActions(item)"
          />
          <button
            type="button"
            class="album-media-more"
            :aria-label="`${t.moreActions}: ${item.name}`"
            @click="openItemActions(item)"
          >
            <Icon name="more" :size="18" />
          </button>
        </div>
      </div>

      <EmptyState
        v-else
        :title="t.albumEmpty"
        :description="t.albumEmptyDesc"
        :action-label="t.addPhotos"
        icon="photos"
        @action="pickerOpen = true"
      />
    </div>

    <MediaLightbox
      :open="lightboxOpen"
      :name="mediaItem?.name ?? ''"
      :mime-type="mediaItem?.mimeType ?? ''"
      :url="lightboxUrl"
      :has-next="hasNextMedia"
      :has-prev="hasPrevMedia"
      @next="nextMedia"
      @prev="prevMedia"
      @download="downloadMedia"
      @close="lightboxOpen = false"
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
  min-width: 0;
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

.album-back {
  min-height: var(--touch-min);
  min-width: var(--touch-min);
  padding: 0 var(--space-sm);
  text-decoration: none;
  flex-shrink: 0;
}

.album-menu-btn { min-width: var(--touch-min); min-height: var(--touch-min); flex-shrink: 0; }
.album-media-cell { position: relative; min-width: 0; }
.album-media-more {
  position: absolute;
  right: 6px;
  top: 6px;
  width: 44px;
  height: 44px;
  border: 0;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(17, 24, 39, 0.72);
  color: #fff;
  cursor: pointer;
  opacity: 0;
  transition: opacity var(--duration-short) var(--ease-standard);
}
.album-media-cell:hover .album-media-more,
.album-media-more:focus-visible { opacity: 1; }
.album-media-more:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }

@media (hover: none), (pointer: coarse) {
  .album-media-more { opacity: 1; }
}

@media (min-width: 768px) {
  .album-title {
    font-size: 1.5rem;
    letter-spacing: -0.02em;
  }
}

@media (prefers-reduced-motion: reduce) {
  .album-media-more { transition: none; }
}
</style>
