<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()

const imageThumbnailsEnabled = ref(auth.user?.imageThumbnailsEnabled ?? true)
const videoThumbnailsEnabled = ref(auth.user?.videoThumbnailsEnabled ?? true)
const trashAutoDeleteEnabled = ref(auth.user?.trashAutoDeleteEnabled ?? false)
const trashRetentionDays = ref(auth.user?.trashRetentionDays ?? 30)
const error = ref('')
const saving = ref(false)

async function savePrefs() {
  error.value = ''
  saving.value = true
  try {
    await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({
        imageThumbnailsEnabled: imageThumbnailsEnabled.value,
        videoThumbnailsEnabled: videoThumbnailsEnabled.value,
        trashAutoDeleteEnabled: trashAutoDeleteEnabled.value,
        trashRetentionDays: Number(trashRetentionDays.value),
      }),
    })
    await auth.loadMe()
    ui.showToast('Settings saved', 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Save failed')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="page-title desktop-only">Settings</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <section class="card section">
      <h2 class="section-title">Trash</h2>
      <label class="toggle-row">
        <input v-model="trashAutoDeleteEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">Auto-delete trash</span>
      </label>
      <p class="field-hint">Permanently delete items older than the retention period.</p>
      <label class="field">
        <span>Retention days</span>
        <input v-model.number="trashRetentionDays" type="number" min="1" inputmode="numeric" />
      </label>
      <button class="btn save-btn" type="button" :disabled="saving" @click="savePrefs">
        {{ saving ? 'Saving…' : 'Save trash settings' }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">Media previews</h2>
      <p class="field-hint section-hint">
        Turn previews off to reduce object-storage bandwidth while browsing Photos.
      </p>
      <label class="toggle-row">
        <input v-model="imageThumbnailsEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">Show image thumbnails</span>
      </label>
      <label class="toggle-row">
        <input v-model="videoThumbnailsEnabled" type="checkbox" class="switch-input" />
        <span class="switch-track" aria-hidden="true"></span>
        <span class="toggle-label">Show video previews</span>
      </label>
      <button class="btn" type="button" :disabled="saving" @click="savePrefs">
        {{ saving ? 'Saving…' : 'Save preview settings' }}
      </button>
    </section>
  </div>
</template>

<style scoped>
.section {
  margin-bottom: var(--space-md);
}

.section-hint {
  margin: calc(var(--space-xs) * -1) 0 var(--space-sm);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: var(--touch-min);
  cursor: pointer;
}

.switch-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-track {
  position: relative;
  flex-shrink: 0;
  width: 44px;
  height: 24px;
  border-radius: var(--radius-pill);
  background: var(--hairline);
  transition: background-color var(--motion-press) var(--ease-standard);
}

.switch-track::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--canvas);
  box-shadow: 0 1px 3px rgba(17, 24, 39, 0.25);
  transition: transform var(--motion-press) var(--ease-standard);
}

.switch-input:checked + .switch-track {
  background: var(--accent);
}

.switch-input:checked + .switch-track::after {
  transform: translateX(20px);
}

.switch-input:focus-visible + .switch-track {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.toggle-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink);
}

.save-btn {
  width: 100%;
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .save-btn {
    width: auto;
  }

  .desktop-only {
    display: block;
  }
}

@media (prefers-reduced-motion: reduce) {
  .switch-track,
  .switch-track::after {
    transition: none;
  }
}
</style>
