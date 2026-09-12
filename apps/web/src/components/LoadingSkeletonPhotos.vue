<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/lib/i18n'

type Variant = 'initial' | 'more'

const props = withDefaults(
  defineProps<{
    variant?: Variant
  }>(),
  { variant: 'initial' },
)

const { t } = useI18n()
const groupCount = computed(() => (props.variant === 'initial' ? 2 : 1))
const photoCount = computed(() => (props.variant === 'initial' ? 8 : 6))
</script>

<template>
  <div aria-busy="true" aria-live="polite">
    <h1 v-if="variant === 'initial'" class="page-title desktop-only sk-title">
      {{ t.photosTitle }}
    </h1>

    <p v-if="variant === 'initial'" class="muted sk-line sk-subtitle desktop-only">{{ t.loading }}</p>

    <section v-if="variant === 'initial'" class="card section">
      <h2 class="section-title sk-line">{{ t.albums }}</h2>
      <div class="toolbar">
        <div class="skeleton sk-input" />
        <div class="skeleton sk-btn ink" />
      </div>
      <div class="list">
        <div v-for="i in 4" :key="i" class="row">
          <div class="skeleton sk-line sk-name" />
          <div class="skeleton sk-icon" />
        </div>
      </div>
    </section>

    <section v-for="g in groupCount" :key="g" class="card section">
      <h2 class="section-title sk-line"> </h2>
      <div class="grid photos">
        <div v-for="i in photoCount" :key="i" class="sk-photo skeleton" />
      </div>
    </section>
  </div>
</template>

<style scoped>
.section {
  margin-bottom: var(--space-md);
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: inline-block;
  }
}

.sk-line {
  height: 14px;
  border-radius: 8px;
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

.sk-title {
  margin-bottom: 0;
}

.sk-input {
  flex: 1;
  height: var(--touch-min);
  border-radius: var(--radius-md);
}

.sk-btn {
  width: 110px;
  height: var(--touch-min);
  border-radius: var(--radius-md);
}

.sk-name {
  width: 60%;
}

.sk-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
}

.sk-photo {
  aspect-ratio: 1;
  border-radius: var(--radius-md);
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

  .placeholder:active {
    transform: none;
  }
}
</style>
