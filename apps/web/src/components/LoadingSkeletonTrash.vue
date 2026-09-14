<script setup lang="ts"></script>

<template>
  <div class="sk-page" aria-busy="true" aria-live="polite">
    <h1 class="page-title desktop-only">Trash</h1>

    <section class="list-section">
      <div class="section-title sk-line"> </div>
      <div class="sk-rows">
        <div v-for="i in 3" :key="i" class="sk-row">
          <div class="skeleton sk-name" />
          <div class="skeleton sk-icon" />
        </div>
      </div>
    </section>

    <section class="list-section">
      <div class="section-title sk-line"> </div>
      <div class="sk-rows">
        <div v-for="i in 5" :key="i" class="sk-row">
          <div class="skeleton sk-name" />
          <div class="skeleton sk-meta desktop-only" />
          <div class="skeleton sk-icon" />
        </div>
      </div>
    </section>
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
  min-height: 200px;
}

.list-section {
  margin-top: var(--space-md);
}

.sk-line {
  width: 180px;
  height: 14px;
  border-radius: 8px;
  background: linear-gradient(
    90deg,
    rgba(229, 231, 235, 0.85) 25%,
    rgba(243, 244, 246, 1) 50%,
    rgba(229, 231, 235, 0.85) 75%
  );
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.2s ease-in-out infinite;
}

.sk-rows {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  margin-top: var(--space-sm);
}

.sk-row {
  min-height: var(--touch-min);
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: 0.55rem 0.65rem;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface);
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

.sk-name {
  flex: 1;
  min-width: 0;
  height: 14px;
}

.sk-meta {
  width: 120px;
  height: 14px;
  border-radius: 999px;
}

.sk-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
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
  .skeleton,
  .sk-line {
    animation: none;
    background: var(--hairline);
  }
}
</style>

