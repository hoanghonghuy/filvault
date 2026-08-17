<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const { toastMessage } = storeToRefs(ui)
</script>

<template>
  <Teleport to="body">
    <div v-if="toastMessage" class="toast" role="status" aria-live="polite">
      {{ toastMessage }}
    </div>
  </Teleport>
</template>

<style scoped>
.toast {
  position: fixed;
  left: var(--space-md);
  right: var(--space-md);
  bottom: calc(var(--bottom-nav-h) + var(--space-md) + env(safe-area-inset-bottom));
  z-index: 1100;
  max-width: 420px;
  margin: 0 auto;
  padding: var(--space-sm) var(--space-md);
  background: var(--primary-cta);
  color: var(--on-ink);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  text-align: center;
  box-shadow: 0 8px 24px rgba(17, 24, 39, 0.2);
}

@media (min-width: 768px) {
  .toast {
    bottom: var(--space-lg);
  }
}
</style>
