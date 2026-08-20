<script setup lang="ts">
defineProps<{ variant?: 'initial' }>()
</script>

<template>
  <div class="sk-page" aria-busy="true" aria-live="polite">
    <div class="album-header">
      <div class="skeleton sk-back" />
      <div class="skeleton sk-title desktop-only" />
      <div class="skeleton sk-menu" />
    </div>

    <div class="grid photos">
      <div v-for="i in 10" :key="i" class="sk-photo skeleton" />
    </div>
  </div>
</template>

<style scoped>
.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: inline-block;
  }
}

.sk-page {
  padding-top: var(--space-md);
}

.album-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
}

.skeleton {
  background: linear-gradient(
    90deg,
    rgba(229, 231, 235, 0.85) 25%,
    rgba(243, 244, 246, 1) 50%,
    rgba(229, 231, 235, 0.85) 75%
  );
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.2s ease-in-out infinite;
  border-radius: 8px;
}

.sk-back {
  width: 96px;
  height: 36px;
}

.sk-title {
  flex: 1;
  height: 14px;
}

.sk-menu {
  width: 40px;
  height: 40px;
  border-radius: 12px;
}

.sk-photo {
  aspect-ratio: 1;
  border-radius: var(--radius-md);
}

.grid.photos {
  display: grid;
  gap: var(--space-xs);
  grid-template-columns: repeat(auto-fill, minmax(108px, 1fr));
}

@media (min-width: 768px) {
  .grid.photos {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  }
}

@keyframes skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton {
    animation: none;
    background: var(--hairline);
  }
}
</style>

