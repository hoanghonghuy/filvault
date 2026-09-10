<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import { formatAuthError } from '@/api/errors'
import { useResendCooldown } from '@/lib/useResendCooldown'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

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
const { remainingSeconds, active: resendCooldownActive, start: startResendCooldown } = useResendCooldown(60)

const requestBusy = computed(() => verifying.value || resending.value)
const resendDisabled = computed(() => requestBusy.value || resendCooldownActive.value)
const resendLabel = computed(() => {
  if (resending.value) return 'Sending…'
  if (resendCooldownActive.value) return `Resend in ${remainingSeconds.value}s`
  return 'Resend code'
})

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
    message.value = 'A new code was sent. It expires in 15 minutes.'
    startResendCooldown()
  } catch (e) {
    error.value = formatAuthError(e, 'Could not resend code. Try again.')
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
    error.value = formatAuthError(e, 'Invalid or expired code. Try again.')
  } finally {
    verifying.value = false
  }
}
</script>

<template>
  <AuthCard title="Verify email" :subtitle="`Enter the 6-digit code sent to ${auth.user?.email ?? 'your email'}.`">
    <template #default="{ titleId }">
      <form :aria-labelledby="titleId" :aria-describedby="error ? 'auth-error' : undefined" @submit.prevent="submit">
        <label class="field">
          <span class="field-label">Verification code</span>
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
          <span class="field-hint">Code expires in 15 minutes.</span>
        </label>

        <output v-if="message" class="auth-notice" role="status" aria-live="polite">{{ message }}</output>
        <div v-if="error" id="auth-error" class="auth-alert" role="alert">{{ error }}</div>

        <div class="auth-actions">
          <button class="btn ink block" type="submit" :disabled="requestBusy" :aria-busy="verifying">
            {{ verifying ? 'Verifying…' : 'Verify email' }}
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
            You can request another code in {{ remainingSeconds }} seconds. You can still verify the current code now.
          </span>
        </div>
      </form>
    </template>
  </AuthCard>
</template>
