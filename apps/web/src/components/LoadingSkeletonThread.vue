<template>
  <div class="sk-thread" aria-busy="true" aria-live="polite">
    <div v-for="i in 6" :key="i" class="sk-row" :class="{ 'sk-alt': i % 3 === 0 }">
      <div class="skeleton sk-bubble" />
    </div>
  </div>
</template>

<style scoped>
.sk-thread {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: var(--space-sm);
  padding: var(--space-md);
}

.sk-row {
  display: flex;
  justify-content: flex-end;
}

.sk-row.sk-alt {
  justify-content: flex-start;
}

.sk-bubble {
  width: min(62%, 380px);
  height: 40px;
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
  border-radius: var(--radius-xl) var(--radius-xl) var(--radius-sm) var(--radius-xl);
}

@keyframes skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
