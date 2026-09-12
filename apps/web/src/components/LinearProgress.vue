<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/lib/i18n'

const props = defineProps<{
  /** Progress from 0 to 1. */
  value: number
  label?: string
}>()

const { t } = useI18n()

const resolvedLabel = computed(() => props.label || t.value.uploadProgressLabel)

const percent = computed(() => {
  const raw = Number.isFinite(props.value) ? props.value : 0
  return Math.round(Math.min(1, Math.max(0, raw)) * 100)
})
</script>

<template>
  <div class="linear-progress" role="status" aria-live="polite">
    <div class="linear-progress__meta">
      <span class="linear-progress__label">{{ resolvedLabel }}</span>
      <span class="linear-progress__percent">{{ percent }}%</span>
    </div>
    <div
      class="linear-progress__track"
      role="progressbar"
      :aria-label="resolvedLabel"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-valuenow="percent"
    >
      <div
        class="linear-progress__fill"
        :style="{ width: `${percent}%` }"
      />
    </div>
  </div>
</template>

<style scoped>
.linear-progress {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  margin-bottom: var(--space-sm);
}

.linear-progress__meta {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-sm);
  color: var(--muted);
  font-size: 0.875rem;
}

.linear-progress__percent {
  font-variant-numeric: tabular-nums;
  color: var(--ink);
  font-weight: 600;
}

.linear-progress__track {
  height: 4px;
  overflow: hidden;
  border-radius: var(--radius-pill);
  background: var(--hairline);
}

.linear-progress__fill {
  height: 100%;
  border-radius: inherit;
  background: var(--accent);
  transition: width var(--motion-press) var(--ease-standard);
}

@media (prefers-reduced-motion: reduce) {
  .linear-progress__fill {
    transition: none;
  }
}
</style>
