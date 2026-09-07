<template>
  <div class="sk-thread" aria-busy="true" aria-live="polite">
    <div class="sk-row sk-in">
      <div class="skeleton sk-av" />
      <div class="skeleton sk-msg" style="width: 55%" />
    </div>
    <div class="sk-row sk-in">
      <div class="sk-av-gap" />
      <div class="skeleton sk-msg" style="width: 40%" />
    </div>
    <div class="sk-row sk-out">
      <div class="skeleton sk-msg" style="width: 48%" />
    </div>
    <div class="sk-row sk-out">
      <div class="skeleton sk-msg" style="width: 30%" />
    </div>
    <div class="sk-row sk-in">
      <div class="skeleton sk-av" />
      <div class="skeleton sk-msg" style="width: 60%" />
    </div>
    <div class="sk-row sk-out">
      <div class="skeleton sk-msg" style="width: 38%" />
    </div>
    <div class="sk-row sk-in">
      <div class="skeleton sk-av" />
      <div class="skeleton sk-msg sk-sm" style="width: 22%" />
    </div>
    <div class="sk-row sk-out">
      <div class="skeleton sk-msg" style="width: 50%" />
    </div>
  </div>
</template>

<style scoped>
.sk-thread {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px 14px;
  flex: 1;
}

.sk-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.sk-row.sk-out {
  justify-content: flex-end;
}

.sk-av {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sk-av-gap {
  width: 28px;
  flex-shrink: 0;
}

.sk-msg {
  height: 36px;
  border-radius: 18px;
}

.sk-msg.sk-sm {
  height: 36px;
}

.sk-out .sk-msg {
  border-radius: 18px 18px 4px 18px;
}

.sk-in .sk-msg {
  border-radius: 18px 18px 18px 4px;
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
    rgba(55, 65, 81, 0.35) 25%,
    rgba(75, 85, 99, 0.45) 50%,
    rgba(55, 65, 81, 0.35) 75%
  );
  background-size: 200% 100%;
}
</style>
