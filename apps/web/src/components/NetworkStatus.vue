<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'

const RECOVERY_VISIBLE_MS = 3000

const { locale } = useI18n()
const online = ref(true)
const recovered = ref(false)
let recoveryTimer: ReturnType<typeof setTimeout> | null = null

const copy = computed(() => {
  if (locale.value === 'en') {
    return online.value
      ? { title: 'Back online', detail: 'Your device connection has been restored.' }
      : { title: 'You’re offline', detail: 'Network actions will be available again when your device reconnects.' }
  }

  return online.value
    ? { title: 'Đã kết nối lại', detail: 'Kết nối mạng của thiết bị đã được khôi phục.' }
    : { title: 'Bạn đang ngoại tuyến', detail: 'Các thao tác cần mạng sẽ khả dụng lại khi thiết bị kết nối.' }
})

function clearRecoveryTimer() {
  if (recoveryTimer) {
    clearTimeout(recoveryTimer)
    recoveryTimer = null
  }
}

function handleOffline() {
  clearRecoveryTimer()
  online.value = false
  recovered.value = false
}

function handleOnline() {
  const wasOffline = !online.value
  online.value = true
  if (!wasOffline) return

  recovered.value = true
  clearRecoveryTimer()
  recoveryTimer = setTimeout(() => {
    recovered.value = false
    recoveryTimer = null
  }, RECOVERY_VISIBLE_MS)
}

onMounted(() => {
  online.value = typeof navigator === 'undefined' ? true : navigator.onLine
  window.addEventListener('offline', handleOffline)
  window.addEventListener('online', handleOnline)
})

onUnmounted(() => {
  clearRecoveryTimer()
  window.removeEventListener('offline', handleOffline)
  window.removeEventListener('online', handleOnline)
})
</script>

<template>
  <Transition name="network-status">
    <aside
      v-if="!online || recovered"
      class="network-status"
      :class="online ? 'is-recovered' : 'is-offline'"
      role="status"
      aria-live="polite"
      aria-atomic="true"
    >
      <Icon :name="online ? 'check' : 'info'" :size="18" aria-hidden="true" />
      <span class="network-copy">
        <strong>{{ copy.title }}</strong>
        <span>{{ copy.detail }}</span>
      </span>
    </aside>
  </Transition>
</template>

<style scoped>
.network-status {
  position: fixed;
  top: max(12px, env(safe-area-inset-top));
  left: 50%;
  z-index: 9000;
  display: flex;
  align-items: flex-start;
  gap: var(--space-xs);
  width: min(calc(100vw - 24px), 560px);
  padding: 10px 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--canvas);
  color: var(--ink);
  box-shadow: 0 12px 32px color-mix(in srgb, var(--ink) 16%, transparent);
  transform: translateX(-50%);
  pointer-events: none;
}

.network-status.is-offline {
  border-color: color-mix(in srgb, var(--warning) 48%, var(--hairline));
}

.network-status.is-recovered {
  border-color: color-mix(in srgb, var(--success) 48%, var(--hairline));
}

.network-copy {
  display: grid;
  gap: 2px;
  min-width: 0;
  font-size: 0.8125rem;
  line-height: 1.35;
}

.network-copy strong {
  font-size: 0.875rem;
}

.network-copy span {
  color: var(--muted);
}

.network-status-enter-active,
.network-status-leave-active {
  transition: opacity 160ms ease, transform 160ms ease;
}

.network-status-enter-from,
.network-status-leave-to {
  opacity: 0;
  transform: translate(-50%, -6px);
}

@media (prefers-reduced-motion: reduce) {
  .network-status-enter-active,
  .network-status-leave-active {
    transition: none;
  }
}
</style>
