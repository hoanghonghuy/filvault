<script setup lang="ts">
type Mode = 'browse' | 'search'

defineProps<{
  mode: Mode
}>()
</script>

<template>
  <div class="sk-page" aria-busy="true" aria-live="polite">
    <div class="toolbar sk-toolbar">
      <div class="skeleton sk-input" />
      <div class="sk-btn-group">
        <div class="skeleton sk-btn" />
        <div class="skeleton sk-btn small" />
      </div>
    </div>

    <div class="list">
      <div v-for="i in (mode === 'search' ? 6 : 8)" :key="i" class="sk-row">
        <div class="skeleton sk-line sk-name" :style="{ width: `${mode === 'search' ? 70 : 62}%` }" />
        <div class="skeleton sk-line sk-meta desktop-only" />
        <div class="skeleton sk-icon" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.sk-page {
  width: 100%;
}

.sk-toolbar {
  display: flex;
  gap: var(--space-xs);
  align-items: center;
  margin-bottom: var(--space-md);
  padding-top: var(--space-xs);
}

.sk-input {
  flex: 1;
  height: var(--touch-min);
  border-radius: var(--radius-md);
}

.sk-btn-group {
  display: flex;
  gap: var(--space-xs);
  align-items: center;
}

.sk-btn {
  width: 110px;
  height: var(--touch-min);
  border-radius: var(--radius-md);
}

.sk-btn.small {
  width: 92px;
}

.list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.sk-row {
  min-height: var(--touch-min);
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: inline-block;
  }
}
</style>