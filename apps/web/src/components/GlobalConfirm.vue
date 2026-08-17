<script setup lang="ts">
import { storeToRefs } from 'pinia'
import BottomSheet from '@/components/BottomSheet.vue'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const { confirmState } = storeToRefs(ui)

function onConfirm() {
  ui.resolveConfirm(true)
}

function onCancel() {
  ui.resolveConfirm(false)
}
</script>

<template>
  <BottomSheet :open="confirmState.open" :title="confirmState.title" @close="onCancel">
    <p v-if="confirmState.message" class="message">{{ confirmState.message }}</p>
    <div class="actions">
      <button
        type="button"
        class="btn block"
        :class="confirmState.danger ? 'danger' : 'ink'"
        @click="onConfirm"
      >
        {{ confirmState.confirmLabel }}
      </button>
      <button type="button" class="btn block ghost" @click="onCancel">
        {{ confirmState.cancelLabel }}
      </button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.message {
  margin: 0 0 var(--space-md);
  color: var(--body);
  font-size: 0.95rem;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}
</style>
