<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
import type { DownloadURL, Timeline, TimelineItem } from '@/api/types'
import PhotoThumb from '@/components/PhotoThumb.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import LoadingSkeletonPhotos from '@/components/LoadingSkeletonPhotos.vue'

const ui = useUiStore()
const { t } = useI18n()
const groups = ref<Timeline['groups']>([])
const nextBefore = ref<string | undefined>()
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
const lightboxOpen = ref(false)
const lightboxUrl = ref('')
const mediaItem = ref<TimelineItem | null>(null)
const favoriteIds = ref<Set<string>>(new Set())

async function loadVideoTimeline(before?: string) {
  const params = new URLSearchParams({ type: 'video' })
  if (before) params.set('before', before)
  return api<Timeline>(`/photos/timeline/filter?${params}`)
}

async function loadFavorites() {
  const out = await api<{ files: Array<{ id: string }> }>('/files/favorites?limit=100')
  favoriteIds.value = new Set(out.files.map((file) => file.id))
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [data] = await Promise.all([loadVideoTimeline(), loadFavorites()])
    groups.value = data.groups
    nextBefore.value = data.nextBefore
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load videos')
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!nextBefore.value || loadingMore.value) return
  loadingMore.value = true
  error.value = ''
  try {
    const data = await loadVideoTimeline(nextBefore.value)
    const merged = [...groups.value]
    for (const group of data.groups) {
      const existing = merged.find((item) => item.date === group.date)
      if (existing) existing.items.push(...group.items)
      else merged.push(group)
    }
    groups.value = merged
    nextBefore.value = data.nextBefore
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load more videos')
  } finally {
    loadingMore.value = false
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
    error.value = formatApiError(e, 'View failed')
  }
}

async function download(item: TimelineItem) {
  error.value = ''
  try {
    const out = await api<DownloadURL>(`/files/${item.id}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, 'Download failed')
  }
}

async function toggleFavorite(item: TimelineItem) {
  const favorited = favoriteIds.value.has(item.id)
  error.value = ''
  try {
    await api(`/files/${item.id}/favorite`, { method: favorited ? 'DELETE' : 'PUT' })
    await loadFavorites()
  } catch (e) {
    error.value = formatApiError(e, 'Failed to update favorite')
  }
}

async function openActions(item: TimelineItem) {
  const favorited = favoriteIds.value.has(item.id)
  const action = await ui.openActionSheet(item.name, [
    { id: 'view', label: t.value.preview, icon: 'eye' },
    { id: 'download', label: t.value.download, icon: 'download' },
    { id: 'favorite', label: favorited ? t.value.removeFromFavorites : t.value.addToFavorites, icon: favorited ? 'star-filled' : 'star' },
  ])
  if (action === 'view') await openLightbox(item)
  if (action === 'download') await download(item)
  if (action === 'favorite') await toggleFavorite(item)
}

const allItems = computed(() => groups.value.flatMap((group) => group.items))
const currentIndex = computed(() => (mediaItem.value ? allItems.value.findIndex((item) => item.id === mediaItem.value?.id) : -1))
const hasPrevious = computed(() => currentIndex.value > 0)
const hasNext = computed(() => currentIndex.value >= 0 && currentIndex.value < allItems.value.length - 1)

async function previousMedia() {
  if (!hasPrevious.value) return
  const item = allItems.value[currentIndex.value - 1]
  if (item) await openLightbox(item)
}

async function nextMedia() {
  if (!hasNext.value) return
  const item = allItems.value[currentIndex.value + 1]
  if (item) await openLightbox(item)
}

onMounted(() => void load())
</script>

<template>
  <div class="video-page">
    <header class="video-header">
      <div>
        <h1 class="page-title">{{ t.filterTypeVideo }}</h1>
        <p class="video-subtitle">{{ t.photosTitle }}</p>
      </div>
      <RouterLink class="btn" to="/photos">{{ t.all }}</RouterLink>
    </header>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonPhotos v-if="loading" variant="initial" />

    <main v-else aria-labelledby="video-timeline-heading">
      <h2 id="video-timeline-heading" class="sr-only">{{ t.filterTypeVideo }}</h2>
      <section v-for="group in groups" :key="group.date" class="timeline-date-group" :aria-label="group.date">
        <div class="timeline-sticky-header">
          <h3 class="timeline-date-title">{{ group.date }}</h3>
          <span class="timeline-count-badge">{{ group.items.length }}</span>
        </div>
        <div class="video-grid">
          <article v-for="item in group.items" :key="item.id" class="video-cell">
            <PhotoThumb
              :mime-type="item.mimeType"
              :name="item.name"
              :thumbnail-url="item.thumbnailUrl"
              @click="openLightbox(item)"
            />
            <button type="button" class="more-btn" :aria-label="`${t.moreActions}: ${item.name}`" @click="openActions(item)">
              <Icon name="more" :size="18" />
            </button>
          </article>
        </div>
      </section>

      <EmptyState v-if="groups.length === 0" :title="t.filterTypeVideo" :description="t.noPhotosDesc" icon="video" />

      <div v-if="nextBefore" class="load-more">
        <LoadingSkeletonPhotos v-if="loadingMore" variant="more" />
        <button v-else type="button" class="btn block" :disabled="loadingMore" @click="loadMore">{{ t.loadMore }}</button>
      </div>
    </main>

    <MediaLightbox
      :open="lightboxOpen"
      :url="lightboxUrl"
      :mime-type="mediaItem?.mimeType ?? ''"
      :name="mediaItem?.name ?? ''"
      :has-prev="hasPrevious"
      :has-next="hasNext"
      @close="lightboxOpen = false"
      @prev="previousMedia"
      @next="nextMedia"
    />
  </div>
</template>

<style scoped>
.video-page { padding-bottom: calc(var(--space-xl) + env(safe-area-inset-bottom)); }
.video-header { display: flex; align-items: center; justify-content: space-between; gap: var(--space-md); margin-bottom: var(--space-lg); }
.page-title { margin: 0; }
.video-subtitle { margin: var(--space-2xs) 0 0; color: var(--text-muted); }
.timeline-date-group + .timeline-date-group { margin-top: var(--space-lg); }
.timeline-sticky-header { display: flex; align-items: center; gap: var(--space-xs); margin-bottom: var(--space-sm); }
.timeline-date-title { margin: 0; font-size: 0.95rem; }
.timeline-count-badge { color: var(--text-muted); font-size: 0.8rem; }
.video-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-xs); }
.video-cell { position: relative; min-width: 0; }
.more-btn { position: absolute; right: var(--space-xs); bottom: var(--space-xs); width: 44px; height: 44px; display: grid; place-items: center; border: 0; border-radius: 50%; background: var(--surface-card); color: var(--text); cursor: pointer; }
.more-btn:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.load-more { margin-top: var(--space-lg); }
@media (min-width: 768px) { .video-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); gap: var(--space-sm); } }
@media (min-width: 1200px) { .video-grid { grid-template-columns: repeat(6, minmax(0, 1fr)); } }
</style>
