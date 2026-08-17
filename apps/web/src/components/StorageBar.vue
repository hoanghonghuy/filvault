<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import type { StorageUsage } from '@/api/types'

const usage = ref<StorageUsage | null>(null)

onMounted(async () => {
  try {
    usage.value = await api<StorageUsage>('/storage')
  } catch {
    usage.value = null
  }
})

defineExpose({ reload: async () => { usage.value = await api<StorageUsage>('/storage') } })
</script>

<template>
  <div v-if="usage" class="storage-bar">
    <div class="label">Storage</div>
    <div class="track">
      <div
        class="fill"
        :style="{ width: `${Math.min(100, (usage.usedBytes / usage.quotaBytes) * 100)}%` }"
      />
    </div>
    <div class="numbers">{{ formatBytes(usage.usedBytes) }} / {{ formatBytes(usage.quotaBytes) }}</div>
  </div>
</template>

<style scoped>
.storage-bar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 1.25rem;
  background: var(--surface-2);
  border-bottom: 1px solid var(--border);
  font-size: 0.85rem;
}

.label {
  color: var(--muted);
}

.track {
  flex: 1;
  height: 8px;
  background: var(--border);
  border-radius: 999px;
  overflow: hidden;
}

.fill {
  height: 100%;
  background: var(--accent);
  border-radius: 999px;
}

.numbers {
  white-space: nowrap;
}
</style>
