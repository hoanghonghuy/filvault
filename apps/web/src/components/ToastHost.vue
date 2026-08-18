<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useUiStore } from '@/stores/ui'
import Icon from '@/components/AppIcon.vue'

const ui = useUiStore()
const { toastMessage, toastType } = storeToRefs(ui)

const iconName = (type: string) => (type === 'error' ? 'alert' : type === 'success' ? 'check' : 'info')
</script>

<template>
  <Teleport to="body">
    <div v-if="toastMessage" class="toast" :class="toastType" role="status" aria-live="polite">
      <Icon :name="iconName(toastType)" :size="18" class="toast-icon" />
      <span>{{ toastMessage }}</span>
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
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--primary-cta);
  color: var(--on-ink);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 8px 24px rgba(17, 24, 39, 0.2);
}

.toast-icon {
  flex-shrink: 0;
}

.toast.error {
  background: var(--danger);
}

.toast.success {
  background: var(--success);
}

@media (min-width: 768px) {
  .toast {
    left: auto;
    right: var(--space-lg);
    bottom: var(--space-lg);
    margin: 0;
    max-width: 360px;
  }
}
</style>
