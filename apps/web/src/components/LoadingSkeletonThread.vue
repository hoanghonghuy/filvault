<template>
  <div class="sk-chat-thread" aria-busy="true" aria-live="polite">
    <!-- Incoming message cluster: bubble 1 with avatar -->
    <div class="sk-chat-row sk-chat-in">
      <div class="sk-chat-shimmer sk-chat-avatar" />
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-55" />
    </div>
    <!-- Incoming message cluster: bubble 2 without avatar -->
    <div class="sk-chat-row sk-chat-in">
      <div class="sk-chat-spacer" />
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-35" />
    </div>

    <!-- Outgoing message cluster: bubble 1 -->
    <div class="sk-chat-row sk-chat-out">
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-45" />
    </div>
    <!-- Outgoing message cluster: bubble 2 -->
    <div class="sk-chat-row sk-chat-out">
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-25" />
    </div>

    <!-- Incoming message with avatar -->
    <div class="sk-chat-row sk-chat-in">
      <div class="sk-chat-shimmer sk-chat-avatar" />
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-65" />
    </div>

    <!-- Outgoing message -->
    <div class="sk-chat-row sk-chat-out">
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-38" />
    </div>

    <!-- Incoming short response -->
    <div class="sk-chat-row sk-chat-in">
      <div class="sk-chat-shimmer sk-chat-avatar" />
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-20" />
    </div>

    <!-- Outgoing message -->
    <div class="sk-chat-row sk-chat-out">
      <div class="sk-chat-shimmer sk-chat-bubble sk-w-50" />
    </div>
  </div>
</template>

<style scoped>
.sk-chat-thread {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px 14px;
  flex: 1;
  width: 100%;
}

.sk-chat-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  width: 100%;
}

.sk-chat-row.sk-chat-out {
  justify-content: flex-end;
}

.sk-chat-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sk-chat-spacer {
  width: 28px;
  flex-shrink: 0;
}

.sk-chat-bubble {
  height: 36px;
  border-radius: 18px;
  max-width: 74%;
}

.sk-chat-out .sk-chat-bubble {
  border-radius: 18px 18px 4px 18px;
}

.sk-chat-in .sk-chat-bubble {
  border-radius: 18px 18px 18px 4px;
}

.sk-w-20 { width: 20%; min-width: 60px; }
.sk-w-25 { width: 25%; min-width: 75px; }
.sk-w-35 { width: 35%; min-width: 105px; }
.sk-w-38 { width: 38%; min-width: 115px; }
.sk-w-45 { width: 45%; min-width: 135px; }
.sk-w-50 { width: 50%; min-width: 150px; }
.sk-w-55 { width: 55%; min-width: 165px; }
.sk-w-65 { width: 65%; min-width: 195px; }

.sk-chat-shimmer {
  background: linear-gradient(
    90deg,
    rgba(229, 231, 235, 0.7) 25%,
    rgba(243, 244, 246, 0.95) 50%,
    rgba(229, 231, 235, 0.7) 75%
  );
  background-size: 200% 100%;
  animation: sk-chat-pulse 1.3s ease-in-out infinite;
}

@keyframes sk-chat-pulse {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

[data-theme='dark'] .sk-chat-shimmer {
  background: linear-gradient(
    90deg,
    rgba(255, 255, 255, 0.05) 25%,
    rgba(255, 255, 255, 0.1) 50%,
    rgba(255, 255, 255, 0.05) 75%
  );
  background-size: 200% 100%;
}

@media (prefers-reduced-motion: reduce) {
  .sk-chat-shimmer,
  [data-theme='dark'] .sk-chat-shimmer {
    animation: none;
    background: var(--hairline);
  }
}
</style>
