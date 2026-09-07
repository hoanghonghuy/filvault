<script setup lang="ts">
import { computed, ref } from 'vue'
import { api, setTokens } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { userInitials } from '@/lib/userInitials'
import { useI18n } from '@/lib/i18n'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()
const { t } = useI18n()

const displayName = ref(auth.user?.displayName ?? '')
const currentPassword = ref('')
const newPassword = ref('')
const error = ref('')
const savingProfile = ref(false)
const changingPassword = ref(false)

const initials = computed(() =>
  userInitials(auth.user?.displayName ?? '', auth.user?.email ?? ''),
)

async function saveProfile() {
  error.value = ''
  savingProfile.value = true
  try {
    await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({ displayName: displayName.value.trim() }),
    })
    await auth.loadMe()
    ui.showToast(t.value.saveProfile, 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Save failed')
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
    ui.showToast(t.value.changePassword, 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Password change failed')
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
  <div class="profile-page">
    <h1 class="page-title desktop-only">{{ t.profile }}</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <section class="card section identity" aria-labelledby="profile-identity-heading">
      <div class="identity-row">
        <span class="avatar" aria-hidden="true">{{ initials }}</span>
        <div class="identity-copy">
          <h2 id="profile-identity-heading" class="identity-name">
            {{ auth.user?.displayName || 'Your account' }}
          </h2>
          <p class="identity-email">{{ auth.user?.email }}</p>
        </div>
      </div>
      <label class="field">
        <span>{{ t.displayName }}</span>
        <input v-model="displayName" autocomplete="nickname" maxlength="80" />
      </label>
      <button class="btn ink save-btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? t.saving : t.saveProfile }}
      </button>
    </section>

    <section class="card section" aria-labelledby="profile-password-heading">
      <h2 id="profile-password-heading" class="section-title">{{ t.password }}</h2>
      <label class="field">
        <span>{{ t.currentPassword }}</span>
        <input v-model="currentPassword" type="password" autocomplete="current-password" />
      </label>
      <label class="field">
        <span>{{ t.newPassword }}</span>
        <input v-model="newPassword" type="password" minlength="8" autocomplete="new-password" />
      </label>
      <button class="btn ink save-btn" type="button" :disabled="changingPassword" @click="changePassword">
        {{ changingPassword ? t.changing : t.changePassword }}
      </button>
    </section>

    <section class="card section">
      <button class="btn danger block" type="button" @click="logout">{{ t.logOut }}</button>
    </section>
  </div>
</template>

<style scoped>
.section {
  margin-bottom: var(--space-md);
}

.identity-row {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  margin-bottom: var(--space-md);
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-pill);
  background: var(--accent-soft);
  color: var(--accent-hover);
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.identity-copy {
  min-width: 0;
}

.identity-name {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.identity-email {
  margin: 2px 0 0;
  font-size: 14px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
</style>
