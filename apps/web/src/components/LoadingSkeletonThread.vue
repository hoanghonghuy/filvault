<template>
  <div class="sk-thread" aria-busy="true" aria-live="polite">
    <div class="sk-row sk-incoming">
      <div class="sk sk-avatar" />
      <div class="sk sk-bubble" style="width: 52%" />
    </div>
    <div class="sk-row sk-incoming sk-followup">
      <div class="sk-spacer" />
      <div class="sk sk-bubble" style="width: 38%" />
    </div>

    <div class="sk-row sk-outgoing">
      <div class="sk sk-bubble" style="width: 44%" />
    </div>
    <div class="sk-row sk-outgoing">
      <div class="sk sk-bubble" style="width: 28%" />
    </div>

    <div class="sk-row sk-incoming">
      <div class="sk sk-avatar" />
      <div class="sk sk-bubble" style="width: 62%" />
    </div>

    <div class="sk-row sk-outgoing">
      <div class="sk sk-bubble" style="width: 35%" />
    </div>

    <div class="sk-row sk-incoming">
      <div class="sk sk-avatar" />
      <div class="sk sk-bubble sk-short" style="width: 20%" />
    </div>

    <div class="sk-row sk-outgoing">
      <div class="sk sk-bubble" style="width: 55%" />
    </div>
  </div>
</template>

<style scoped>
.sk-thread {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  flex: 1;
}

.sk-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.sk-row.sk-outgoing {
  justify-content: flex-end;
}

.sk-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sk-spacer {
  width: 28px;
  flex-shrink: 0;
}

.sk-bubble {
  height: 34px;
}

.sk-bubble.sk-short {
  height: 34px;
}

.sk-outgoing .sk-bubble {
  border-radius: 18px 18px 4px 18px;
}

.sk-incoming .sk-bubble {
  border-radius: 18px 18px 18px 4px;
}

.sk-incoming:not(.sk-followup) .sk-bubble {
  border-radius: 18px 18px 18px 4px;
}

.sk {
  background: rgba(255, 255, 255, 0.06);
  animation: sk-pulse 1.4s ease-in-out infinite;
}

@keyframes sk-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.7; }
}

[data-theme='light'] .sk,
:root:not([data-theme='dark']) .sk {
  background: rgba(0, 0, 0, 0.08);
}
</style>
