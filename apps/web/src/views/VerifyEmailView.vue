<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import { formatAuthError } from '@/api/errors'
import { useI18n } from '@/lib/i18n'
import { useResendCooldown } from '@/lib/useResendCooldown'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const { locale } = useI18n()

const copy = computed(() =>
  locale.value === 'en'
    ? {
        title: 'Verify email',
        subtitlePrefix: 'Enter the 6-digit code sent to',
        emailFallback: 'your email',
        code: 'Verification code',
        codeHint: 'Code expires in 15 minutes.',
        verifying: 'Verifying…',
        verify: 'Verify email',
        sending: 'Sending…',
        resend: 'Resend code',
        resendIn: (seconds: number) => `Resend in ${seconds}s`,
        cooldown: (seconds: number) =>
          `You can request another code in ${seconds} seconds. You can still verify the current code now.`,
        sent: 'A new code was sent. It expires in 15 minutes.',
        resendFailed: 'Could not resend code. Try again.',
        invalidCode: 'Invalid or expired code. Try again.',
        wrongEmail: 'Wrong email? You can sign out of this unverified session and register again.',
        switching: 'Switching…',
        differentEmail: 'Use a different email',
      }
    : {
        title: 'Xác minh email',
        subtitlePrefix: 'Nhập mã 6 chữ số đã được gửi tới',
        emailFallback: 'email của bạn',
        code: 'Mã xác minh',
        codeHint: 'Mã sẽ hết hạn sau 15 phút.',
        verifying: 'Đang xác minh…',
        verify: 'Xác minh email',
        sending: 'Đang gửi…',
        resend: 'Gửi lại mã',
        resendIn: (seconds: number) => `Gửi lại sau ${seconds} giây`,
        cooldown: (seconds: number) =>
          `Bạn có thể yêu cầu mã khác sau ${seconds} giây. Bạn vẫn có thể xác minh mã hiện tại ngay bây giờ.`,
        sent: 'Mã mới đã được gửi và sẽ hết hạn sau 15 phút.',
        resendFailed: 'Không thể gửi lại mã. Hãy thử lại.',
        invalidCode: 'Mã không hợp lệ hoặc đã hết hạn. Hãy thử lại.',
        wrongEmail: 'Sai email? Bạn có thể đăng xuất khỏi phiên chưa xác minh này và đăng ký lại.',
        switching: 'Đang chuyển…',
        differentEmail: 'Dùng email khác',
      },
)

onMounted(() => {
  if (auth.isAuthenticated && auth.isVerified) {
    void router.replace('/')
  }
})

const code = ref('')
const message = ref('')
const error = ref('')
const verifying = ref(false)
const resending = ref(false)
const switchingEmail = ref(false)
const { remainingSeconds, active: resendCooldownActive, start: startResendCooldown } = useResendCooldown(60)

const requestBusy = computed(() => verifying.value || resending.value || switchingEmail.value)
const resendDisabled = computed(() => requestBusy.value || resendCooldownActive.value)
const resendLabel = computed(() => {
  if (resending.value) return copy.value.sending
  if (resendCooldownActive.value) return copy.value.resendIn(remainingSeconds.value)
  return copy.value.resend
})
const subtitle = computed(
  () => `${copy.value.subtitlePrefix} ${auth.user?.email ?? copy.value.emailFallback}.`,
)

function onCodeInput(event: Event) {
  const input = event.target as HTMLInputElement
  code.value = input.value.replace(/\D/g, '').slice(0, 6)
}

async function resend() {
  if (!auth.user?.email || resendDisabled.value) return
  error.value = ''
  message.value = ''
  resending.value = true
  try {
    await auth.resendVerification(auth.user.email)
    message.value = copy.value.sent
    startResendCooldown()
  } catch (e) {
    error.value = formatAuthError(e, copy.value.resendFailed)
  } finally {
    resending.value = false
  }
}

async function submit() {
  if (!auth.user?.email || requestBusy.value) return
  error.value = ''
  message.value = ''
  verifying.value = true
  try {
    await auth.verifyEmail(auth.user.email, code.value)
    await router.replace('/')
  } catch (e) {
    error.value = formatAuthError(e, copy.value.invalidCode)
  } finally {
    verifying.value = false
  }
}

async function useDifferentEmail() {
  if (requestBusy.value) return
  switchingEmail.value = true
  error.value = ''
  message.value = ''
  await auth.logout()
  await router.replace('/register')
}
</script>

<template>
  <AuthCard :title="copy.title" :subtitle="subtitle">
    <template #default="{ titleId }">
      <form :aria-labelledby="titleId" :aria-describedby="error ? 'auth-error' : undefined" @submit.prevent="submit">
        <label class="field">
          <span class="field-label">{{ copy.code }}</span>
          <input
            id="verify-code"
            class="auth-otp"
            :value="code"
            name="otp"
            type="text"
            inputmode="numeric"
            pattern="[0-9]*"
            maxlength="6"
            required
            autocomplete="one-time-code"
            enterkeyhint="done"
            :disabled="requestBusy"
            :aria-invalid="error ? true : undefined"
            @input="onCodeInput"
          />
          <span class="field-hint">{{ copy.codeHint }}</span>
        </label>

        <output v-if="message" class="auth-notice" role="status" aria-live="polite">{{ message }}</output>
        <div v-if="error" id="auth-error" class="auth-alert" role="alert">{{ error }}</div>

        <div class="auth-actions">
          <button class="btn ink block" type="submit" :disabled="requestBusy" :aria-busy="verifying">
            {{ verifying ? copy.verifying : copy.verify }}
          </button>
          <button
            class="btn block"
            type="button"
            :disabled="resendDisabled"
            :aria-busy="resending"
            :aria-describedby="resendCooldownActive ? 'resend-cooldown-status' : undefined"
            @click="resend"
          >
            {{ resendLabel }}
          </button>
          <span v-if="resendCooldownActive" id="resend-cooldown-status" class="field-hint">
            {{ copy.cooldown(remainingSeconds) }}
          </span>
        </div>

        <div class="auth-recovery">
          <p class="field-hint">{{ copy.wrongEmail }}</p>
          <button
            class="btn block ghost"
            type="button"
            :disabled="requestBusy"
            :aria-busy="switchingEmail"
            @click="useDifferentEmail"
          >
            {{ switchingEmail ? copy.switching : copy.differentEmail }}
          </button>
        </div>
      </form>
    </template>
  </AuthCard>
</template>

<style scoped>
.auth-recovery {
  display: grid;
  gap: var(--space-xs);
  margin-top: var(--space-md);
  padding-top: var(--space-md);
  border-top: 1px solid var(--hairline);
}

.auth-recovery .field-hint {
  margin: 0;
}
</style>
