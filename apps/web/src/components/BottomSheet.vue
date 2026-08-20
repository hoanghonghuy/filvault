<script setup lang="ts">
import { nextTick, onUnmounted, ref, useId, watch } from 'vue'

const props = defineProps<{
  open: boolean
  title?: string
}>()

const emit = defineEmits<{
  close: []
}>()

const titleId = useId()
const panelRef = ref<HTMLElement | null>(null)
let previousFocus: HTMLElement | null = null

function onBackdropClick() {
  emit('close')
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
  }
}

function unlockPage() {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', onKeydown)
  previousFocus?.focus()
  previousFocus = null
}

watch(
  () => props.open,
  async (open) => {
    if (open) {
      previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
      document.body.style.overflow = 'hidden'
      document.addEventListener('keydown', onKeydown)
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
    <Transition name="sheet" @after-leave="unlockPage">
      <div v-if="open" class="sheet-root" role="presentation">
        <div class="sheet-backdrop" @click="onBackdropClick" />
        <div
          ref="panelRef"
          class="sheet-panel"
          role="dialog"
          aria-modal="true"
          tabindex="-1"
          :aria-labelledby="title ? titleId : undefined"
          :aria-label="title ? undefined : 'Dialog'"
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
  background: var(--hairline);
  border-radius: var(--radius-pill);
}

.sheet-title {
  margin: 0 0 var(--space-md);
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--ink);
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
  transition: transform var(--motion-exit) var(--ease-exit);
}

.sheet-enter-from .sheet-panel,
.sheet-leave-to .sheet-panel {
  transform: translateY(100%);
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
      opacity var(--motion-exit) var(--ease-exit);
  }

  .sheet-enter-from .sheet-panel,
  .sheet-leave-to .sheet-panel {
    transform: translateY(16px);
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
