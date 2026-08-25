<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'

const props = defineProps<{
  open: boolean
  name: string
  mimeType: string
  url: string
}>()

const emit = defineEmits<{
  close: []
  download: []
}>()

const isImage = computed(() => props.mimeType.startsWith('image/'))
const isVideo = computed(() => props.mimeType.startsWith('video/'))
const dialogRef = ref<HTMLDialogElement | null>(null)
const captionsUrl = 'data:text/vtt;charset=utf-8,WEBVTT%0A'
const { t } = useI18n()

watch(
  () => props.open,
  (open) => {
    document.body.style.overflow = open ? 'hidden' : ''
    if (open && !dialogRef.value?.open) {
      dialogRef.value?.showModal()
      return
    }
    if (!open && dialogRef.value?.open) dialogRef.value.close()
  },
)

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <Transition name="page">
      <dialog
        v-show="open"
        ref="dialogRef"
        class="lightbox"
        :aria-label="name"
        @close="emit('close')"
        @keydown.esc="emit('close')"
      >
        <button class="lightbox-backdrop" type="button" :aria-label="t.closePreview" @click="emit('close')" />
        <div class="lightbox-panel">
          <div class="lightbox-bar">
            <p class="lightbox-title">{{ name }}</p>
            <button type="button" class="btn ghost" @click="emit('download')">{{ t.download }}</button>
            <button type="button" class="btn icon-only" :aria-label="t.closePreview" @click="emit('close')">
              <Icon name="close" :size="18" />
            </button>
          </div>
          <div class="lightbox-media">
            <img v-if="isImage" :src="url" :alt="name" />
            <video v-else-if="isVideo" :src="url" controls playsinline>
              <track kind="captions" label="No captions" srclang="en" :src="captionsUrl" default />
            </video>
            <div v-else class="lightbox-fallback">
              <Icon name="file" :size="36" />
              <p>{{ t.previewUnavailable }}</p>
            </div>
          </div>
        </div>
      </dialog>
    </Transition>
  </Teleport>
</template>

<style scoped>
.lightbox {
  position: fixed;
  inset: 0;
  z-index: 1100;
  display: none;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: var(--space-md);
  border: none;
  background: transparent;
}

.lightbox[open] {
  display: flex;
}

.lightbox::backdrop {
  background: transparent;
}

.lightbox-backdrop {
  position: absolute;
  inset: 0;
  border: none;
  background: var(--overlay);
  cursor: zoom-out;
}

.lightbox-panel {
  position: relative;
  z-index: 1;
  width: min(100%, 980px);
  max-height: 92vh;
  max-height: 92dvh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: var(--radius-xl);
  background: var(--canvas);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.28);
}

.lightbox-bar {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-sm);
  border-bottom: 1px solid var(--hairline);
}

.lightbox-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  overflow: hidden;
  color: var(--ink);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lightbox-media {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 320px;
  background: var(--surface-soft);
}

.lightbox-media img,
.lightbox-media video {
  display: block;
  max-width: 100%;
  max-height: calc(92vh - 72px);
  object-fit: contain;
}

.lightbox-fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  color: var(--muted);
}

@media (max-width: 767px) {
  .lightbox {
    padding: 0;
  }

  .lightbox-panel {
    width: 100%;
    height: 100%;
    max-height: none;
    border-radius: 0;
  }

  .lightbox-media {
    flex: 1;
  }
}
</style>
