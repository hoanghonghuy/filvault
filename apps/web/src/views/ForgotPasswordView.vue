<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import { api, ApiError } from '@/api/client'
import { useI18n } from '@/lib/i18n'

const router = useRouter()
const { locale } = useI18n()

type Phase = 'request' | 'reset'

const copy = computed(() =>
  locale.value === 'en'
    ? {
        requestTitle: 'Reset your password',
        resetTitle: 'Enter your reset token',
        requestSubtitle: 'Enter your account email to request a password reset.',
        resetSubtitle: 'Paste the reset token from your email, then choose a new password.',
        email: 'Email',
        requestFailed: 'Could not request a password reset. Check your connection and try again.',
        sending: 'Sending…',
        send: 'Send reset instructions',
        haveToken: 'I already have a reset token',
        sentPrefix: 'If an account exists for',
        sentSuffix: ', reset instructions were sent.',
        token: 'Reset token',
        newPassword: 'New password',
        passwordHint: 'Use at least 8 characters.',
        confirmPassword: 'Confirm new password',
        tooShort: 'Use at least 8 characters for your new password.',
        mismatch: 'Passwords do not match.',
        invalidToken: 'This reset token is invalid or expired. Request a new one and try again.',
        resetFailed: 'Could not reset your password. Check your connection and try again.',
        resetting: 'Resetting…',
        reset: 'Reset password',
        requestNew: 'Request a new reset token',
        remembered: 'Remembered your password?',
        back: 'Back to sign in',
      }
    : {
        requestTitle: 'Đặt lại mật khẩu',
        resetTitle: 'Nhập mã đặt lại mật khẩu',
        requestSubtitle: 'Nhập email tài khoản để yêu cầu đặt lại mật khẩu.',
        resetSubtitle: 'Dán mã đặt lại từ email rồi chọn mật khẩu mới.',
        email: 'Email',
        requestFailed: 'Không thể gửi yêu cầu đặt lại mật khẩu. Hãy kiểm tra kết nối và thử lại.',
        sending: 'Đang gửi…',
        send: 'Gửi hướng dẫn đặt lại',
        haveToken: 'Tôi đã có mã đặt lại mật khẩu',
        sentPrefix: 'Nếu tồn tại tài khoản với email',
        sentSuffix: ', hướng dẫn đặt lại đã được gửi.',
        token: 'Mã đặt lại',
        newPassword: 'Mật khẩu mới',
        passwordHint: 'Sử dụng ít nhất 8 ký tự.',
        confirmPassword: 'Xác nhận mật khẩu mới',
        tooShort: 'Hãy sử dụng ít nhất 8 ký tự cho mật khẩu mới.',
        mismatch: 'Mật khẩu xác nhận không khớp.',
        invalidToken: 'Mã đặt lại không hợp lệ hoặc đã hết hạn. Hãy yêu cầu mã mới và thử lại.',
        resetFailed: 'Không thể đặt lại mật khẩu. Hãy kiểm tra kết nối và thử lại.',
        resetting: 'Đang đặt lại…',
        reset: 'Đặt lại mật khẩu',
        requestNew: 'Yêu cầu mã đặt lại mới',
        remembered: 'Đã nhớ mật khẩu?',
        back: 'Quay lại đăng nhập',
      },
)

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
    error.value = copy.value.requestFailed
  } finally {
    requestLoading.value = false
  }
}

async function resetPassword() {
  if (resetDisabled.value) return
  error.value = ''

  if (newPassword.value.length < 8) {
    error.value = copy.value.tooShort
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = copy.value.mismatch
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
      error.value = copy.value.invalidToken
    } else {
      error.value = copy.value.resetFailed
    }
  } finally {
    resetLoading.value = false
  }
}

function enterResetPhase() {
  error.value = ''
  phase.value = 'reset'
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
    :title="phase === 'request' ? copy.requestTitle : copy.resetTitle"
    :subtitle="phase === 'request' ? copy.requestSubtitle : copy.resetSubtitle"
  >
    <template #default="{ titleId }">
      <form
        v-if="phase === 'request'"
        :aria-labelledby="titleId"
        :aria-describedby="error ? 'password-recovery-error' : undefined"
        @submit.prevent="requestReset"
      >
        <label class="field">
          <span class="field-label">{{ copy.email }}</span>
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
          {{ requestLoading ? copy.sending : copy.send }}
        </button>

        <button type="button" class="token-ready-btn" @click="enterResetPhase">
          {{ copy.haveToken }}
        </button>
      </form>

      <form
        v-else
        :aria-labelledby="titleId"
        :aria-describedby="error ? 'password-recovery-error' : undefined"
        @submit.prevent="resetPassword"
      >
        <div v-if="requestComplete" class="privacy-status" role="status">
          {{ copy.sentPrefix }} <strong>{{ email.trim() }}</strong>{{ copy.sentSuffix }}
        </div>

        <label class="field">
          <span class="field-label">{{ copy.token }}</span>
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
          <label class="field-label" for="recovery-password">{{ copy.newPassword }}</label>
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
          <p id="password-recovery-requirement" class="password-requirement">{{ copy.passwordHint }}</p>
        </div>

        <div class="field">
          <label class="field-label" for="recovery-confirm-password">{{ copy.confirmPassword }}</label>
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
          {{ resetLoading ? copy.resetting : copy.reset }}
        </button>

        <button type="button" class="token-ready-btn" :disabled="resetLoading" @click="startOver">
          {{ copy.requestNew }}
        </button>
      </form>
    </template>

    <template #footer>
      <p class="auth-switch">
        {{ copy.remembered }}
        <RouterLink class="link-accent" to="/login">{{ copy.back }}</RouterLink>
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
