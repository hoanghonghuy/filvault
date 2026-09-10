<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import { api, ApiError } from '@/api/client'

const router = useRouter()

type Phase = 'request' | 'reset'

const phase = ref<Phase>('request')
const email = ref('')
const token = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const requestLoading = ref(false)
const resetLoading = ref(false)
const requestComplete = ref(false)
const error = ref('')

const requestDisabled = computed(() => requestLoading.value || !email.value.trim())
const resetDisabled = computed(
  () =>
    resetLoading.value ||
    !token.value.trim() ||
    !newPassword.value ||
    !confirmPassword.value,
)

async function requestReset() {
  if (requestDisabled.value) return
  error.value = ''
  requestLoading.value = true
  try {
    await api<void>(
      '/auth/password/forgot',
      { method: 'POST', body: JSON.stringify({ email: email.value.trim() }) },
      { auth: false },
    )
    requestComplete.value = true
    phase.value = 'reset'
  } catch {
    error.value = 'Could not request a password reset. Check your connection and try again.'
  } finally {
    requestLoading.value = false
  }
}

async function resetPassword() {
  if (resetDisabled.value) return
  error.value = ''

  if (newPassword.value.length < 8) {
    error.value = 'Use at least 8 characters for your new password.'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Passwords do not match.'
    return
  }

  resetLoading.value = true
  try {
    await api<void>(
      '/auth/password/reset',
      {
        method: 'POST',
        body: JSON.stringify({ token: token.value.trim(), newPassword: newPassword.value }),
      },
      { auth: false },
    )
    await router.replace({ name: 'login', query: { reset: 'success' } })
  } catch (cause) {
    if (cause instanceof ApiError && cause.status === 400) {
      error.value = 'This reset token is invalid or expired. Request a new one and try again.'
    } else {
      error.value = 'Could not reset your password. Check your connection and try again.'
    }
  } finally {
    resetLoading.value = false
  }
}

function startOver() {
  phase.value = 'request'
  requestComplete.value = false
  token.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  error.value = ''
}
</script>

<template>
  <AuthCard
    :title="phase === 'request' ? 'Reset your password' : 'Enter your reset token'"
    :subtitle="
      phase === 'request'
        ? 'Enter your account email to request a password reset.'
        : 'Paste the reset token from your email, then choose a new password.'
    "
  >
    <template #default="{ titleId }">
      <form
        v-if="phase === 'request'"
        :aria-labelledby="titleId"
        :aria-describedby="error ? 'password-recovery-error' : undefined"
        @submit.prevent="requestReset"
      >
        <label class="field">
          <span class="field-label">Email</span>
          <input
            id="recovery-email"
            v-model="email"
            type="email"
            name="email"
            required
            autocomplete="email"
            inputmode="email"
            autocapitalize="none"
            autocorrect="off"
            spellcheck="false"
            :disabled="requestLoading"
          />
        </label>

        <div v-if="error" id="password-recovery-error" class="auth-alert" role="alert">{{ error }}</div>

        <button
          type="submit"
          class="btn ink block auth-submit"
          :disabled="requestDisabled"
          :aria-busy="requestLoading"
        >
          {{ requestLoading ? 'Sending…' : 'Send reset instructions' }}
        </button>

        <button type="button" class="token-ready-btn" @click="phase = 'reset'">
          I already have a reset token
        </button>
      </form>

      <form
        v-else
        :aria-labelledby="titleId"
        :aria-describedby="error ? 'password-recovery-error' : undefined"
        @submit.prevent="resetPassword"
      >
        <div v-if="requestComplete" class="privacy-status" role="status">
          If an account exists for <strong>{{ email.trim() }}</strong>, reset instructions were sent.
        </div>

        <label class="field">
          <span class="field-label">Reset token</span>
          <input
            id="recovery-token"
            v-model="token"
            type="text"
            name="token"
            required
            autocomplete="one-time-code"
            autocapitalize="none"
            autocorrect="off"
            spellcheck="false"
            :disabled="resetLoading"
          />
        </label>

        <div class="field">
          <label class="field-label" for="recovery-password">New password</label>
          <PasswordInput
            id="recovery-password"
            v-model="newPassword"
            name="new-password"
            autocomplete="new-password"
            required
            :minlength="8"
            :disabled="resetLoading"
            :aria-invalid="Boolean(error)"
            :aria-describedby="error ? 'password-recovery-error' : 'password-recovery-requirement'"
          />
          <p id="password-recovery-requirement" class="password-requirement">Use at least 8 characters.</p>
        </div>

        <div class="field">
          <label class="field-label" for="recovery-confirm-password">Confirm new password</label>
          <PasswordInput
            id="recovery-confirm-password"
            v-model="confirmPassword"
            name="confirm-password"
            autocomplete="new-password"
            required
            :minlength="8"
            :disabled="resetLoading"
            :aria-invalid="Boolean(error)"
            :aria-describedby="error ? 'password-recovery-error' : undefined"
          />
        </div>

        <div v-if="error" id="password-recovery-error" class="auth-alert" role="alert">{{ error }}</div>

        <button
          type="submit"
          class="btn ink block auth-submit"
          :disabled="resetDisabled"
          :aria-busy="resetLoading"
        >
          {{ resetLoading ? 'Resetting…' : 'Reset password' }}
        </button>

        <button type="button" class="token-ready-btn" :disabled="resetLoading" @click="startOver">
          Request a new reset token
        </button>
      </form>
    </template>

    <template #footer>
      <p class="auth-switch">
        Remembered your password?
        <RouterLink class="link-accent" to="/login">Back to sign in</RouterLink>
      </p>
    </template>
  </AuthCard>
</template>

<style scoped>
.privacy-status {
  margin-bottom: var(--space-md);
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  color: var(--ink);
  font-size: 0.875rem;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.password-requirement {
  margin: var(--space-xxs) 0 0;
  color: var(--muted);
  font-size: 0.75rem;
  line-height: 1.4;
}

.token-ready-btn {
  display: flex;
  width: 100%;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  margin-top: var(--space-xs);
  border: 0;
  background: transparent;
  color: var(--accent);
  font: inherit;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
}

.token-ready-btn:hover:not(:disabled) {
  text-decoration: underline;
}

.token-ready-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.token-ready-btn:disabled {
  cursor: default;
  opacity: 0.55;
}
</style>
