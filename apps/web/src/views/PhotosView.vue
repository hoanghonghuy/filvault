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
const timelineTabRef = ref<HTMLButtonElement | null>(null)
const albumsTabRef = ref<HTMLButtonElement | null>(null)

const groups = ref<Timeline['groups']>([])
const nextBefore = ref<string | undefined>()
const albums = ref<Album[]>([])
const newAlbumName = ref('')
const activePhotoTab = ref<'timeline' | 'albums'>('timeline')
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')

const mediaOpen = ref(false)
const mediaItem = ref<TimelineItem | null>(null)
const pendingSheetAction = ref<'view' | 'download' | null>(null)
const favoriteIds = ref<Set<string>>(new Set())
const lightboxOpen = ref(false)
const lightboxUrl = ref('')

function selectPhotoTab(tab: 'timeline' | 'albums', focus = false) {
  activePhotoTab.value = tab
  if (focus) {
    requestAnimationFrame(() => (tab === 'timeline' ? timelineTabRef.value : albumsTabRef.value)?.focus())
  }
}

function handleTabKeydown(event: KeyboardEvent) {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
  event.preventDefault()
  if (event.key === 'Home') return selectPhotoTab('timeline', true)
  if (event.key === 'End') return selectPhotoTab('albums', true)
  selectPhotoTab(activePhotoTab.value === 'timeline' ? 'albums' : 'timeline', true)
}

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
      if (existing) existing.items.push(...newGroup.items)
      else merged.push(newGroup)
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
    await api(`/photos/albums/${album.id}`, { method: 'PATCH', body: JSON.stringify({ name }) })
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

function openMediaActions(item: TimelineItem) {
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
  if (mediaItem.value) await toggleFavorite(mediaItem.value)
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

const { pullDistance, isRefreshing, attachListeners } = usePullToRefresh(photosPageRef, { onRefresh: load })

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
    <div
      v-if="pullDistance > 0 || isRefreshing"
      class="pull-refresh-bar"
      :style="{ height: `${pullDistance}px` }"
      role="status"
      aria-live="polite"
    >
      <span class="pull-icon" :class="{ spin: isRefreshing }" aria-hidden="true">{{ isRefreshing ? '↻' : '↓' }}</span>
      <span class="sr-only">{{ isRefreshing ? t.loading : 'Pull to refresh photos' }}</span>
    </div>

    <h1 class="page-title desktop-only">{{ t.photosTitle }}</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonPhotos v-if="loading" variant="initial" />

    <div class="tabs-header">
      <div class="tabs-pill-list" role="tablist" aria-label="Photos views" @keydown="handleTabKeydown">
        <button
          id="photos-tab-timeline"
          ref="timelineTabRef"
          type="button"
          role="tab"
          class="tab-pill"
          :class="{ active: activePhotoTab === 'timeline' }"
          :aria-selected="activePhotoTab === 'timeline'"
          aria-controls="photos-panel-timeline"
          :tabindex="activePhotoTab === 'timeline' ? 0 : -1"
          @click="selectPhotoTab('timeline')"
        >
          {{ t.timeline }}
        </button>
        <button
          id="photos-tab-albums"
          ref="albumsTabRef"
          type="button"
          role="tab"
          class="tab-pill"
          :class="{ active: activePhotoTab === 'albums' }"
          :aria-selected="activePhotoTab === 'albums'"
          aria-controls="photos-panel-albums"
          :tabindex="activePhotoTab === 'albums' ? 0 : -1"
          @click="selectPhotoTab('albums')"
        >
          {{ t.albums }} ({{ albums.length }})
        </button>
      </div>
    </div>

    <div
      v-if="activePhotoTab === 'timeline'"
      id="photos-panel-timeline"
      class="timeline-container"
      role="tabpanel"
      aria-labelledby="photos-tab-timeline"
      tabindex="0"
    >
      <section v-for="group in groups" :key="group.date" class="timeline-date-group" :aria-label="group.date">
        <div class="timeline-sticky-header">
          <h2 class="timeline-date-title">{{ group.date }}</h2>
          <span class="timeline-count-badge">{{ group.items.length }}</span>
        </div>
        <div class="grid photos">
          <div
            v-for="(item, index) in group.items"
            :key="item.id"
            class="photo-cell appear"
            :style="{ animationDelay: cellDelay(index) }"
          >
            <PhotoThumb
              :mime-type="item.mimeType"
              :name="item.name"
              :thumbnail-url="item.thumbnailUrl"
              @click="openLightbox(item)"
            />
            <button
              type="button"
              class="photo-more-btn"
              :aria-label="`${t.albumMenu}: ${item.name}`"
              @click="openMediaActions(item)"
            >
              <Icon name="more" :size="18" />
            </button>
          </div>
        </div>
      </section>

      <EmptyState v-if="groups.length === 0" :title="t.noPhotos" :description="t.noPhotosDesc" icon="photos" />

      <div v-if="nextBefore" class="load-more">
        <LoadingSkeletonPhotos v-if="loadingMore" variant="more" />
        <button v-else type="button" class="btn block" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? t.loading : t.loadMore }}
        </button>
      </div>
    </div>

    <div
      v-else
      id="photos-panel-albums"
      class="albums-container"
      role="tabpanel"
      aria-labelledby="photos-tab-albums"
      tabindex="0"
    >
      <form class="album-form" @submit.prevent="createAlbum">
        <label class="field album-field">
          <span class="sr-only">{{ t.newAlbumName }}</span>
          <input v-model="newAlbumName" type="text" :placeholder="t.newAlbumName" autocomplete="off" />
        </label>
        <button class="btn accent" type="submit" :disabled="!newAlbumName.trim()">{{ t.create }}</button>
      </form>

      <div v-if="albums.length" class="albums-grid">
        <article v-for="(album, index) in albums" :key="album.id" class="album-card appear" :style="{ animationDelay: cellDelay(index) }">
          <RouterLink class="album-primary" :to="`/photos/albums/${album.id}`" :aria-label="`${album.name}, ${album.itemCount} ${t.photosTitle.toLowerCase()}`">
            <div class="album-cover" aria-hidden="true">
              <img v-if="album.coverUrl" :src="album.coverUrl" alt="" loading="lazy" />
              <Icon v-else-if="album.coverFileId" name="video" :size="24" />
              <Icon v-else name="photos" :size="24" />
            </div>
            <div class="album-text-col">
              <span class="album-title">{{ album.name }}</span>
              <span class="album-sub">{{ album.itemCount }} {{ t.photosTitle.toLowerCase() }}</span>
            </div>
          </RouterLink>
          <button class="btn icon-only album-more-btn" type="button" :aria-label="`${t.albumMenu}: ${album.name}`" @click="openAlbumActions(album)">
            <Icon name="more" :size="18" />
          </button>
        </article>
      </div>
      <EmptyState v-else compact :title="t.noAlbums" :description="t.noAlbumsDesc" icon="photos" />
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
}

.pull-icon.spin { animation: spin 800ms linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.tabs-header {
  margin-bottom: var(--space-md);
  border-bottom: 1px solid var(--hairline);
  padding-bottom: 4px;
}

.tabs-pill-list { display: flex; align-items: center; gap: 16px; }
.tab-pill {
  background: transparent;
  border: none;
  padding: 8px 0;
  min-height: 44px;
  font-size: 15px;
  font-weight: 600;
  color: var(--muted);
  cursor: pointer;
  position: relative;
  display: inline-flex;
  align-items: center;
}
.tab-pill.active { color: var(--ink); }
.tab-pill:focus-visible { outline: 2px solid var(--accent); outline-offset: 3px; border-radius: 4px; }
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

.timeline-container:focus-visible,
.albums-container:focus-visible { outline: 2px solid var(--accent); outline-offset: 4px; border-radius: var(--radius-sm); }
.timeline-date-group { margin-bottom: var(--space-lg); }
.timeline-sticky-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: var(--canvas);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  margin-bottom: 8px;
}
.timeline-date-title { font-size: 15px; font-weight: 600; color: var(--ink); margin: 0; }
.timeline-count-badge { font-size: 12px; font-weight: 500; color: var(--muted); }
.grid.photos { display: grid; grid-template-columns: repeat(3, 1fr); gap: 4px; }
.photo-cell { position: relative; min-width: 0; }
.photo-more-btn {
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
.photo-cell:hover .photo-more-btn,
.photo-more-btn:focus-visible { opacity: 1; }
.photo-more-btn:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }

.albums-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; margin-top: var(--space-md); }
.album-card {
  position: relative;
  display: flex;
  flex-direction: column;
  border-radius: var(--radius-lg, 16px);
  background: var(--surface-card, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--hairline);
  overflow: hidden;
  transition: transform 0.15s ease, border-color 0.15s ease;
}
.album-card:focus-within { border-color: var(--accent); }
.album-primary { color: inherit; text-decoration: none; min-width: 0; }
.album-primary:focus-visible { outline: 2px solid var(--accent); outline-offset: -2px; border-radius: inherit; }
.album-cover {
  width: 100%;
  aspect-ratio: 16 / 10;
  background: var(--surface-card, rgba(255, 255, 255, 0.08));
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  overflow: hidden;
}
.album-cover img { width: 100%; height: 100%; object-fit: cover; }
.album-text-col { min-width: 0; padding: 10px 56px 10px 12px; }
.album-title {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.album-sub { display: block; font-size: 12px; color: var(--muted); margin-top: 2px; }
.album-more-btn { position: absolute; right: 6px; bottom: 5px; color: var(--muted); flex-shrink: 0; min-width: 44px; min-height: 44px; }
.album-form { display: flex; gap: 8px; margin-bottom: var(--space-md); }
.album-field { flex: 1; margin-bottom: 0; min-width: 0; }
.load-more { margin-top: var(--space-md); }
.desktop-only { display: none; }

@media (hover: none), (pointer: coarse) {
  .photo-more-btn { opacity: 1; }
}
@media (min-width: 640px) { .grid.photos { grid-template-columns: repeat(4, 1fr); gap: 6px; } }
@media (min-width: 768px) {
  .albums-grid { grid-template-columns: repeat(3, 1fr); gap: 16px; }
  .desktop-only { display: block; }
}
@media (min-width: 1024px) {
  .grid.photos { grid-template-columns: repeat(6, 1fr); gap: 8px; }
  .albums-grid { grid-template-columns: repeat(4, 1fr); }
}
@media (prefers-reduced-motion: reduce) {
  .pull-refresh-bar,
  .photo-more-btn,
  .album-card { transition: none; }
  .pull-icon.spin { animation: none; }
}
</style>
