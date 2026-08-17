<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, api, setTokens } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()

const displayName = ref(auth.user?.displayName ?? '')
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
        <input v-model="trashAutoDeleteEnabled" type="checkbox" />
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

.toggle-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: var(--touch-min);
  cursor: pointer;
}

.toggle-row input[type='checkbox'] {
  width: 18px;
  height: 18px;
  accent-color: var(--accent);
  cursor: pointer;
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
