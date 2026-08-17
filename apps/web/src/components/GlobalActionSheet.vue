<script setup lang="ts">
import { storeToRefs } from 'pinia'
import BottomSheet from '@/components/BottomSheet.vue'
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
        class="btn block action"
        :class="{ danger: item.danger }"
        @click="pick(item.id)"
      >
        {{ item.label }}
      </button>
      <button type="button" class="btn block ghost" @click="onClose">Cancel</button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.action {
  justify-content: flex-start;
}

.action.danger {
  color: var(--danger);
  border-color: var(--danger-soft);
}
</style>
