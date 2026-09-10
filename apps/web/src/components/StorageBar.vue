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

defineExpose({ reload: storage.reload })
</script>

<template>
  <div class="storage-bar" :data-state="storage.state.value">
    <template v-if="storage.state.value === 'loading'">
      <span class="label">{{ t.storageLabel }}</span>
      <p class="status-message muted" role="status">{{ t.storageLoading }}</p>
    </template>

    <template v-else-if="storage.state.value === 'usage-unavailable'">
      <Icon name="info" :size="16" class="status-icon muted" aria-hidden="true" />
      <p class="status-message muted" role="status">{{ t.storageUnavailable }}</p>
      <button
        type="button"
        class="retry-btn"
        data-testid="storage-retry"
        @click="storage.reload()"
      >
        {{ t.retry }}
      </button>
    </template>

    <template v-else>
      <span class="label desktop-only">{{ t.storageLabel }}</span>
      <div
        class="track"
        role="progressbar"
        :aria-valuenow="storage.progressAria.value.valuenow"
        aria-valuemin="0"
        aria-valuemax="100"
        :aria-label="storage.progressAria.value.label"
      >
        <div
          class="fill"
          :class="storage.toneClass.value"
          :style="{ '--fill-ratio': storage.fillRatio.value }"
        />
      </div>
      <div class="numbers" aria-hidden="true">
        {{ formatBytes(storage.usedBytes.value) }} / {{ formatBytes(storage.quotaBytes.value) }}
      </div>

      <div
        v-if="storage.state.value === 'near-quota'"
        class="status near-quota"
        role="status"
      >
        <Icon name="alert" :size="14" class="status-icon" aria-hidden="true" />
        <span class="status-text">{{ t.storageNearQuota }}</span>
      </div>

      <div
        v-else-if="storage.state.value === 'full-quota'"
        class="status full-quota"
        role="status"
      >
        <Icon name="alert" :size="14" class="status-icon" aria-hidden="true" />
        <span class="status-copy">
          <span class="status-text">{{ t.storageFullQuota }}</span>
          <span class="status-hint">{{ t.storageFullQuotaHint }}</span>
        </span>
        <RouterLink to="/trash" class="action-link" data-testid="storage-action">
          {{ t.storageFreeUp }}
        </RouterLink>
      </div>
    </template>
  </div>
</template>

<style scoped>
.storage-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-xs) var(--space-sm);
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

.status-message {
  margin: 0;
  min-width: 0;
  flex: 1 1 8rem;
  font-size: 0.75rem;
  line-height: 1.3;
}

.muted {
  color: var(--muted);
}

.retry-btn {
  flex-shrink: 0;
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--canvas);
  color: var(--ink);
  font-size: 0.75rem;
  font-weight: 600;
}

.retry-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.track {
  flex: 1 1 5rem;
  min-width: 4rem;
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
  flex-shrink: 0;
  white-space: nowrap;
  font-size: 0.75rem;
  font-variant-numeric: tabular-nums;
}

.status {
  display: flex;
  align-items: flex-start;
  gap: var(--space-xs);
  flex: 1 1 100%;
  min-width: 0;
  font-size: 0.75rem;
  line-height: 1.3;
}

.status.near-quota {
  color: var(--warning);
}

.status.full-quota {
  color: var(--danger);
}

.status-icon {
  flex-shrink: 0;
  margin-top: 1px;
}

.status-copy {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
  flex: 1;
}

.status-text {
  font-weight: 600;
}

.status-hint {
  color: var(--muted);
  font-weight: 400;
}

.action-link {
  flex-shrink: 0;
  align-self: center;
  padding: 2px var(--space-sm);
  border-radius: var(--radius-pill);
  background: var(--danger-soft, rgba(239, 68, 68, 0.12));
  color: var(--danger);
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
}

.action-link:focus-visible {
  outline: 2px solid var(--danger);
  outline-offset: 2px;
}

@media (min-width: 768px) {
  .desktop-only {
    display: inline;
  }

  .numbers {
    font-size: 0.8125rem;
  }

  .status {
    flex: 0 1 auto;
    flex-basis: auto;
  }

  .status.full-quota {
    flex: 1 1 auto;
  }
}
</style>
