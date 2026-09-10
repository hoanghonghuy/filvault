<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ApiError, api, formatBytes } from '@/api/client'
import type { PublicShareMeta } from '@/api/types'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'

const route = useRoute()
const { locale } = useI18n()

type MetaState = 'loading' | 'ready' | 'retryable-error' | 'unavailable'

const metaState = ref<MetaState>('loading')
const meta = ref<PublicShareMeta | null>(null)
const downloading = ref(false)
const downloadError = ref(false)

const copy = computed(() =>
  locale.value === 'vi'
    ? {
        loading: 'Đang tải liên kết…',
        unavailable: 'Liên kết này không còn khả dụng',
        temporary: 'Tạm thời không thể tải liên kết này.',
        retry: 'Thử lại',
        expires: 'Hết hạn',
        download: 'Tải xuống',
        downloading: 'Đang chuẩn bị tải xuống…',
        downloadFailed: 'Không thể chuẩn bị tệp tải xuống. Liên kết vẫn còn hiệu lực; hãy thử lại.',
        sharedVia: 'Được chia sẻ qua Filvault',
      }
    : {
        loading: 'Loading link…',
        unavailable: 'This link is not available',
        temporary: 'This link could not be loaded right now.',
        retry: 'Retry',
        expires: 'Expires',
        download: 'Download',
        downloading: 'Preparing download…',
        downloadFailed: 'Could not prepare the download. The link is still available; try again.',
        sharedVia: 'Shared via Filvault',
      },
)

const expiresLabel = computed(() => {
  if (!meta.value?.expiresAt) return ''
  const date = new Date(meta.value.expiresAt)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(locale.value === 'vi' ? 'vi-VN' : 'en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
})

function isTerminalMetadataError(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 404 || error.code === 'NOT_FOUND')
}

async function loadMeta() {
  metaState.value = 'loading'
  downloadError.value = false
  try {
    meta.value = await api<PublicShareMeta>(`/public/shares/${route.params.token}`)
    metaState.value = 'ready'
  } catch (error) {
    meta.value = null
    metaState.value = isTerminalMetadataError(error) ? 'unavailable' : 'retryable-error'
  }
}

async function download() {
  if (!meta.value || downloading.value) return
  downloading.value = true
  downloadError.value = false
  try {
    const out = await api<{ downloadUrl: string }>(`/public/shares/${route.params.token}/download`)
    window.location.href = out.downloadUrl
  } catch {
    downloadError.value = true
  } finally {
    downloading.value = false
  }
}

onMounted(loadMeta)
</script>

<template>
  <div class="public-page">
    <div class="share-card">
      <span class="brand-mark" aria-hidden="true">F</span>

      <div v-if="metaState === 'loading'" class="state" aria-busy="true" aria-live="polite">
        {{ copy.loading }}
      </div>

      <div v-else-if="metaState === 'unavailable'" class="state error-state" role="status">
        <Icon name="alert" :size="28" class="error-icon" />
        <p>{{ copy.unavailable }}</p>
      </div>

      <div v-else-if="metaState === 'retryable-error'" class="state error-state" role="alert">
        <Icon name="alert" :size="28" class="error-icon" />
        <p>{{ copy.temporary }}</p>
        <button type="button" class="btn ghost retry-btn" @click="loadMeta">{{ copy.retry }}</button>
      </div>

      <template v-else-if="meta">
        <h1 class="file-name">{{ meta.name }}</h1>
        <p class="file-meta">
          {{ meta.mimeType }} · {{ formatBytes(meta.sizeBytes) }}
          <template v-if="expiresLabel"> · {{ copy.expires }} {{ expiresLabel }} </template>
        </p>

        <p v-if="downloadError" class="download-error" role="alert">{{ copy.downloadFailed }}</p>
        <button
          type="button"
          class="btn ink block download-btn"
          :disabled="downloading"
          :aria-busy="downloading ? 'true' : undefined"
          @click="download"
        >
          {{ downloading ? copy.downloading : copy.download }}
        </button>
      </template>

      <p class="caption muted">{{ copy.sharedVia }}</p>
    </div>
  </div>
</template>

<style scoped>
.public-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: max(var(--space-md), env(safe-area-inset-top)) max(var(--space-md), env(safe-area-inset-right))
    max(var(--space-md), env(safe-area-inset-bottom)) max(var(--space-md), env(safe-area-inset-left));
  background: var(--surface-soft);
}

.share-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-md);
  width: 100%;
  max-width: 420px;
  min-width: 0;
  padding: var(--space-xl);
  border-radius: var(--radius-xl);
  background: var(--surface);
  box-shadow: 0 1px 3px rgb(0 0 0 / 8%);
  text-align: center;
}

.brand-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg);
  background: var(--accent);
  color: var(--on-accent);
  font-size: 1.5rem;
  font-weight: 700;
}

.file-name {
  margin: 0;
  max-width: 100%;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--ink);
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.file-meta {
  margin: 0;
  max-width: 100%;
  font-size: 0.875rem;
  color: var(--muted);
  overflow-wrap: anywhere;
}

.download-btn,
.retry-btn {
  min-height: 44px;
}

.download-btn {
  width: 100%;
}

.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  max-width: 100%;
  color: var(--muted);
}

.state p,
.download-error {
  margin: 0;
}

.error-icon,
.download-error {
  color: var(--danger);
}

.download-error {
  width: 100%;
  font-size: 0.875rem;
  line-height: 1.45;
}

.caption {
  margin: 0;
  font-size: 0.75rem;
}

.muted {
  color: var(--muted);
}

@media (max-width: 479px) {
  .share-card {
    padding: var(--space-lg);
  }
}
</style>
