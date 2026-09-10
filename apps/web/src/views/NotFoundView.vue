<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Icon from '@/components/AppIcon.vue'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from '@/lib/i18n'

const router = useRouter()
const auth = useAuthStore()
const { locale } = useI18n()

const copy = computed(() => {
  if (locale.value === 'vi') {
    return {
      eyebrow: 'Không tìm thấy trang',
      title: 'Đường dẫn này không tồn tại',
      body: 'Liên kết có thể đã cũ hoặc địa chỉ đã được nhập sai. Bạn có thể quay lại nơi an toàn mà không mất thông tin tài khoản.',
      home: 'Về trang chủ',
      files: 'Mở Tệp của tôi',
      back: 'Quay lại',
    }
  }
  return {
    eyebrow: 'Page not found',
    title: 'This path does not exist',
    body: 'The link may be stale or the address may have been mistyped. You can recover safely without exposing account details.',
    home: 'Go home',
    files: 'Open My Files',
    back: 'Go back',
  }
})

const canGoBack = computed(() => window.history.length > 1)

function goBack() {
  if (canGoBack.value) router.back()
  else void router.push(auth.isAuthenticated ? '/' : '/login')
}
</script>

<template>
  <main class="not-found" aria-labelledby="not-found-title">
    <div class="not-found-card">
      <span class="icon-wrap" aria-hidden="true"><Icon name="search" :size="28" /></span>
      <p class="eyebrow">404 · {{ copy.eyebrow }}</p>
      <h1 id="not-found-title">{{ copy.title }}</h1>
      <p class="body-copy">{{ copy.body }}</p>

      <div class="actions">
        <RouterLink v-if="auth.isAuthenticated" to="/" class="btn ink">{{ copy.home }}</RouterLink>
        <RouterLink v-if="auth.isAuthenticated" to="/files" class="btn ghost">{{ copy.files }}</RouterLink>
        <RouterLink v-else to="/login" class="btn ink">{{ copy.home }}</RouterLink>
        <button type="button" class="btn ghost" @click="goBack">{{ copy.back }}</button>
      </div>
    </div>
  </main>
</template>

<style scoped>
.not-found {
  min-height: min(72vh, 720px);
  display: grid;
  place-items: center;
  padding: clamp(var(--space-md), 5vw, var(--space-2xl));
}

.not-found-card {
  width: min(100%, 560px);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: clamp(var(--space-lg), 5vw, var(--space-2xl));
  border: 1px solid var(--hairline);
  border-radius: var(--radius-lg);
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.icon-wrap {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  margin-bottom: var(--space-md);
  border-radius: var(--radius-md);
  background: var(--accent-soft);
  color: var(--accent);
}

.eyebrow {
  margin: 0 0 var(--space-xs);
  color: var(--muted);
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  color: var(--ink);
  font-size: clamp(1.75rem, 5vw, 2.5rem);
  line-height: 1.1;
}

.body-copy {
  max-width: 48ch;
  margin: var(--space-md) 0 0;
  color: var(--muted);
  line-height: 1.6;
}

.actions {
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-xs);
  margin-top: var(--space-lg);
}

.actions .btn {
  min-height: var(--touch-min);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
}

@media (max-width: 639px) {
  .not-found {
    place-items: start stretch;
    padding: var(--space-md);
  }

  .not-found-card {
    margin-top: max(var(--space-lg), env(safe-area-inset-top));
  }

  .actions {
    flex-direction: column;
  }

  .actions .btn {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .not-found-card,
  .actions .btn {
    scroll-behavior: auto;
    transition: none;
  }
}
</style>
