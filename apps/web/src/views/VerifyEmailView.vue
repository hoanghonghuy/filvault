<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import { formatAuthError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const code = ref('')
const message = ref('')
const error = ref('')
const verifying = ref(false)
const resending = ref(false)

const busy = computed(() => verifying.value || resending.value)

function onCodeInput(event: Event) {
  const input = event.target as HTMLInputElement
  code.value = input.value.replace(/\D/g, '').slice(0, 6)
}

async function resend() {
  if (!auth.user?.email || busy.value) return
  error.value = ''
  message.value = ''
  resending.value = true
  try {
    await auth.resendVerification(auth.user.email)
    message.value = 'A new code was sent. It expires in 15 minutes.'
  } catch (e) {
    error.value = formatAuthError(e, 'Could not resend code. Try again.')
  } finally {
    resending.value = false
  }
}

async function submit() {
  if (!auth.user?.email || busy.value) return
  error.value = ''
  message.value = ''
  verifying.value = true
  try {
    await auth.verifyEmail(auth.user.email, code.value)
    await router.push('/files')
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
            :disabled="busy"
            :aria-invalid="error ? true : undefined"
            @input="onCodeInput"
          />
          <span class="field-hint">Code expires in 15 minutes.</span>
        </label>

        <output v-if="message" class="auth-notice">{{ message }}</output>
        <div v-if="error" id="auth-error" class="auth-alert" role="alert">{{ error }}</div>

        <div class="auth-actions">
          <button class="btn ink block" type="submit" :disabled="busy" :aria-busy="verifying">
            {{ verifying ? 'Verifying…' : 'Verify email' }}
          </button>
          <button class="btn block" type="button" :disabled="busy" :aria-busy="resending" @click="resend">
            {{ resending ? 'Sending…' : 'Resend code' }}
          </button>
        </div>
      </form>
    </template>
  </AuthCard>
</template>
