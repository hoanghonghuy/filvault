<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  actionLabel?: string
}>()

const emit = defineEmits<{
  action: []
}>()
</script>

<template>
  <div class="empty">
    <p class="empty-title">{{ title }}</p>
    <p v-if="description" class="empty-desc">{{ description }}</p>
    <button v-if="actionLabel" type="button" class="btn ink" @click="emit('action')">
      {{ actionLabel }}
    </button>
  </div>
</template>

<style scoped>
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xl) var(--space-md);
  text-align: center;
  background: var(--surface-card);
  border: 1px dashed var(--hairline);
  border-radius: var(--radius-lg);
}

.empty-title {
  margin: 0;
  font-weight: 600;
  color: var(--ink);
}

.empty-desc {
  margin: 0;
  font-size: 0.9rem;
  color: var(--muted);
  max-width: 28rem;
}
</style>
