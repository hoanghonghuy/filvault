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
  top: calc(var(--space-md) + env(safe-area-inset-top));
  right: var(--space-md);
  left: auto;
  z-index: 1100;
  max-width: min(calc(100vw - 32px), 380px);
  margin: 0;
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
    top: var(--space-lg);
    right: var(--space-lg);
    max-width: 380px;
  }
}

.toast-enter-active {
  transition:
    opacity var(--motion-enter, 0.25s) var(--ease-enter, ease-out),
    transform var(--motion-enter, 0.25s) var(--ease-enter, ease-out);
}

.toast-leave-active {
  transition:
    opacity var(--motion-exit, 0.2s) var(--ease-exit, ease-in),
    transform var(--motion-exit, 0.2s) var(--ease-exit, ease-in);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-16px);
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
