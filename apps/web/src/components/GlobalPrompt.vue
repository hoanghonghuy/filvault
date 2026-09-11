<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import BottomSheet from '@/components/BottomSheet.vue'
import { useI18n } from '@/lib/i18n'
import { sharedDialogCopy } from '@/lib/sharedDialogCopy'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const { promptState } = storeToRefs(ui)
const { locale } = useI18n()
const copy = computed(() => sharedDialogCopy(locale.value))
const confirmLabel = computed(() => promptState.value.confirmLabel ?? copy.value.save)
const cancelLabel = computed(() => promptState.value.cancelLabel ?? copy.value.cancel)

function onSubmit() {
  const value = promptState.value.value.trim()
  ui.resolvePrompt(value || null)
}

function onCancel() {
  ui.resolvePrompt(null)
}
</script>

<template>
  <BottomSheet :open="promptState.open" :title="promptState.title" @close="onCancel">
    <label v-if="promptState.label" class="field">
      <span>{{ promptState.label }}</span>
      <input v-model="promptState.value" type="text" autofocus @keyup.enter="onSubmit" />
    </label>
    <label v-else class="field">
      <span class="sr-only">{{ promptState.title }}</span>
      <input v-model="promptState.value" type="text" autofocus @keyup.enter="onSubmit" />
    </label>
    <div class="actions">
      <button type="button" class="btn block ink" @click="onSubmit">
        {{ confirmLabel }}
      </button>
      <button type="button" class="btn block ghost" @click="onCancel">
        {{ cancelLabel }}
      </button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}
</style>
