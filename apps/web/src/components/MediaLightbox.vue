<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'

const props = withDefaults(
  defineProps<{
    open: boolean
    name: string
    mimeType: string
    url: string
    hasNext?: boolean
    hasPrev?: boolean
  }>(),
  {
    hasNext: false,
    hasPrev: false,
  },
)

const emit = defineEmits<{
  close: []
  download: []
  next: []
  prev: []
}>()

const isImage = computed(() => props.mimeType.startsWith('image/'))
const isVideo = computed(() => props.mimeType.startsWith('video/'))
const isPdf = computed(() => props.mimeType === 'application/pdf')
const isAudio = computed(() => props.mimeType.startsWith('audio/'))

const dialogRef = ref<HTMLDialogElement | null>(null)
const videoRef = ref<HTMLVideoElement | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)

function pauseMedia() {
  if (videoRef.value) videoRef.value.pause()
  if (audioRef.value) audioRef.value.pause()
}

const captionsUrl = 'data:text/vtt;charset=utf-8,WEBVTT%0A'
const { t } = useI18n()

// Touch gestures & zoom
const scale = ref(1)
let baseScale = 1
const translateX = ref(0)
const translateY = ref(0)
const isTouching = ref(false)

let touchStartX = 0
let touchStartY = 0
let initialPinchDistance = 0
let lastTapTime = 0

function onTouchStart(e: TouchEvent) {
  if (e.touches.length === 2) {
    const [t1, t2] = [e.touches[0], e.touches[1]]
    if (t1 && t2) {
      initialPinchDistance = Math.hypot(t2.clientX - t1.clientX, t2.clientY - t1.clientY)
      baseScale = scale.value
    }
    return
  }

  if (e.touches.length === 1 && e.touches[0]) {
    touchStartX = e.touches[0].clientX
    touchStartY = e.touches[0].clientY
    isTouching.value = true

    // Double tap to zoom
    const now = Date.now()
    if (now - lastTapTime < 300 && isImage.value) {
      scale.value = scale.value > 1 ? 1 : 2.2
      translateX.value = 0
      translateY.value = 0
      lastTapTime = 0
      return
    }
    lastTapTime = now
  }
}

function onTouchMove(e: TouchEvent) {
  if (e.touches.length === 2 && e.touches[0] && e.touches[1] && initialPinchDistance > 0 && isImage.value) {
    const dist = Math.hypot(
      e.touches[1].clientX - e.touches[0].clientX,
      e.touches[1].clientY - e.touches[0].clientY,
    )
    scale.value = Math.min(Math.max(1, baseScale * (dist / initialPinchDistance)), 3)
    if (e.cancelable) e.preventDefault()
    return
  }

  if (e.touches.length === 1 && e.touches[0] && isTouching.value) {
    const dx = e.touches[0].clientX - touchStartX
    const dy = e.touches[0].clientY - touchStartY

    if (scale.value > 1) {
      // Pan when zoomed
      translateX.value = dx
      translateY.value = dy
      if (e.cancelable) e.preventDefault()
    } else {
      if (Math.abs(dy) > Math.abs(dx) && dy > 0) {
        translateY.value = dy
        if (e.cancelable) e.preventDefault()
      } else if (Math.abs(dx) > Math.abs(dy)) {
        translateX.value = dx * 0.7
        if (e.cancelable) e.preventDefault()
      }
    }
  }
}

function onTouchEnd() {
  isTouching.value = false
  initialPinchDistance = 0

  if (scale.value === 1) {
    if (translateY.value > 90) {
      emit('close')
    } else if (translateX.value < -50 && props.hasNext) {
      emit('next')
    } else if (translateX.value > 50 && props.hasPrev) {
      emit('prev')
    }
  }

  translateX.value = 0
  translateY.value = 0
}

function onKeydown(event: KeyboardEvent) {
  const target = event.target as HTMLElement | null
  const isMediaControl = target?.tagName === 'VIDEO' || target?.tagName === 'AUDIO' || target?.tagName === 'INPUT'
  if (isMediaControl) return

  if (event.key === 'ArrowRight' && props.hasNext) {
    emit('next')
  } else if (event.key === 'ArrowLeft' && props.hasPrev) {
    emit('prev')
  }
}

watch(
  () => props.open,
  (open) => {
    document.body.style.overflow = open ? 'hidden' : ''
    scale.value = 1
    translateX.value = 0
    translateY.value = 0
    if (open) {
      window.addEventListener('keydown', onKeydown)
      if (!dialogRef.value?.open) {
        dialogRef.value?.showModal()
      }
    } else {
      pauseMedia()
      window.removeEventListener('keydown', onKeydown)
      if (dialogRef.value?.open) dialogRef.value.close()
    }
  },
)

watch(
  () => props.url,
  () => {
    pauseMedia()
    scale.value = 1
    baseScale = 1
    translateX.value = 0
    translateY.value = 0
  },
)

onUnmounted(() => {
  pauseMedia()
  document.body.style.overflow = ''
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="lightbox">
      <dialog
        v-show="open"
        ref="dialogRef"
        class="lightbox"
        :aria-label="name"
        @close.prevent="emit('close')"
        @keydown.esc="emit('close')"
      >
        <button class="lightbox-backdrop" type="button" :aria-label="t.closePreview" @click="emit('close')" />
        <div class="lightbox-panel">
          <div class="lightbox-bar">
            <button
              v-if="hasPrev"
              type="button"
              class="btn icon-only nav-btn"
              title="Previous"
              aria-label="Previous item"
              @click="emit('prev')"
            >
              <Icon name="arrow-left" :size="18" />
            </button>
            <p class="lightbox-title">{{ name }}</p>
            <button
              v-if="hasNext"
              type="button"
              class="btn icon-only nav-btn"
              title="Next"
              aria-label="Next item"
              @click="emit('next')"
            >
              <Icon name="arrow-right" :size="18" />
            </button>
            <button type="button" class="btn ghost" @click="emit('download')">{{ t.download }}</button>
            <button type="button" class="btn icon-only" :aria-label="t.closePreview" @click="emit('close')" :title="t.closePreview">
              <Icon name="close" :size="18" />
            </button>
          </div>
          <div
            class="lightbox-media"
            @touchstart="onTouchStart"
            @touchmove="onTouchMove"
            @touchend="onTouchEnd"
            @touchcancel="onTouchEnd"
          >
            <div
              class="media-wrapper"
              :style="{
                transform: `translate(${translateX}px, ${translateY}px) scale(${scale})`,
                transition: isTouching ? 'none' : 'transform 200ms var(--ease-standard)',
              }"
            >
              <img v-if="isImage" :src="url" :alt="name" draggable="false" />
              <video v-else-if="isVideo" ref="videoRef" :src="url" controls playsinline>
                <track kind="captions" label="No captions" srclang="en" :src="captionsUrl" default />
              </video>
              <iframe v-else-if="isPdf" :src="url" class="pdf-frame" title="PDF preview" />
              <audio v-else-if="isAudio" ref="audioRef" :src="url" controls class="audio-player" />
              <div v-else class="lightbox-fallback">
                <Icon name="file" :size="36" />
                <p>{{ t.previewUnavailable }}</p>
              </div>
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
  overflow: hidden;
}

.media-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  transform-origin: center center;
}

.media-wrapper img,
.media-wrapper video {
  display: block;
  max-width: 100%;
  max-height: calc(92vh - 72px);
  object-fit: contain;
  user-select: none;
}

.pdf-frame {
  width: 100%;
  height: calc(92vh - 72px);
  min-height: 480px;
  border: none;
}

.audio-player {
  width: min(400px, 90%);
}

.nav-btn {
  flex-shrink: 0;
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

  .media-wrapper img,
  .media-wrapper video,
  .pdf-frame {
    max-height: calc(100dvh - 60px);
  }
}

.lightbox-enter-active,
.lightbox-leave-active {
  transition: opacity var(--duration-medium, 200ms) var(--ease-standard);
}

.lightbox-enter-from,
.lightbox-leave-to {
  opacity: 0;
}
</style>
