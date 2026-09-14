<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, formatBytes } from '@/api/client'
import type { PublicShareMeta } from '@/api/types'
import Icon from '@/components/AppIcon.vue'

const route = useRoute()

const loading = ref(true)
const failed = ref(false)
const meta = ref<PublicShareMeta | null>(null)

onMounted(async () => {
  try {
    meta.value = await api<PublicShareMeta>(`/public/shares/${route.params.token}`)
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
})

async function download() {
  if (!meta.value) return
  try {
    const out = await api<{ downloadUrl: string }>(`/public/shares/${route.params.token}/download`)
    window.location.href = out.downloadUrl
  } catch {
    failed.value = true
  }
}
</script>

<template>
  <div class="public-page">
    <div class="share-card">
      <span class="brand-mark" aria-hidden="true">F</span>

      <div v-if="loading" class="state" aria-busy="true" aria-live="polite">Loading…</div>

      <div v-else-if="failed || !meta" class="state error-state">
        <Icon name="alert" :size="28" class="error-icon" />
        <p>This link is not available</p>
      </div>

      <template v-else>
        <h1 class="file-name">{{ meta.name }}</h1>
        <p class="file-meta">
          {{ meta.mimeType }} · {{ formatBytes(meta.sizeBytes) }}
          <template v-if="meta.expiresAt">
            · Expires {{ new Date(meta.expiresAt).toLocaleString() }}
          </template>
        </p>
        <button type="button" class="btn ink block download-btn" @click="download">Download</button>
      </template>

      <p class="caption muted">Shared via Filvault</p>
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
  padding: var(--space-md);
  background: var(--surface-soft);
}

.share-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-md);
  width: 100%;
  max-width: 420px;
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
  font-size: 0.875rem;
  color: var(--muted);
  overflow-wrap: anywhere;
}

.download-btn {
  width: 100%;
}

.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  color: var(--muted);
}

.error-icon {
  color: var(--danger);
}

.caption {
  margin: 0;
  font-size: 0.75rem;
}

.muted {
  color: var(--muted);
}
</style>
