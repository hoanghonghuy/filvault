<script setup lang="ts">
defineProps<{ mimeType: string; name: string }>()
</script>

<template>
  <div class="placeholder" :title="name">
    <span class="icon">{{ mimeType.startsWith('video/') ? '▶' : '🖼' }}</span>
    <span class="name">{{ name }}</span>
  </div>
</template>

<style scoped>
.placeholder {
  aspect-ratio: 1;
  border: 1px dashed var(--border);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 0.5rem;
  background: var(--surface-2);
  text-align: center;
}

.icon {
  font-size: 1.5rem;
}

.name {
  font-size: 0.75rem;
  color: var(--muted);
  word-break: break-all;
}
</style>
