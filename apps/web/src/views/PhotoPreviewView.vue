<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import MediaLightbox from '@/components/MediaLightbox.vue'
import Icon from '@/components/AppIcon.vue'
import type { DownloadURL } from '@/api/types'

interface OwnedFile {
  id: string
  name: string
  mimeType: string
  sizeBytes: number
  status: string
  createdAt: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref('')
const item = ref<OwnedFile | null>(null)
const previewUrl = ref('')
const lightboxOpen = ref(false)

function isPhotosMedia(file: OwnedFile) {
  return file.status === 'READY' && (file.mimeType.startsWith('image/') || file.mimeType.startsWith('video/'))
}

async function loadPreview() {
  loading.value = true
  error.value = ''
  const id = typeof route.params.id === 'string' ? route.params.id : ''
  if (!id) {
    error.value = 'This photo or video is unavailable.'
    loading.value = false
    return
  }

  try {
    const file = await api<OwnedFile>(`/files/${encodeURIComponent(id)}`)
    if (!isPhotosMedia(file)) {
      error.value = 'This item cannot be previewed in Photos.'
      return
    }
    const out = await api<DownloadURL>(`/files/${encodeURIComponent(id)}/download`)
    item.value = file
    previewUrl.value = out.downloadUrl
    lightboxOpen.value = true
  } catch {
    error.value = 'This photo or video is unavailable.'
  } finally {
    loading.value = false
  }
}

function closePreview() {
  lightboxOpen.value = false
  void router.replace('/photos')
}

function download() {
  if (!previewUrl.value) return
  window.open(previewUrl.value, '_blank', 'noopener')
}

onMounted(() => {
  void loadPreview()
})
</script>

<template>
  <section class="photo-preview-page" aria-labelledby="photo-preview-title">
    <RouterLink class="back-link" to="/photos">
      <Icon name="arrow-left" :size="18" />
      <span>Photos</span>
    </RouterLink>

    <div v-if="loading" class="preview-state" aria-busy="true" aria-live="polite">
      <div class="preview-skeleton skeleton" aria-hidden="true"></div>
      <p id="photo-preview-title">Loading preview…</p>
    </div>

    <div v-else-if="error" class="preview-state" role="alert">
      <Icon name="photos" :size="36" />
      <h1 id="photo-preview-title">Preview unavailable</h1>
      <p>{{ error }}</p>
      <RouterLink class="btn accent" to="/photos">Open Photos</RouterLink>
    </div>

    <h1 v-else id="photo-preview-title" class="sr-only">{{ item?.name }}</h1>

    <MediaLightbox
      :open="lightboxOpen"
      :name="item?.name ?? ''"
      :mime-type="item?.mimeType ?? ''"
      :url="previewUrl"
      :has-next="false"
      :has-prev="false"
      @download="download"
      @close="closePreview"
    />
  </section>
</template>

<style scoped>
.photo-preview-page {
  min-height: min(70vh, 720px);
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.back-link {
  min-height: 44px;
  width: fit-content;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--ink);
  text-decoration: none;
  border-radius: var(--radius-sm);
}

.back-link:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 3px;
}

.preview-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  text-align: center;
  color: var(--muted);
  padding: var(--space-xl) var(--space-md);
}

.preview-state h1,
.preview-state p {
  margin: 0;
}

.preview-state h1 {
  color: var(--ink);
  font-size: 1.25rem;
}

.preview-skeleton {
  width: min(100%, 360px);
  aspect-ratio: 4 / 3;
  border-radius: var(--radius-lg);
}
</style>
