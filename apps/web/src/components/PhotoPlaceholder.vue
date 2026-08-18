<script setup lang="ts">
defineProps<{ mimeType: string; name: string }>()

const emit = defineEmits<{ click: [] }>()

const isVideo = (mime: string) => mime.startsWith('video/')
</script>

<template>
  <button type="button" class="placeholder" :aria-label="name" :title="name" @click="emit('click')">
    <span class="badge" :class="{ video: isVideo(mimeType) }">
      {{ isVideo(mimeType) ? 'Video' : 'Photo' }}
    </span>
    <span class="name">{{ name }}</span>
  </button>
</template>

<style scoped>
.placeholder {
  aspect-ratio: 1;
  width: 100%;
  border: 1px dashed var(--hairline);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 0.5rem;
  background: var(--surface-soft);
  text-align: center;
  cursor: pointer;
}

.placeholder:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

@media (hover: hover) {
  .placeholder:hover {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
}

.badge {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--accent);
  background: var(--accent-soft);
  padding: 0.2rem 0.45rem;
  border-radius: var(--radius-sm);
}

.badge.video {
  color: var(--ink);
  background: var(--surface-card);
}

.name {
  font-size: 0.75rem;
  color: var(--muted);
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
