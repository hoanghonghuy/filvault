<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'

defineProps<{
  selectedCount: number
  totalCount: number
}>()

const emit = defineEmits<{
  close: []
  'select-all': []
  'clear-selection': []
  delete: []
  download: []
  move: []
  favorite: []
}>()

const { t } = useI18n()

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <aside
    class="batch-bar"
    role="toolbar"
    aria-label="Selection actions"
  >
      <div class="batch-left">
        <button
          type="button"
          class="btn-icon"
          :aria-label="t.closeSelection"
          @click="emit('close')"
        >
          <Icon name="close" :size="20" />
        </button>
        <span class="batch-count" aria-live="polite">
          {{ selectedCount }} {{ t.selected }}
        </span>
      </div>

      <div class="batch-actions">
        <button
          v-if="totalCount > 0 && selectedCount < totalCount"
          type="button"
          class="btn-text"
          @click="emit('select-all')"
        >
          {{ t.selectAll }}
        </button>
        <button
          v-else-if="totalCount > 0"
          type="button"
          class="btn-text"
          @click="emit('clear-selection')"
        >
          {{ t.deselectAll }}
        </button>

        <div class="divider" aria-hidden="true" />

        <button
          type="button"
          class="btn-icon"
          :title="t.downloadSelected"
          :aria-label="t.downloadSelected"
          :disabled="selectedCount === 0"
          @click="emit('download')"
        >
          <Icon name="download" :size="20" />
        </button>

        <button
          type="button"
          class="btn-icon"
          :title="t.addToFavorites"
          :aria-label="t.addToFavorites"
          :disabled="selectedCount === 0"
          @click="emit('favorite')"
        >
          <Icon name="star" :size="20" />
        </button>

        <button
          type="button"
          class="btn-icon"
          :title="t.moveSelected"
          :aria-label="t.moveSelected"
          :disabled="selectedCount === 0"
          @click="emit('move')"
        >
          <Icon name="move" :size="20" />
        </button>

        <button
          type="button"
          class="btn-icon danger"
          :title="t.moveToTrash"
          :aria-label="t.moveToTrash"
          :disabled="selectedCount === 0"
          @click="emit('delete')"
        >
          <Icon name="trash" :size="20" />
        </button>
      </div>
    </aside>
</template>

<style scoped>
.batch-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  padding: var(--space-xs) var(--space-md) calc(var(--space-xs) + env(safe-area-inset-bottom));
  background: var(--canvas);
  border-top: 1px solid var(--hairline);
  box-shadow: 0 -4px 16px rgba(17, 24, 39, 0.08);
}

@media (min-width: 768px) {
  .batch-bar {
    left: 50%;
    right: auto;
    bottom: var(--space-lg);
    transform: translateX(-50%);
    min-width: 440px;
    max-width: 600px;
    border-radius: var(--radius-pill);
    border: 1px solid var(--hairline);
    padding: var(--space-xs) var(--space-lg);
    box-shadow: 0 8px 30px rgba(17, 24, 39, 0.12);
  }
}

.batch-left {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.batch-count {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--ink);
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: var(--space-xxs);
}

.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--touch-min);
  height: var(--touch-min);
  padding: 0;
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--ink);
  cursor: pointer;
  transition: background var(--duration-short) var(--ease-standard);
}

.btn-icon:hover {
  background: var(--surface-soft);
}

.btn-icon:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-icon.danger {
  color: var(--danger);
}

.btn-icon.danger:hover {
  background: var(--danger-soft);
}

.btn-icon:active,
.btn-text:active {
  transform: scale(0.97);
}

.btn-text {
  min-height: var(--touch-min);
  padding: 0 var(--space-xs);
  border: none;
  background: transparent;
  color: var(--accent);
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
}

.divider {
  width: 1px;
  height: 20px;
  margin: 0 var(--space-xxs);
  background: var(--hairline);
}

.action-bar-enter-active,
.action-bar-leave-active {
  transition: transform var(--duration-medium) var(--ease-standard),
    opacity var(--duration-medium) var(--ease-standard);
}

.action-bar-enter-from,
.action-bar-leave-to {
  opacity: 0;
  transform: translateY(100%);
}

@media (min-width: 768px) {
  .action-bar-enter-from,
  .action-bar-leave-to {
    transform: translate(-50%, 20px);
  }
}
</style>
