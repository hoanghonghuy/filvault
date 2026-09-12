<template>
  <div class="sk-rail" aria-busy="true" aria-live="polite">
    <div v-for="i in 7" :key="i" class="sk-conv-row">
      <div class="skeleton sk-avatar" />
      <div class="sk-conv-meta">
        <div class="sk-conv-top">
          <div class="skeleton sk-name" :style="{ width: `${55 + (i * 7) % 30}%` }" />
          <div class="skeleton sk-time" />
        </div>
        <div class="sk-conv-bottom">
          <div class="skeleton sk-preview" :style="{ width: `${60 + (i * 11) % 35}%` }" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sk-rail {
  display: flex;
  flex-direction: column;
  padding: var(--space-xs) var(--space-sm);
  gap: var(--space-xs);
}

.sk-conv-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xs) var(--space-xs);
  min-height: 60px;
}

.sk-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sk-conv-meta {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.sk-conv-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-sm);
}

.sk-name {
  height: 14px;
  border-radius: 4px;
}

.sk-time {
  width: 36px;
  height: 10px;
  border-radius: 4px;
  flex-shrink: 0;
}

.sk-conv-bottom {
  display: flex;
  align-items: center;
}

.sk-preview {
  height: 12px;
  border-radius: 4px;
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
}

@keyframes skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

[data-theme='dark'] .skeleton {
  background: linear-gradient(
    90deg,
    rgba(31, 41, 55, 0.8) 25%,
    rgba(55, 65, 81, 0.9) 50%,
    rgba(31, 41, 55, 0.8) 75%
  );
  background-size: 200% 100%;
}

@media (prefers-reduced-motion: reduce) {
  .skeleton,
  [data-theme='dark'] .skeleton {
    animation: none;
    background: var(--hairline);
  }
}
</style>
