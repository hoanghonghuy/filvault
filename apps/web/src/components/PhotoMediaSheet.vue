<script setup lang="ts">
import BottomSheet from '@/components/BottomSheet.vue'

defineProps<{
  open: boolean
  name: string
}>()

const emit = defineEmits<{
  view: []
  download: []
  close: []
}>()
</script>

<template>
  <BottomSheet :open="open" :title="name" @close="emit('close')">
    <div class="actions">
      <button type="button" class="btn block ink" @click="emit('view')">View</button>
      <button type="button" class="btn block" @click="emit('download')">Download</button>
      <button type="button" class="btn block ghost" @click="emit('close')">Close</button>
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
