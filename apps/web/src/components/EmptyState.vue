<script setup lang="ts">
import Icon from '@/components/AppIcon.vue'

defineProps<{
  title: string
  description?: string
  actionLabel?: string
  icon?: string
  compact?: boolean
}>()

const emit = defineEmits<{
  action: []
}>()
</script>

<template>
  <div class="empty" :class="{ compact }">
    <Icon v-if="icon" :name="icon" :size="32" class="empty-icon" />
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

.empty-icon {
  color: var(--muted-soft);
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

.empty.compact {
  align-items: flex-start;
  gap: var(--space-xxs);
  padding: var(--space-sm) 0;
  text-align: left;
  background: none;
  border: none;
  border-radius: 0;
}

.empty.compact .empty-icon {
  display: none;
}
</style>
