<script setup lang="ts">
import { ref } from 'vue'
import { ApiError, api, setTokens } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/api/types'

const auth = useAuthStore()

const displayName = ref(auth.user?.displayName ?? '')
const trashAutoDeleteEnabled = ref(auth.user?.trashAutoDeleteEnabled ?? false)
const trashRetentionDays = ref(auth.user?.trashRetentionDays ?? 30)
const currentPassword = ref('')
const newPassword = ref('')
const message = ref('')
const error = ref('')

async function saveProfile() {
  error.value = ''
  message.value = ''
  try {
    const user = await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({
        displayName: displayName.value,
        trashAutoDeleteEnabled: trashAutoDeleteEnabled.value,
        trashRetentionDays: Number(trashRetentionDays.value),
      }),
    })
    await auth.loadMe()
    message.value = 'Settings saved.'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Save failed'
  }
}

async function changePassword() {
  error.value = ''
  message.value = ''
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
    message.value = 'Password updated.'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Password change failed'
  }
}
</script>

<template>
  <div>
    <h1 class="page-title">Settings</h1>
    <p v-if="message" class="muted">{{ message }}</p>
    <p v-if="error" class="error">{{ error }}</p>

    <section class="card" style="margin-bottom: 1rem">
      <h2>Profile</h2>
      <label class="field">
        <span>Display name</span>
        <input v-model="displayName" />
      </label>
      <button class="btn primary" type="button" @click="saveProfile">Save profile</button>
    </section>

    <section class="card" style="margin-bottom: 1rem">
      <h2>Trash</h2>
      <label class="field">
        <span>
          <input v-model="trashAutoDeleteEnabled" type="checkbox" />
          Auto-delete trash
        </span>
      </label>
      <label class="field">
        <span>Retention days</span>
        <input v-model.number="trashRetentionDays" type="number" min="1" />
      </label>
      <button class="btn" type="button" @click="saveProfile">Save trash settings</button>
    </section>

    <section class="card">
      <h2>Password</h2>
      <label class="field">
        <span>Current password</span>
        <input v-model="currentPassword" type="password" autocomplete="current-password" />
      </label>
      <label class="field">
        <span>New password</span>
        <input v-model="newPassword" type="password" minlength="8" autocomplete="new-password" />
      </label>
      <button class="btn primary" type="button" @click="changePassword">Change password</button>
    </section>
  </div>
</template>
