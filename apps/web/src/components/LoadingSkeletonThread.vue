<template>
  <div class="sk-thread" aria-busy="true" aria-live="polite">
    <div class="sk-date-wrap">
      <div class="skeleton sk-date" />
    </div>

    <!-- Message 1: Incoming with avatar -->
    <div class="sk-row sk-incoming">
      <div class="skeleton sk-avatar" />
      <div class="skeleton sk-bubble sk-b1" />
    </div>

    <!-- Message 2: Incoming cluster follow-up -->
    <div class="sk-row sk-incoming sk-followup">
      <div class="sk-avatar-spacer" />
      <div class="skeleton sk-bubble sk-b2" />
    </div>

    <!-- Message 3: Outgoing -->
    <div class="sk-row sk-outgoing">
      <div class="skeleton sk-bubble sk-b3" />
    </div>

    <!-- Message 4: Outgoing follow-up -->
    <div class="sk-row sk-outgoing">
      <div class="skeleton sk-bubble sk-b4" />
    </div>

    <!-- Message 5: Incoming with avatar -->
    <div class="sk-row sk-incoming">
      <div class="skeleton sk-avatar" />
      <div class="skeleton sk-bubble sk-b5" />
    </div>

    <!-- Message 6: Outgoing -->
    <div class="sk-row sk-outgoing">
      <div class="skeleton sk-bubble sk-b6" />
    </div>
  </div>
</template>

<style scoped>
.sk-thread {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-md);
}

.sk-date-wrap {
  display: flex;
  justify-content: center;
  margin-bottom: var(--space-sm);
}

.sk-date {
  width: 90px;
  height: 18px;
  border-radius: 999px;
}

.sk-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  margin-bottom: 2px;
}

.sk-row.sk-outgoing {
  justify-content: flex-end;
}

.sk-row.sk-incoming {
  justify-content: flex-start;
}

.sk-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sk-avatar-spacer {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.sk-bubble {
  min-height: 38px;
}

.sk-outgoing .sk-bubble {
  border-radius: 18px 18px 4px 18px;
}

.sk-incoming .sk-bubble {
  border-radius: 18px 18px 18px 4px;
}

.sk-b1 {
  width: min(58%, 320px);
}

.sk-b2 {
  width: min(76%, 420px);
  min-height: 52px;
}

.sk-b3 {
  width: min(42%, 240px);
}

.sk-b4 {
  width: min(65%, 360px);
}

.sk-b5 {
  width: min(48%, 280px);
}

.sk-b6 {
  width: min(35%, 190px);
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
</style>
