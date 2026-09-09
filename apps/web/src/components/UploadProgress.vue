<script setup lang="ts">
import { computed } from 'vue'
import type { UploadItemStatus } from '@/lib/uploadQueue'

export interface UploadProgressItem {
  id: string
  name: string
  status: UploadItemStatus
  progress: number
  error?: string
}

const props = withDefaults(
  defineProps<{
    aggregateProgress?: number | null
    items?: UploadProgressItem[]
    /** Legacy single-value progress used by ChatView. */
    progress?: number | null
    label?: string
  }>(),
  {
    aggregateProgress: null,
    items: () => [],
  },
)

const emit = defineEmits<{
  retry: [id: string]
  cancel: [id: string]
  dismiss: []
}>()

const effectiveProgress = computed(() => props.aggregateProgress ?? props.progress ?? null)

const percent = computed(() => Math.round(Math.min(1, Math.max(0, effectiveProgress.value ?? 0)) * 100))

const queueMode = computed(() => props.items.length > 0 && props.aggregateProgress !== null)

const visible = computed(() => {
  if (queueMode.value) return true
  return props.progress !== null && props.progress !== undefined
})

const summary = computed(() => {
  const active = props.items.filter((item) => item.status !== 'cancelled')
  const completed = active.filter((item) => item.status === 'completed').length
  const failed = active.filter((item) => item.status === 'failed').length
  const uploading = active.filter((item) => item.status === 'uploading').length
  const queued = active.filter((item) => item.status === 'queued').length
  return { completed, failed, uploading, queued, total: active.length }
})

const canDismiss = computed(() =>
  props.items.length > 0 &&
  props.items.every((item) => item.status === 'completed' || item.status === 'cancelled'),
)

function statusLabel(status: UploadItemStatus): string {
  switch (status) {
    case 'queued':
      return 'Đang chờ'
    case 'uploading':
      return 'Đang tải lên'
    case 'completed':
      return 'Hoàn tất'
    case 'failed':
      return 'Thất bại'
    case 'cancelled':
      return 'Đã hủy'
    default:
      return status
  }
}

function itemPercent(item: UploadProgressItem): number {
  if (item.status === 'completed') return 100
  if (item.status === 'uploading') return Math.round(item.progress * 100)
  return 0
}
</script>

<template>
  <section v-if="visible" class="upload-progress" aria-live="polite">
    <div class="upload-progress-head">
      <div class="upload-progress-title">
        <span>{{ label || 'Đang tải lên…' }}</span>
        <span v-if="queueMode" class="upload-progress-summary">
          {{ summary.completed }}/{{ summary.total }}
          <template v-if="summary.failed > 0"> · {{ summary.failed }} thất bại</template>
        </span>
      </div>
      <div class="upload-progress-actions">
        <span class="upload-progress-percent">{{ percent }}%</span>
        <button
          v-if="queueMode && canDismiss"
          type="button"
          class="upload-dismiss-btn"
          aria-label="Đóng tiến trình tải lên"
          @click="emit('dismiss')"
        >
          Đóng
        </button>
      </div>
    </div>
    <div
      class="track"
      role="progressbar"
      :aria-valuenow="percent"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-label="label || 'Tiến trình tải lên tổng'"
    >
      <div class="fill" :style="{ transform: `scaleX(${effectiveProgress ?? 0})` }" />
    </div>

    <ul v-if="queueMode" class="upload-item-list" aria-label="Danh sách tệp đang tải lên">
      <li
        v-for="item in items"
        :key="item.id"
        class="upload-item"
        :data-status="item.status"
      >
        <div class="upload-item-main">
          <span class="upload-item-name" :title="item.name">{{ item.name }}</span>
          <span class="upload-item-status">
            <span class="sr-only">{{ statusLabel(item.status) }}</span>
            <span aria-hidden="true">{{ statusLabel(item.status) }}</span>
            <template v-if="item.status === 'uploading'"> · {{ itemPercent(item) }}%</template>
          </span>
        </div>
        <div
          v-if="item.status === 'uploading'"
          class="upload-item-track"
          role="progressbar"
          :aria-label="`Tiến trình ${item.name}`"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-valuenow="itemPercent(item)"
        >
          <div class="upload-item-fill" :style="{ transform: `scaleX(${item.progress})` }" />
        </div>
        <p v-if="item.status === 'failed' && item.error" class="upload-item-error" role="alert">
          {{ item.error }}
        </p>
        <div v-if="item.status === 'failed' || item.status === 'queued' || item.status === 'uploading'" class="upload-item-actions">
          <button
            v-if="item.status === 'failed'"
            type="button"
            class="upload-action-btn"
            @click="emit('retry', item.id)"
          >
            Thử lại
          </button>
          <button
            v-if="item.status === 'queued' || item.status === 'uploading'"
            type="button"
            class="upload-action-btn subtle"
            @click="emit('cancel', item.id)"
          >
            Hủy
          </button>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.upload-progress {
  margin-bottom: var(--space-sm);
  padding: var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
}

.upload-progress-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-sm);
  margin-bottom: var(--space-xxs);
  color: var(--muted);
  font-size: 13px;
  font-weight: 500;
}

.upload-progress-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.upload-progress-summary {
  font-size: 12px;
  color: var(--muted);
  font-weight: 400;
}

.upload-progress-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  flex-shrink: 0;
}

.upload-progress-percent {
  font-variant-numeric: tabular-nums;
  color: var(--ink);
  font-weight: 600;
}

.upload-dismiss-btn {
  min-height: var(--touch-min, 44px);
  padding: 0 var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  color: var(--ink);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.track {
  height: 6px;
  overflow: hidden;
  background: var(--hairline);
  border-radius: var(--radius-pill);
}

.fill {
  height: 100%;
  background: var(--accent);
  border-radius: var(--radius-pill);
  transform-origin: left center;
  transition: transform var(--duration-medium) var(--ease-standard);
}

.upload-item-list {
  list-style: none;
  margin: var(--space-sm) 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  max-height: min(40vh, 280px);
  overflow: auto;
}

.upload-item {
  padding: var(--space-xs);
  border-radius: var(--radius-sm);
  background: var(--canvas);
  border: 1px solid var(--hairline-soft, var(--hairline));
}

.upload-item-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-sm);
}

.upload-item-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ink);
  font-size: 13px;
  font-weight: 500;
}

.upload-item-status {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}

.upload-item[data-status='failed'] .upload-item-status {
  color: var(--danger);
}

.upload-item[data-status='completed'] .upload-item-status {
  color: var(--success, #059669);
}

.upload-item-track {
  height: 4px;
  margin-top: var(--space-xxs);
  overflow: hidden;
  background: var(--hairline);
  border-radius: var(--radius-pill);
}

.upload-item-fill {
  height: 100%;
  background: var(--accent);
  border-radius: var(--radius-pill);
  transform-origin: left center;
  transition: transform var(--duration-medium) var(--ease-standard);
}

.upload-item-error {
  margin: var(--space-xxs) 0 0;
  color: var(--danger);
  font-size: 12px;
  line-height: 1.4;
}

.upload-item-actions {
  display: flex;
  gap: var(--space-xs);
  margin-top: var(--space-xxs);
}

.upload-action-btn {
  min-height: var(--touch-min, 44px);
  padding: 0 var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.upload-action-btn.subtle {
  color: var(--muted);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (prefers-reduced-motion: reduce) {
  .fill,
  .upload-item-fill {
    transition: none;
  }
}

@media (min-width: 768px) {
  .upload-item-list {
    max-height: 320px;
  }
}
</style>
