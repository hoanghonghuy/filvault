<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import BottomSheet from '@/components/BottomSheet.vue'
import { useI18n } from '@/lib/i18n'
import { sharedDialogCopy } from '@/lib/sharedDialogCopy'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const { confirmState } = storeToRefs(ui)
const { locale } = useI18n()
const copy = computed(() => sharedDialogCopy(locale.value))
const confirmLabel = computed(() => confirmState.value.confirmLabel ?? copy.value.confirm)
const cancelLabel = computed(() => confirmState.value.cancelLabel ?? copy.value.cancel)

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
        {{ confirmLabel }}
      </button>
      <button type="button" class="btn block ghost" @click="onCancel">
        {{ cancelLabel }}
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
