<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import type { StorageUsage } from '@/api/types'

const usage = ref<StorageUsage | null>(null)

const fillClass = computed(() => {
  if (!usage.value) return ''
  const pct = (usage.value.usedBytes / usage.value.quotaBytes) * 100
  if (pct >= 100) return 'danger'
  if (pct >= 90) return 'warning'
  return ''
})

async function reload() {
  try {
    usage.value = await api<StorageUsage>('/storage')
  } catch {
    usage.value = null
  }
}

onMounted(reload)

defineExpose({ reload })
</script>

<template>
  <div v-if="usage" class="storage-bar">
    <span class="label desktop-only">Storage</span>
    <div
      class="track"
      role="progressbar"
      :aria-valuenow="Math.round((usage.usedBytes / usage.quotaBytes) * 100)"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-label="`${formatBytes(usage.usedBytes)} of ${formatBytes(usage.quotaBytes)} used`"
    >
      <div
        class="fill"
        :class="fillClass"
        :style="{ '--fill-ratio': Math.min(1, usage.usedBytes / usage.quotaBytes) }"
      />
    </div>
    <div class="numbers" aria-hidden="true">{{ formatBytes(usage.usedBytes) }} / {{ formatBytes(usage.quotaBytes) }}</div>
  </div>
</template>

<style scoped>
.storage-bar {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xs) var(--space-md);
  background: var(--surface-soft);
  border-bottom: 1px solid var(--hairline);
  font-size: 0.8125rem;
}

.label {
  color: var(--muted);
  flex-shrink: 0;
}

.desktop-only {
  display: none;
}

.track {
  flex: 1;
  height: 8px;
  background: var(--hairline);
  border-radius: var(--radius-pill);
  overflow: hidden;
}

.fill {
  height: 100%;
  background: var(--accent);
  border-radius: var(--radius-pill);
  transform: scaleX(var(--fill-ratio, 0));
  transform-origin: left center;
  transition: transform var(--duration-medium) var(--ease-standard);
}

@media (prefers-reduced-motion: reduce) {
  .fill {
    transition: none;
  }
}

.fill.warning {
  background: var(--warning);
}

.fill.danger {
  background: var(--danger);
}

.numbers {
  color: var(--muted);
  white-space: nowrap;
  font-size: 0.75rem;
}

@media (min-width: 768px) {
  .desktop-only {
    display: inline;
  }

  .numbers {
    font-size: 0.8125rem;
  }
}
</style>
