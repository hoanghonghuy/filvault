<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { UploadItemStatus } from '@/lib/uploadQueue'
import { useI18n } from '@/lib/i18n'

export interface UploadProgressItem {
  id: string
  name: string
  resolvedName?: string
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
  dismissFailed: [id: string]
  dismiss: []
}>()

const { t } = useI18n()

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
  const uploading = active.filter(
    (item) => item.status === 'uploading' || item.status === 'finalizing',
  ).length
  const queued = active.filter((item) => item.status === 'queued').length
  return { completed, failed, uploading, queued, total: active.length }
})

const hasActiveTransfers = computed(() =>
  props.items.some(
    (item) =>
      item.status === 'queued' ||
      item.status === 'uploading' ||
      item.status === 'finalizing',
  ),
)

const canClearSettled = computed(
  () => queueMode.value && props.items.length > 0 && !hasActiveTransfers.value,
)

const statusLabels: Record<UploadItemStatus, keyof typeof t.value> = {
  queued: 'uploadStatusQueued',
  uploading: 'uploadStatusUploading',
  finalizing: 'uploadStatusFinalizing',
  completed: 'uploadStatusCompleted',
  failed: 'uploadStatusFailed',
  cancelled: 'uploadStatusCancelled',
}

function statusLabel(status: UploadItemStatus): string {
  return t.value[statusLabels[status]]
}

function itemPercent(item: UploadProgressItem): number {
  if (item.status === 'completed') return 100
  if (item.status === 'uploading' || item.status === 'finalizing') {
    return item.status === 'finalizing' ? 100 : Math.round(item.progress * 100)
  }
  return 0
}

function displayName(item: UploadProgressItem): string {
  if (item.status === 'completed' && item.resolvedName) {
    return item.resolvedName
  }
  return item.name
}

function formatTemplate(template: string, values: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (_match, key: string) => String(values[key] ?? ''))
}

function conflictNote(item: UploadProgressItem): string | null {
  if (!item.resolvedName || item.resolvedName === item.name) return null
  return formatTemplate(t.value.uploadConflictRenamed, { name: item.resolvedName })
}

const liveAnnouncement = ref('')
const trackedStatuses = new Map<string, UploadItemStatus>()

watch(
  () => props.items.map((item) => ({ id: item.id, status: item.status, name: displayName(item) })),
  (snapshots) => {
    const messages: string[] = []
    const seen = new Set<string>()

    for (const snapshot of snapshots) {
      seen.add(snapshot.id)
      const previous = trackedStatuses.get(snapshot.id)
      if (previous === snapshot.status) continue

      trackedStatuses.set(snapshot.id, snapshot.status)
      if (previous === undefined) continue

      if (snapshot.status === 'completed') {
        messages.push(formatTemplate(t.value.uploadLiveCompleted, { name: snapshot.name }))
      } else if (snapshot.status === 'failed') {
        messages.push(formatTemplate(t.value.uploadLiveFailed, { name: snapshot.name }))
      } else if (snapshot.status === 'cancelled') {
        messages.push(formatTemplate(t.value.uploadLiveCancelled, { name: snapshot.name }))
      }
    }

    for (const id of trackedStatuses.keys()) {
      if (!seen.has(id)) trackedStatuses.delete(id)
    }

    if (messages.length > 0) {
      liveAnnouncement.value = messages.join('. ')
    }
  },
)
</script>

<template>
  <section v-if="visible" class="upload-progress">
    <p class="upload-live-status sr-only" aria-live="polite" aria-atomic="true">
      {{ liveAnnouncement }}
    </p>

    <div class="upload-progress-head">
      <div class="upload-progress-title">
        <span>{{ label ?? t.uploadProgressLabel }}</span>
        <span v-if="queueMode" class="upload-progress-summary">
          {{ summary.completed }}/{{ summary.total }}
          <template v-if="summary.failed > 0">
            · {{ summary.failed }} {{ t.uploadSummaryFailed }}
          </template>
        </span>
      </div>
      <div class="upload-progress-actions">
        <span class="upload-progress-percent" aria-hidden="true">{{ percent }}%</span>
        <button
          v-if="canClearSettled"
          type="button"
          class="upload-dismiss-btn"
          :aria-label="t.uploadClearSettledAria"
          @click="emit('dismiss')"
        >
          {{ t.uploadClearSettled }}
        </button>
      </div>
    </div>
    <div
      class="track"
      role="progressbar"
      :aria-valuenow="percent"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-label="label ?? t.uploadProgressAggregateAria"
    >
      <div class="fill" :style="{ transform: `scaleX(${effectiveProgress ?? 0})` }" />
    </div>

    <ul v-if="queueMode" class="upload-item-list" :aria-label="t.uploadProgressListAria">
      <li
        v-for="item in items"
        :key="item.id"
        class="upload-item"
        :data-status="item.status"
      >
        <div class="upload-item-main">
          <div class="upload-item-names">
            <span class="upload-item-name" :title="displayName(item)">{{ displayName(item) }}</span>
            <span
              v-if="conflictNote(item)"
              class="upload-item-renamed"
              :title="conflictNote(item) ?? undefined"
            >
              {{ conflictNote(item) }}
            </span>
          </div>
          <span class="upload-item-status">
            <span class="sr-only">{{ statusLabel(item.status) }}</span>
            <span aria-hidden="true">{{ statusLabel(item.status) }}</span>
            <template v-if="item.status === 'uploading' || item.status === 'finalizing'">
              · {{ itemPercent(item) }}%
            </template>
          </span>
        </div>
        <div
          v-if="item.status === 'uploading' || item.status === 'finalizing'"
          class="upload-item-track"
          role="progressbar"
          :aria-label="formatTemplate(t.uploadItemProgressAria, { name: displayName(item) })"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-valuenow="itemPercent(item)"
        >
          <div
            class="upload-item-fill"
            :style="{ transform: `scaleX(${item.status === 'finalizing' ? 1 : item.progress})` }"
          />
        </div>
        <p v-if="item.status === 'failed' && item.error" class="upload-item-error" role="alert">
          {{ item.error }}
        </p>
        <div
          v-if="item.status === 'failed' || item.status === 'queued' || item.status === 'uploading'"
          class="upload-item-actions"
        >
          <button
            v-if="item.status === 'failed'"
            type="button"
            class="upload-action-btn"
            @click="emit('retry', item.id)"
          >
            {{ t.retry }}
          </button>
          <button
            v-if="item.status === 'failed'"
            type="button"
            class="upload-action-btn subtle"
            :aria-label="t.uploadDismissFailedAria"
            @click="emit('dismissFailed', item.id)"
          >
            {{ t.uploadDismissFailed }}
          </button>
          <button
            v-if="item.status === 'queued' || item.status === 'uploading'"
            type="button"
            class="upload-action-btn subtle"
            @click="emit('cancel', item.id)"
          >
            {{ t.cancel }}
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

.upload-live-status {
  margin: 0;
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

.upload-item-names {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
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

.upload-item-renamed {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.35;
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
  overflow-wrap: anywhere;
}

.upload-item-actions {
  display: flex;
  flex-wrap: wrap;
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
