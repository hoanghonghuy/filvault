<script setup lang="ts">
import { storeToRefs } from 'pinia'
import BottomSheet from '@/components/BottomSheet.vue'
import Icon from '@/components/AppIcon.vue'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const { actionSheetState } = storeToRefs(ui)

function pick(id: string) {
  ui.resolveActionSheet(id)
}

function onClose() {
  ui.resolveActionSheet(null)
}
</script>

<template>
  <BottomSheet :open="actionSheetState.open" :title="actionSheetState.title" @close="onClose">
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
        <span class="sheet-row-label">{{ item.label }}</span>
      </button>
    </div>
    <div class="sheet-cancel-group">
      <button type="button" class="btn block ink" @click="onClose">Cancel</button>
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
