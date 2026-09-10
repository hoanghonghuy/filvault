<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import BottomSheet from '@/components/BottomSheet.vue'
import Icon from '@/components/AppIcon.vue'
import { useUiStore, type ActionSheetItem } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'

const ui = useUiStore()
const { actionSheetState } = storeToRefs(ui)
const { t } = useI18n()

const actionLabelsById = computed<Record<string, string>>(() => ({
  open: t.value.open,
  rename: t.value.rename,
  move: t.value.move,
  vault: t.value.vaultMoveToVault,
  trash: t.value.moveToTrash,
  preview: t.value.preview,
  view: t.value.preview,
  download: t.value.download,
  restore: t.value.restore,
  add: t.value.addPhotos,
  'set-cover': t.value.setAsCover,
  'remove-cover': t.value.removeCover,
  remove: t.value.removeFromAlbum,
}))

// Transitional fallback for callers whose action id is context-dependent (for example
// `delete` and `favorite`). New/unambiguous actions should localize by stable id so
// changing English display copy cannot silently break locale switching.
const legacyActionLabels = computed<Record<string, string>>(() => ({
  'Move to vault': t.value.vaultMoveToVault,
  'Move to trash': t.value.moveToTrash,
  'Add to favorites': t.value.addToFavorites,
  'Remove from favorites': t.value.removeFromFavorites,
  'Share link': t.value.shareLink,
  'Share with user': t.value.shareWithUser,
  'Delete forever': t.value.deleteForever,
  'Rename album': t.value.renameAlbum,
  'Delete album': t.value.deleteAlbum,
  'Remove from this album': t.value.removeFromAlbum,
}))

function displayLabel(item: ActionSheetItem) {
  return actionLabelsById.value[item.id] ?? legacyActionLabels.value[item.label] ?? item.label
}

function pick(id: string) {
  ui.resolveActionSheet(id)
}

function onClose() {
  ui.resolveActionSheet(null)
}
</script>

<template>
  <BottomSheet :open="actionSheetState.open" :title="actionSheetState.title" @close="onClose" @after-leave="ui.notifyActionSheetAfterLeave()">
    <div class="actions">
      <button
        v-for="item in actionSheetState.items"
        :key="item.id"
        type="button"
        class="sheet-row"
        :class="{ danger: item.danger }"
        @click="pick(item.id)"
      >
        <Icon :name="item.icon ?? 'file'" :size="20" class="sheet-row-icon" />
        <span class="sheet-row-label">{{ displayLabel(item) }}</span>
      </button>
    </div>
    <div class="sheet-cancel-group">
      <button type="button" class="btn block ink" @click="onClose">{{ t.cancel }}</button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.actions {
  display: flex;
  flex-direction: column;
}

.sheet-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  min-height: 48px;
  padding: 0 var(--space-sm);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--ink);
  font-size: 15px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
}

.sheet-row:hover {
  background: var(--surface-soft);
}

.sheet-row:active {
  transform: scale(0.98);
}

.sheet-row-icon {
  flex-shrink: 0;
  color: var(--muted);
}

.sheet-row-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sheet-row.danger,
.sheet-row.danger .sheet-row-icon {
  color: var(--danger);
}

.sheet-cancel-group {
  margin-top: var(--space-sm);
  padding-top: var(--space-sm);
  border-top: 1px solid var(--hairline-soft);
}
</style>
