<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import { usePullToRefresh } from '@/lib/usePullToRefresh'
import PhotoThumb from '@/components/PhotoThumb.vue'
import PhotoMediaSheet from '@/components/PhotoMediaSheet.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import LoadingSkeletonPhotos from '@/components/LoadingSkeletonPhotos.vue'
import { cellDelay } from '@/lib/motion'
import { useI18n } from '@/lib/i18n'
import type { Album, DownloadURL, Timeline, TimelineItem } from '@/api/types'

const router = useRouter()
const ui = useUiStore()
const { t } = useI18n()
const photosPageRef = ref<HTMLElement | null>(null)

const groups = ref<Timeline['groups']>([])
const nextBefore = ref<string | undefined>()
const albums = ref<Album[]>([])
const newAlbumName = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')

const mediaOpen = ref(false)
const mediaItem = ref<TimelineItem | null>(null)
const pendingSheetAction = ref<'view' | 'download' | null>(null)
const favoriteIds = ref<Set<string>>(new Set())
const lightboxOpen = ref(false)
const lightboxUrl = ref('')

async function loadTimeline(before?: string) {
  const params = new URLSearchParams()
  if (before) params.set('before', before)
  return api<Timeline>(`/photos/timeline${params.toString() ? `?${params}` : ''}`)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await loadTimeline()
    groups.value = data.groups
    nextBefore.value = data.nextBefore
    const list = await api<{ albums: Album[] }>('/photos/albums')
    albums.value = list.albums
    await loadFavorites()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load photos')
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!nextBefore.value || loadingMore.value) return
  loadingMore.value = true
  error.value = ''
  try {
    const data = await loadTimeline(nextBefore.value)
    const merged = [...groups.value]
    for (const newGroup of data.groups) {
      const existing = merged.find((g) => g.date === newGroup.date)
      if (existing) {
        existing.items.push(...newGroup.items)
      } else {
        merged.push(newGroup)
      }
    }
    groups.value = merged
    nextBefore.value = data.nextBefore
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load more')
  } finally {
    loadingMore.value = false
  }
}

async function createAlbum() {
  if (!newAlbumName.value.trim()) return
  error.value = ''
  try {
    await api('/photos/albums', {
      method: 'POST',
      body: JSON.stringify({ name: newAlbumName.value.trim() }),
    })
    newAlbumName.value = ''
    ui.showToast('Album created')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to create album')
  }
}

async function openAlbumActions(album: Album) {
  const action = await ui.openActionSheet(album.name, [
    { id: 'open', label: 'Open', icon: 'arrow-right' },
    { id: 'rename', label: 'Rename', icon: 'pencil' },
    { id: 'delete', label: 'Delete album', icon: 'trash', danger: true },
  ])
  if (!action) return
  if (action === 'open') await router.push(`/photos/albums/${album.id}`)
  if (action === 'rename') await renameAlbum(album)
  if (action === 'delete') await deleteAlbum(album.id, album.name)
}

async function renameAlbum(album: Album) {
  const name = await ui.prompt({ title: t.value.renameAlbum, label: t.value.displayName, initialValue: album.name })
  if (!name || name === album.name) return
  error.value = ''
  try {
    await api(`/photos/albums/${album.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ name }),
    })
    ui.showToast(t.value.renameAlbum)
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Rename failed')
  }
}

async function deleteAlbum(id: string, name: string) {
  const ok = await ui.confirm({
    title: `${t.value.deleteAlbum}?`,
    message: `"${name}" ${t.value.deleteAlbumConfirm}`,
    confirmLabel: t.value.deleteAlbum,
    danger: true,
  })
  if (!ok) return
  error.value = ''
  try {
    await api(`/photos/albums/${id}`, { method: 'DELETE' })
    ui.showToast('Album deleted')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Delete failed')
  }
}

function openMedia(item: TimelineItem) {
  mediaItem.value = item
  mediaOpen.value = true
}

const isFavorited = (id: string) => favoriteIds.value.has(id)

async function toggleFavorite(item: TimelineItem) {
  const wasFavorited = isFavorited(item.id)
  error.value = ''
  try {
    if (wasFavorited) {
      await api(`/files/${item.id}/favorite`, { method: 'DELETE' })
      ui.showToast(`Removed "${item.name}" from favorites`)
    } else {
      await api(`/files/${item.id}/favorite`, { method: 'PUT' })
      ui.showToast(`Added "${item.name}" to favorites`)
    }
    await loadFavorites()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to update favorite')
  }
}

async function loadFavorites() {
  const out = await api<{ files: Array<{ id: string }> }>('/files/favorites?limit=100')
  favoriteIds.value = new Set(out.files.map((f) => f.id))
}

async function toggleFavoriteFromSheet() {
  if (!mediaItem.value) return
  await toggleFavorite(mediaItem.value)
}

async function viewMedia() {
  if (!mediaItem.value) return
  pendingSheetAction.value = 'view'
  mediaOpen.value = false
}

async function openLightbox(item: TimelineItem) {
  error.value = ''
  mediaItem.value = item
  try {
    const out = await api<DownloadURL>(`/files/${item.id}/download`)
    lightboxUrl.value = out.downloadUrl
    lightboxOpen.value = true
  } catch (e) {
    error.value = formatApiError(e, 'View failed')
  }
}

async function handleSheetAfterLeave() {
  const action = pendingSheetAction.value
  pendingSheetAction.value = null
  if (!action || !mediaItem.value) return
  if (action === 'view') await openLightbox(mediaItem.value)
  if (action === 'download') await downloadMedia()
}

const allTimelineItems = computed(() => groups.value.flatMap((g) => g.items))
const currentLightboxIndex = computed(() =>
  mediaItem.value ? allTimelineItems.value.findIndex((i) => i.id === mediaItem.value?.id) : -1,
)
const hasNextMedia = computed(
  () => currentLightboxIndex.value !== -1 && currentLightboxIndex.value < allTimelineItems.value.length - 1,
)
const hasPrevMedia = computed(() => currentLightboxIndex.value > 0)

async function nextMedia() {
  if (!hasNextMedia.value) return
  const nextItem = allTimelineItems.value[currentLightboxIndex.value + 1]
  if (nextItem) await openLightbox(nextItem)
}

async function prevMedia() {
  if (!hasPrevMedia.value) return
  const prevItem = allTimelineItems.value[currentLightboxIndex.value - 1]
  if (prevItem) await openLightbox(prevItem)
}

async function downloadMedia() {
  if (!mediaItem.value) return
  if (mediaOpen.value) {
    pendingSheetAction.value = 'download'
    mediaOpen.value = false
    return
  }
  error.value = ''
  try {
    const out = await api<DownloadURL>(`/files/${mediaItem.value.id}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, 'Download failed')
  }
}

const { pullDistance, isRefreshing, attachListeners } = usePullToRefresh(photosPageRef, {
  onRefresh: load,
})

onMounted(() => {
  void load()
  if (photosPageRef.value) attachListeners(photosPageRef.value)
})

onBeforeUnmount(() => {
  mediaOpen.value = false
  pendingSheetAction.value = null
})
</script>

<template>
  <div ref="photosPageRef" class="photos-page">
    <div v-if="pullDistance > 0 || isRefreshing" class="pull-refresh-bar" :style="{ height: `${pullDistance}px` }">
      <span class="pull-icon" :class="{ spin: isRefreshing }">{{ isRefreshing ? '↻' : '↓' }}</span>
    </div>
    <h1 class="page-title desktop-only">{{ t.photosTitle }}</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonPhotos v-if="loading" variant="initial" />
    <div v-else>
      <section class="section" aria-labelledby="albums-heading">
        <h2 id="albums-heading" class="section-title">{{ t.albums }}</h2>
        <form class="album-form" @submit.prevent="createAlbum">
          <label class="field album-field">
            <span class="sr-only">{{ t.newAlbumName }}</span>
            <input v-model="newAlbumName" type="text" :placeholder="t.newAlbumName" autocomplete="off" />
          </label>
          <button class="btn ink" type="submit" :disabled="!newAlbumName.trim()">{{ t.create }}</button>
        </form>
        <div v-if="albums.length" class="list">
          <div
            v-for="(album, index) in albums"
            :key="album.id"
            class="row tappable appear"
            :style="{ animationDelay: cellDelay(index) }"
            @click="router.push(`/photos/albums/${album.id}`)"
          >
            <div class="album-cover" aria-hidden="true">
              <img v-if="album.coverUrl" :src="album.coverUrl" alt="" loading="lazy" />
              <Icon v-else-if="album.coverFileId" name="video" :size="20" />
              <Icon v-else name="photos" :size="20" />
            </div>
            <button type="button" class="name link-btn" @click.stop="router.push(`/photos/albums/${album.id}`)">
              {{ album.name }}
            </button>
            <span class="meta album-count">{{ album.itemCount }}</span>
            <button class="btn icon-only" type="button" :aria-label="t.albumMenu" @click.stop="openAlbumActions(album)">
              <Icon name="more" :size="18" />
            </button>
          </div>
        </div>
        <EmptyState
          v-else
          compact
          :title="t.noAlbums"
          :description="t.noAlbumsDesc"
          icon="photos"
        />
      </section>

      <section v-for="group in groups" :key="group.date" class="section" :aria-label="group.date">
        <h2 class="section-title">{{ group.date }}</h2>
        <div class="grid photos">
          <PhotoThumb
            v-for="(item, index) in group.items"
            :key="item.id"
            :mime-type="item.mimeType"
            :name="item.name"
            :thumbnail-url="item.thumbnailUrl"
            class="appear"
            :style="{ animationDelay: cellDelay(index) }"
            @click="openMedia(item)"
          />
        </div>
      </section>

      <EmptyState
        v-if="groups.length === 0"
        :title="t.noPhotos"
        :description="t.noPhotosDesc"
        icon="photos"
      />

      <div v-if="nextBefore" class="load-more">
        <LoadingSkeletonPhotos v-if="loadingMore" variant="more" />
        <button v-else type="button" class="btn block" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? t.loading : t.loadMore }}
        </button>
      </div>

      <PhotoMediaSheet
        :open="mediaOpen"
        :name="mediaItem?.name ?? ''"
        :favorited="mediaItem ? isFavorited(mediaItem.id) : false"
        @view="viewMedia"
        @download="downloadMedia"
        @favorite="toggleFavoriteFromSheet"
        @close="mediaOpen = false"
        @after-leave="handleSheetAfterLeave"
      />

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
    </div>
  </div>
</template>

<style scoped>
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

.section {
  margin-bottom: var(--space-lg);
}

.album-form {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-xs);
  margin-bottom: var(--space-sm);
}

.album-field {
  flex: 1 1 100%;
  margin-bottom: 0;
}

.album-form .btn {
  width: 100%;
}

.link-btn {
  flex: 1;
  min-width: 0;
  text-align: left;
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  font-weight: 500;
  color: var(--ink);
  cursor: pointer;
}

.album-count {
  flex-shrink: 0;
  min-width: 24px;
  text-align: center;
  padding: 0.1rem 0.5rem;
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--muted);
  font-size: 0.75rem;
  font-weight: 600;
}

.album-cover {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--surface-card);
  color: var(--muted);
  overflow: hidden;
}

.album-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.load-more {
  margin-top: var(--space-md);
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: block;
  }

  .album-field {
    flex: 1 1 auto;
  }

  .album-form .btn {
    width: auto;
  }
}
</style>
