<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'
import { useUiStore } from '@/stores/ui'

const props = defineProps<{
  selectedCount: number
  selectedFileCount: number
  selectedFolderCount: number
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
  vault: []
}>()

const { t } = useI18n()
const ui = useUiStore()

const hasFiles = computed(() => props.selectedFileCount > 0)
const hasFolders = computed(() => props.selectedFolderCount > 0)
const isMixed = computed(() => hasFiles.value && hasFolders.value)
const canUseFileActions = computed(() => hasFiles.value)
const hasSelection = computed(() => props.selectedCount > 0)

const countDetail = computed(() => {
  if (!isMixed.value) return ''
  return t.value.selectionFilesFolders
    .replace('{files}', String(props.selectedFileCount))
    .replace('{folders}', String(props.selectedFolderCount))
})

function fileScopedLabel(singleKey: keyof typeof t.value, pluralKey: keyof typeof t.value): string {
  if (!hasFiles.value) return t.value.filesOnlyDisabled
  if (props.selectedFolderCount > 0) {
    const template = t.value[pluralKey] as string
    return template.replace('{n}', String(props.selectedFileCount))
  }
  return t.value[singleKey] as string
}

const downloadLabel = computed(() => fileScopedLabel('downloadSelected', 'downloadNFiles'))
const favoriteLabel = computed(() => fileScopedLabel('addToFavorites', 'favoriteNFiles'))
const vaultLabel = computed(() => fileScopedLabel('vaultMoveToVault', 'vaultNFiles'))

async function openMoreActions() {
  if (!canUseFileActions.value) return
  const action = await ui.openActionSheet(t.value.moreActions, [
    { id: 'download', label: downloadLabel.value, icon: 'download' },
    { id: 'favorite', label: favoriteLabel.value, icon: 'star' },
    { id: 'vault', label: vaultLabel.value, icon: 'lock' },
  ])
  if (action === 'download') emit('download')
  else if (action === 'favorite') emit('favorite')
  else if (action === 'vault') emit('vault')
}

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
    :aria-label="t.selectItems"
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
        <span class="batch-count-main">{{ selectedCount }} {{ t.selected }}</span>
        <span v-if="countDetail" class="batch-count-detail">{{ countDetail }}</span>
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
        class="btn-icon file-only-action"
        :title="downloadLabel"
        :aria-label="downloadLabel"
        :disabled="!canUseFileActions"
        @click="emit('download')"
      >
        <Icon name="download" :size="20" />
      </button>

      <button
        type="button"
        class="btn-icon file-only-action"
        :title="favoriteLabel"
        :aria-label="favoriteLabel"
        :disabled="!canUseFileActions"
        @click="emit('favorite')"
      >
        <Icon name="star" :size="20" />
      </button>

      <button
        type="button"
        class="btn-icon"
        :title="t.moveSelected"
        :aria-label="t.moveSelected"
        :disabled="!hasSelection"
        @click="emit('move')"
      >
        <Icon name="move" :size="20" />
      </button>

      <button
        type="button"
        class="btn-icon file-only-action"
        :title="vaultLabel"
        :aria-label="vaultLabel"
        :disabled="!canUseFileActions"
        @click="emit('vault')"
      >
        <Icon name="lock" :size="20" />
      </button>

      <button
        type="button"
        class="btn-icon danger"
        :title="t.moveToTrash"
        :aria-label="t.moveToTrash"
        :disabled="!hasSelection"
        @click="emit('delete')"
      >
        <Icon name="trash" :size="20" />
      </button>

      <button
        v-if="canUseFileActions"
        type="button"
        class="btn-icon more-menu-btn"
        :title="t.moreActions"
        :aria-label="t.moreActions"
        @click="openMoreActions"
      >
        <Icon name="more" :size="20" />
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
  min-width: 0;
  flex-shrink: 1;
}

.batch-count {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.batch-count-main {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
}

.batch-count-detail {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: var(--space-xxs);
  flex-shrink: 0;
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

.btn-icon:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.btn-icon:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-icon.danger {
  color: var(--danger);
}

.btn-icon.danger:hover:not(:disabled) {
  background: var(--danger-soft);
}

.btn-icon:active:not(:disabled),
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
  white-space: nowrap;
}

.btn-text:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.divider {
  width: 1px;
  height: 20px;
  margin: 0 var(--space-xxs);
  background: var(--hairline);
  flex-shrink: 0;
}

.more-menu-btn {
  display: inline-flex;
}

@media (max-width: 767px) {
  .file-only-action {
    display: none;
  }

  .btn-text {
    font-size: 0.8125rem;
    padding: 0 var(--space-xxs);
  }
}

@media (min-width: 768px) {
  .more-menu-btn {
    display: none;
  }
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
