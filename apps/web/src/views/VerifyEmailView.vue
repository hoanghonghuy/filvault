<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const code = ref('')
const message = ref('')
const error = ref('')
const loading = ref(false)

async function resend() {
  if (!auth.user?.email) return
  error.value = ''
  message.value = ''
  loading.value = true
  try {
    await auth.resendVerification(auth.user.email)
    message.value = 'Verification code sent. Check console logs in dev.'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Resend failed'
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!auth.user?.email) return
  error.value = ''
  message.value = ''
  loading.value = true
  try {
    await auth.verifyEmail(auth.user.email, code.value)
    await router.push('/files')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Verification failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="card auth-card">
      <h1>Verify email</h1>
      <p class="muted">Enter the 6-digit code sent to {{ auth.user?.email }}.</p>
      <form @submit.prevent="submit">
        <label class="field">
          <span>Code</span>
          <input v-model="code" inputmode="numeric" maxlength="6" required />
        </label>
        <p v-if="message" class="muted">{{ message }}</p>
        <p v-if="error" class="error">{{ error }}</p>
        <div class="toolbar">
          <button class="btn ink" type="submit" :disabled="loading">Verify</button>
          <button class="btn" type="button" :disabled="loading" @click="resend">Resend code</button>
        </div>
      </form>
    </div>
  </div>
</template>
