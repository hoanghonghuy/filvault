<script setup lang="ts">
import { ref } from 'vue'
import Icon from '@/components/AppIcon.vue'

const props = defineProps<{
  mimeType: string
  name: string
  thumbnailUrl?: string
}>()

const emit = defineEmits<{
  click: []
}>()

const failed = ref(false)
const isVideo = props.mimeType.startsWith('video/')

function handleError() {
  failed.value = true
}
</script>

<template>
  <button type="button" class="photo-thumb" :title="name" @click="emit('click')">
    <template v-if="thumbnailUrl && !failed">
      <img
        v-if="!isVideo"
        :src="thumbnailUrl"
        :alt="name"
        loading="lazy"
        decoding="async"
        @error="handleError"
      />
      <video
        v-else
        :src="thumbnailUrl"
        :aria-label="name"
        muted
        playsinline
        preload="metadata"
        @error="handleError"
      />
      <span v-if="isVideo" class="media-badge">
        <Icon name="video" :size="14" />
      </span>
    </template>
    <span v-else class="fallback" :class="{ video: isVideo }">
      <Icon :name="isVideo ? 'video' : 'image'" :size="28" />
      <span class="name">{{ name }}</span>
    </span>
  </button>
</template>

<style scoped>
.photo-thumb {
  position: relative;
  aspect-ratio: 1;
  width: 100%;
  overflow: hidden;
  padding: 0;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  cursor: pointer;
}

.photo-thumb:hover {
  border-color: var(--accent);
}

.photo-thumb:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

img,
video {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.media-badge {
  position: absolute;
  right: var(--space-xs);
  bottom: var(--space-xs);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-pill);
  background: rgba(17, 24, 39, 0.72);
  color: var(--on-ink);
}

.fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  gap: var(--space-xs);
  padding: var(--space-sm);
  color: var(--accent);
}

.fallback.video {
  color: var(--muted);
}

.name {
  display: -webkit-box;
  max-width: 100%;
  overflow: hidden;
  color: var(--muted);
  font-size: 0.75rem;
  line-height: 1.3;
  word-break: break-all;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
</style>
