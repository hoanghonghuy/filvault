<script lang="ts">
// Shared across all mounted BottomSheet instances: Escape (and focus trap)
// must only affect the topmost open sheet, not every open sheet at once.
const openSheets: number[] = []
let sheetCounter = 0
let lockedCount = 0
</script>

<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, useId, watch } from 'vue'
import { useI18n } from '@/lib/i18n'

const props = defineProps<{
  open: boolean
  title?: string
}>()

const emit = defineEmits<{
  close: []
  'after-leave': []
}>()

const { locale } = useI18n()
const fallbackDialogLabel = computed(() => (locale.value === 'vi' ? 'Hộp thoại' : 'Dialog'))
const titleId = useId()
const panelRef = ref<HTMLElement | null>(null)
const sheetId = sheetCounter++
let previousFocus: HTMLElement | null = null
let released = true

const dragOffset = ref(0)
const isDragging = ref(false)
let touchStartY = 0
let touchStartTime = 0
let canDrag = false

function onTouchStart(event: TouchEvent) {
  if (!panelRef.value) return
  const touch = event.touches[0]
  if (!touch) return
  const target = event.target as HTMLElement | null
  const isHandle = target?.closest('.sheet-handle') !== null
  const isTop = panelRef.value.scrollTop <= 0
  if (isHandle || isTop) {
    canDrag = true
    touchStartY = touch.clientY
    touchStartTime = Date.now()
    isDragging.value = true
  } else {
    canDrag = false
  }
}

function onTouchMove(event: TouchEvent) {
  if (!canDrag) return
  const touch = event.touches[0]
  if (!touch) return
  const deltaY = touch.clientY - touchStartY
  if (deltaY > 0) {
    dragOffset.value = deltaY
    if (event.cancelable) event.preventDefault()
  } else {
    dragOffset.value = 0
  }
}

function onTouchEnd() {
  if (!canDrag) return
  isDragging.value = false
  canDrag = false
  const elapsed = Date.now() - touchStartTime
  const velocity = dragOffset.value / Math.max(elapsed, 1)
  if (dragOffset.value > 90 || (dragOffset.value > 40 && velocity > 0.5)) {
    emit('close')
  } else {
    dragOffset.value = 0
  }
}

function onBackdropClick() {
  emit('close')
}

function isTopSheet() {
  return openSheets[openSheets.length - 1] === sheetId
}

function onKeydown(event: KeyboardEvent) {
  if (!isTopSheet()) return
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key === 'Tab') trapFocus(event)
}

function trapFocus(event: KeyboardEvent) {
  const panel = panelRef.value
  if (!panel) return
  const focusables = panel.querySelectorAll<HTMLElement>(
    'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
  )
  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  if (!first || !last) {
    event.preventDefault()
    panel.focus()
    return
  }
  const active = document.activeElement
  if (event.shiftKey && (active === first || active === panel)) {
    event.preventDefault()
    last.focus()
    return
  }
  if (!event.shiftKey && active === last) {
    event.preventDefault()
    first.focus()
  }
}

function lockPage() {
  if (!released) {
    unlockPage()
  }
  released = false
  openSheets.push(sheetId)
  if (lockedCount === 0) document.body.style.overflow = 'hidden'
  lockedCount++
  previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
  document.addEventListener('keydown', onKeydown)
}

function unlockPage() {
  if (released) return
  released = true
  const idx = openSheets.indexOf(sheetId)
  if (idx !== -1) openSheets.splice(idx, 1)
  lockedCount--
  if (lockedCount <= 0) {
    lockedCount = 0
    document.body.style.overflow = ''
  }
  document.removeEventListener('keydown', onKeydown)
  previousFocus?.focus()
  previousFocus = null
}

function onAfterLeave() {
  dragOffset.value = 0
  isDragging.value = false
  unlockPage()
  emit('after-leave')
}

watch(
  () => props.open,
  async (open) => {
    if (open) {
      dragOffset.value = 0
      isDragging.value = false
      lockPage()
      await nextTick()
      panelRef.value?.focus()
    }
  },
  { immediate: true },
)

onUnmounted(unlockPage)
</script>

<template>
  <Teleport to="body">
    <Transition name="sheet" @after-leave="onAfterLeave">
      <div v-if="open" class="sheet-root" role="presentation">
        <div class="sheet-backdrop" @click="onBackdropClick" />
        <div
          ref="panelRef"
          class="sheet-panel"
          :class="{ 'sheet-dragging': isDragging }"
          :style="{
            transform: dragOffset > 0 ? `translateY(${dragOffset}px)` : undefined,
            transition: isDragging ? 'none' : undefined,
          }"
          role="dialog"
          aria-modal="true"
          tabindex="-1"
          :aria-labelledby="title ? titleId : undefined"
          :aria-label="title ? undefined : fallbackDialogLabel"
          @touchstart="onTouchStart"
          @touchmove="onTouchMove"
          @touchend="onTouchEnd"
          @touchcancel="onTouchEnd"
        >
          <div class="sheet-handle" aria-hidden="true" />
          <h2 v-if="title" :id="titleId" class="sheet-title">{{ title }}</h2>
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.sheet-root {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.sheet-backdrop {
  position: absolute;
  inset: 0;
  background: var(--overlay);
}



.sheet-panel {
  position: relative;
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  max-height: 90dvh;
  overflow-y: auto;
  background: var(--canvas);
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  padding: var(--space-sm) var(--space-md) max(var(--space-md), env(safe-area-inset-bottom));
}

.sheet-panel:focus {
  outline: none;
}

.sheet-panel:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}

@media (min-width: 768px) {
  .sheet-root {
    align-items: center;
    padding: var(--space-md);
  }

  .sheet-panel {
    border-radius: var(--radius-xl);
    max-height: 85vh;
  }
}

.sheet-handle {
  width: 32px;
  height: 4px;
  margin: 0 auto var(--space-sm);
  background: var(--muted-soft);
  border-radius: var(--radius-pill);
  cursor: grab;
  touch-action: none;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
}

.sheet-panel {
  -webkit-tap-highlight-color: transparent;
}

.sheet-panel.sheet-dragging {
  user-select: none;
}

.sheet-title {
  margin: 0 0 var(--space-md);
  font-size: 1rem;
  font-weight: 600;
  color: var(--ink);
}

.sheet-enter-active {
  transition: opacity var(--duration-long) var(--ease-standard);
}

.sheet-leave-active {
  transition: opacity var(--duration-medium) var(--ease-emphasized-accelerate);
}

.sheet-enter-active .sheet-backdrop {
  transition: opacity var(--motion-enter) var(--ease-enter);
}

.sheet-leave-active .sheet-backdrop {
  transition: opacity var(--motion-exit) var(--ease-exit);
}

.sheet-enter-from .sheet-backdrop,
.sheet-leave-to .sheet-backdrop {
  opacity: 0;
}

.sheet-enter-active .sheet-panel {
  transition: transform var(--motion-enter) var(--ease-enter);
}

.sheet-leave-active .sheet-panel {
  transition: transform var(--motion-exit) var(--ease-exit) !important;
}

.sheet-enter-from .sheet-panel {
  transform: translateY(100%);
}

.sheet-leave-to .sheet-panel {
  transform: translateY(100%) !important;
}

@media (min-width: 768px) {
  .sheet-enter-active .sheet-panel,
  .sheet-leave-active .sheet-panel {
    transition:
      transform var(--motion-enter) var(--ease-enter),
      opacity var(--motion-enter) var(--ease-enter);
  }

  .sheet-leave-active .sheet-panel {
    transition:
      transform var(--motion-exit) var(--ease-exit),
      opacity var(--motion-exit) var(--ease-exit) !important;
  }

  .sheet-enter-from .sheet-panel {
    transform: translateY(16px);
    opacity: 0;
  }

  .sheet-leave-to .sheet-panel {
    transform: translateY(16px) !important;
    opacity: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .sheet-enter-active .sheet-backdrop,
  .sheet-leave-active .sheet-backdrop,
  .sheet-enter-active .sheet-panel,
  .sheet-leave-active .sheet-panel {
    transition: none;
  }

  .sheet-enter-from .sheet-panel,
  .sheet-leave-to .sheet-panel {
    transform: none;
  }
}
</style>
