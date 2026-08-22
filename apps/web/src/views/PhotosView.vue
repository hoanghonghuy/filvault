<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import PhotoThumb from '@/components/PhotoThumb.vue'
import PhotoMediaSheet from '@/components/PhotoMediaSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import Icon from '@/components/AppIcon.vue'
import LoadingSkeletonPhotos from '@/components/LoadingSkeletonPhotos.vue'
import { cellDelay } from '@/lib/motion'
import type { Album, DownloadURL, Timeline, TimelineItem } from '@/api/types'

const router = useRouter()
const ui = useUiStore()

const groups = ref<Timeline['groups']>([])
const nextBefore = ref<string | undefined>()
const albums = ref<Album[]>([])
const newAlbumName = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')

const mediaOpen = ref(false)
const mediaItem = ref<TimelineItem | null>(null)

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
    groups.value = [...groups.value, ...data.groups]
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
  if (action === 'open') await router.push(`/photos/albums/${album.id}`)
  if (action === 'rename') await renameAlbum(album)
  if (action === 'delete') await deleteAlbum(album.id, album.name)
}

async function renameAlbum(album: Album) {
  const name = await ui.prompt({ title: 'Rename album', label: 'Name', initialValue: album.name })
  if (!name || name === album.name) return
  error.value = ''
  try {
    await api(`/photos/albums/${album.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ name }),
    })
    ui.showToast('Album renamed')
    await load()
  } catch (e) {
    error.value = formatApiError(e, 'Rename failed')
  }
}

async function deleteAlbum(id: string, name: string) {
  const ok = await ui.confirm({
    title: 'Delete album?',
    message: `"${name}" will be removed. Your files stay in My Files.`,
    confirmLabel: 'Delete album',
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

onMounted(load)
</script>

<template>
  <div class="photos-page">
    <h1 class="page-title desktop-only">Photos</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <LoadingSkeletonPhotos v-if="loading" variant="initial" />
    <div v-else>
      <section class="section" aria-labelledby="albums-heading">
        <h2 id="albums-heading" class="section-title">Albums</h2>
        <form class="album-form" @submit.prevent="createAlbum">
          <label class="field album-field">
            <span class="sr-only">New album name</span>
            <input v-model="newAlbumName" type="text" placeholder="New album name" autocomplete="off" />
          </label>
          <button class="btn ink" type="submit" :disabled="!newAlbumName.trim()">Create</button>
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
            <button class="btn icon-only" type="button" aria-label="Album actions" @click.stop="openAlbumActions(album)">
              <Icon name="more" :size="18" />
            </button>
          </div>
        </div>
        <EmptyState
          v-else
          compact
          title="No albums yet"
          description="Create an album to group photos and videos."
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
        title="No photos yet"
        description="Upload images or videos in My Files — they will appear here automatically."
        icon="photos"
      />

      <div v-if="nextBefore" class="load-more">
        <LoadingSkeletonPhotos v-if="loadingMore" variant="more" />
        <button v-else type="button" class="btn block" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? 'Loading…' : 'Load more' }}
        </button>
      </div>

      <PhotoMediaSheet
        :open="mediaOpen"
        :name="mediaItem?.name ?? ''"
        @view="viewMedia"
        @download="downloadMedia"
        @close="mediaOpen = false"
      />
    </div>
  </div>
</template>

<style scoped>
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
