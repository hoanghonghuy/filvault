<script setup lang="ts">
const props = defineProps<{
  progress: number | null
}>()

const percent = () => Math.round((props.progress ?? 0) * 100)
</script>

<template>
  <div v-if="progress !== null" class="upload-progress" aria-live="polite">
    <div class="upload-progress-head">
      <span>Uploading…</span>
      <span>{{ percent() }}%</span>
    </div>
    <div
      class="track"
      role="progressbar"
      :aria-valuenow="percent()"
      aria-valuemin="0"
      aria-valuemax="100"
      aria-label="Upload progress"
    >
      <div class="fill" :style="{ transform: `scaleX(${progress})` }" />
    </div>
  </div>
</template>

<style scoped>
.upload-progress {
  margin-bottom: var(--space-sm);
}

.upload-progress-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  margin-bottom: var(--space-xxs);
  color: var(--muted);
  font-size: 13px;
  font-weight: 500;
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
  transition: transform var(--duration-medium) var(--ease-standard);
}

@media (prefers-reduced-motion: reduce) {
  .fill {
    transition: none;
  }
}
</style>
