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
    <Transition name="toast">
      <div v-if="toastMessage" class="toast" :class="toastType" role="status" aria-live="polite">
        <Icon :name="iconName(toastType)" :size="18" class="toast-icon" />
        <span>{{ toastMessage }}</span>
      </div>
    </Transition>
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

.toast-enter-active {
  transition:
    opacity var(--motion-enter) var(--ease-enter),
    transform var(--motion-enter) var(--ease-enter);
}

.toast-leave-active {
  transition:
    opacity var(--motion-exit) var(--ease-exit),
    transform var(--motion-exit) var(--ease-exit);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

@media (prefers-reduced-motion: reduce) {
  .toast-enter-active,
  .toast-leave-active {
    transition: none;
  }

  .toast-enter-from,
  .toast-leave-to {
    transform: none;
  }
}
</style>
