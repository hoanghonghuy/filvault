<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const displayName = ref('')
const inviteCode = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.register({
      email: email.value,
      password: password.value,
      displayName: displayName.value,
      inviteCode: inviteCode.value,
    })
    await router.push('/verify-email')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="card auth-card">
      <h1>Register</h1>
      <form @submit.prevent="submit">
        <label class="field">
          <span>Email</span>
          <input v-model="email" type="email" required autocomplete="email" />
        </label>
        <label class="field">
          <span>Display name</span>
          <input v-model="displayName" required />
        </label>
        <label class="field">
          <span>Password</span>
          <input v-model="password" type="password" required minlength="8" autocomplete="new-password" />
        </label>
        <label class="field">
          <span>Invite code</span>
          <input v-model="inviteCode" required />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="btn ink block" type="submit" :disabled="loading">
          {{ loading ? 'Creating…' : 'Create account' }}
        </button>
      </form>
      <p class="muted">
        Already have an account?
        <RouterLink to="/login">Login</RouterLink>
      </p>
    </div>
  </div>
</template>
