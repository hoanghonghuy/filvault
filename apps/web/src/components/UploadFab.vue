<script setup lang="ts">
defineProps<{
  label: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  click: []
}>()
</script>

<template>
  <button
    type="button"
    class="fab"
    :aria-label="label"
    :disabled="disabled"
    @click="emit('click')"
  >
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
    </svg>
  </button>
</template>

<style scoped>
.fab {
  position: fixed;
  right: var(--space-md);
  bottom: calc(var(--bottom-nav-h) + var(--space-md) + env(safe-area-inset-bottom));
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--fab-size);
  height: var(--fab-size);
  border: none;
  border-radius: var(--radius-pill);
  background: var(--accent);
  color: var(--on-accent);
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(13, 148, 136, 0.35);
}

.fab:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 3px;
}

.fab:hover:not(:disabled) {
  background: var(--accent-hover);
}

.fab:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (min-width: 768px) {
  .fab {
    display: none;
  }
}
</style>
