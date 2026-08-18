<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, api, setTokens } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()

const displayName = ref(auth.user?.displayName ?? '')
const imageThumbnailsEnabled = ref(auth.user?.imageThumbnailsEnabled ?? true)
const videoThumbnailsEnabled = ref(auth.user?.videoThumbnailsEnabled ?? true)
const trashAutoDeleteEnabled = ref(auth.user?.trashAutoDeleteEnabled ?? false)
const trashRetentionDays = ref(auth.user?.trashRetentionDays ?? 30)
const currentPassword = ref('')
const newPassword = ref('')
const error = ref('')
const savingProfile = ref(false)
const changingPassword = ref(false)

async function saveProfile() {
  error.value = ''
  savingProfile.value = true
  try {
    await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({
        displayName: displayName.value,
        imageThumbnailsEnabled: imageThumbnailsEnabled.value,
        videoThumbnailsEnabled: videoThumbnailsEnabled.value,
        trashAutoDeleteEnabled: trashAutoDeleteEnabled.value,
        trashRetentionDays: Number(trashRetentionDays.value),
      }),
    })
    await auth.loadMe()
    ui.showToast('Settings saved')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Save failed'
  } finally {
    savingProfile.value = false
  }
}

async function changePassword() {
  error.value = ''
  changingPassword.value = true
  try {
    const tokens = await api<{ accessToken: string; refreshToken: string }>('/users/me/password', {
      method: 'POST',
      body: JSON.stringify({
        currentPassword: currentPassword.value,
        newPassword: newPassword.value,
      }),
    })
    setTokens(tokens.accessToken, tokens.refreshToken)
    await auth.loadMe()
    currentPassword.value = ''
    newPassword.value = ''
    ui.showToast('Password updated')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Password change failed'
  } finally {
    changingPassword.value = false
  }
}

async function logout() {
  await auth.logout()
  window.location.href = '/login'
}
</script>

<template>
  <div>
    <h1 class="page-title desktop-only">Settings</h1>
    <p v-if="error" class="error">{{ error }}</p>

    <section class="card section">
      <h2 class="section-title">Profile</h2>
      <p class="muted">{{ auth.user?.email }}</p>
      <label class="field">
        <span>Display name</span>
        <input v-model="displayName" />
      </label>
      <button class="btn ink" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? 'Saving…' : 'Save profile' }}
      </button>
    </section>

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
        <input v-model.number="trashRetentionDays" type="number" min="1" />
      </label>
      <button class="btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? 'Saving…' : 'Save trash settings' }}
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
      <button class="btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? 'Saving…' : 'Save preview settings' }}
      </button>
    </section>

    <section class="card section">
      <h2 class="section-title">Password</h2>
      <label class="field">
        <span>Current password</span>
        <input v-model="currentPassword" type="password" autocomplete="current-password" />
      </label>
      <label class="field">
        <span>New password</span>
        <input v-model="newPassword" type="password" minlength="8" autocomplete="new-password" />
      </label>
      <button class="btn ink" type="button" :disabled="changingPassword" @click="changePassword">
        {{ changingPassword ? 'Changing…' : 'Change password' }}
      </button>
    </section>

    <section class="card section">
      <button class="btn danger block" type="button" @click="logout">Log out</button>
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
  transition: background-color 150ms ease;
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
  transition: transform 150ms ease;
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

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: block;
  }
}
</style>
