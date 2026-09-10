<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { formatBytes } from '@/api/client'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'
import { useStorageUsage } from '@/lib/useStorageUsage'

const { t } = useI18n()
const storage = useStorageUsage(formatBytes)

onMounted(() => {
  void storage.reload()
})
</script>

<template>
  <section class="storage-card" :data-state="storage.state.value" :aria-label="t.storageLabel">
    <div class="storage-head">
      <div class="storage-title-wrap">
        <span class="storage-cloud-badge">
          <Icon name="cloud" :size="16" aria-hidden="true" />
        </span>
        <span class="storage-card-title">{{ t.myCloud }}</span>
      </div>
      <RouterLink to="/settings" class="storage-manage-link">
        {{ t.manageStorage }}
        <Icon name="chevron-right" :size="14" aria-hidden="true" />
      </RouterLink>
    </div>

    <p v-if="storage.state.value === 'loading'" class="storage-state muted" role="status">
      {{ t.storageLoading }}
    </p>

    <div v-else-if="storage.state.value === 'usage-unavailable'" class="storage-unavailable" role="status">
      <span class="storage-state-copy">
        <Icon name="info" :size="16" aria-hidden="true" />
        <span>{{ t.storageUnavailable }}</span>
      </span>
      <button type="button" class="storage-retry" data-testid="overview-storage-retry" @click="storage.reload()">
        {{ t.retry }}
      </button>
    </div>

    <template v-else>
      <div
        class="storage-progress-track"
        role="progressbar"
        :aria-valuenow="storage.progressAria.value.valuenow"
        aria-valuemin="0"
        aria-valuemax="100"
        :aria-label="storage.progressAria.value.label"
      >
        <div
          class="storage-progress-fill"
          :class="storage.toneClass.value"
          :style="{ '--fill-ratio': storage.fillRatio.value }"
        />
      </div>

      <div class="storage-meta" aria-hidden="true">
        <span class="storage-numbers">{{ formatBytes(storage.usedBytes.value) }} / {{ formatBytes(storage.quotaBytes.value) }}</span>
        <span class="storage-percentage">{{ storage.percent.value }}%</span>
      </div>

      <div v-if="storage.state.value === 'near-quota'" class="storage-state warning" role="status">
        <Icon name="alert" :size="14" aria-hidden="true" />
        <span>{{ t.storageNearQuota }}</span>
      </div>

      <div v-else-if="storage.state.value === 'full-quota'" class="storage-state danger" role="status">
        <Icon name="alert" :size="14" aria-hidden="true" />
        <span>{{ t.storageFullQuota }}</span>
        <RouterLink to="/trash" class="storage-free-up">{{ t.storageFreeUp }}</RouterLink>
      </div>
    </template>
  </section>
</template>

<style scoped>
.storage-card {
  background: var(--surface-soft, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--hairline);
  border-radius: var(--radius-lg, 16px);
  padding: 14px 16px;
  margin-bottom: var(--space-lg);
}

.storage-head,
.storage-meta,
.storage-unavailable,
.storage-state,
.storage-state-copy {
  display: flex;
  align-items: center;
}

.storage-head,
.storage-meta,
.storage-unavailable {
  justify-content: space-between;
}

.storage-head {
  gap: var(--space-sm);
  margin-bottom: 10px;
}

.storage-title-wrap {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
  color: var(--ink);
}

.storage-cloud-badge {
  width: 28px;
  height: 28px;
  flex: 0 0 28px;
  border-radius: 50%;
  background: var(--accent-soft, rgba(0, 132, 255, 0.12));
  color: var(--accent, #0084ff);
  display: flex;
  align-items: center;
  justify-content: center;
}

.storage-manage-link,
.storage-free-up {
  display: inline-flex;
  align-items: center;
  min-height: var(--touch-min);
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
}

.storage-manage-link {
  flex-shrink: 0;
  gap: 2px;
}

.storage-progress-track {
  width: 100%;
  height: 6px;
  background: var(--hairline, rgba(255, 255, 255, 0.1));
  border-radius: 9999px;
  overflow: hidden;
  margin-bottom: 8px;
}

.storage-progress-fill {
  width: 100%;
  height: 100%;
  background: var(--accent);
  border-radius: 9999px;
  transform: scaleX(var(--fill-ratio, 0));
  transform-origin: left center;
  transition: transform var(--duration-medium) var(--ease-standard);
}

.storage-progress-fill.warning {
  background: var(--warning);
}

.storage-progress-fill.danger {
  background: var(--danger);
}

.storage-meta {
  gap: var(--space-sm);
  font-size: 12px;
  font-weight: 500;
  color: var(--muted);
}

.storage-numbers {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.storage-percentage {
  flex-shrink: 0;
  font-weight: 600;
  color: var(--ink);
}

.storage-state,
.storage-state-copy {
  gap: var(--space-xs);
}

.storage-state {
  margin: var(--space-xs) 0 0;
  min-width: 0;
  font-size: 12px;
  line-height: 1.35;
}

.storage-state.warning {
  color: var(--warning);
}

.storage-state.danger {
  color: var(--danger);
  flex-wrap: wrap;
}

.storage-state.danger .storage-free-up {
  margin-left: auto;
  color: var(--danger);
}

.storage-unavailable {
  gap: var(--space-sm);
  min-width: 0;
}

.storage-state-copy {
  min-width: 0;
  color: var(--muted);
  font-size: 12px;
}

.storage-retry {
  min-height: var(--touch-min);
  flex-shrink: 0;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  padding: 0 var(--space-sm);
  background: var(--canvas);
  color: var(--ink);
  font: inherit;
  font-size: 12px;
  font-weight: 600;
}

.storage-retry:focus-visible,
.storage-manage-link:focus-visible,
.storage-free-up:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.muted {
  color: var(--muted);
}

@media (max-width: 359px) {
  .storage-head,
  .storage-unavailable {
    align-items: flex-start;
    flex-wrap: wrap;
  }
}

@media (prefers-reduced-motion: reduce) {
  .storage-progress-fill {
    transition: none;
  }
}
</style>
