<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : null
    if (!auth.isVerified) {
      await router.push('/verify-email')
      return
    }
    await router.push(redirect ?? '/files')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="card auth-card">
      <h1>Login</h1>
      <form @submit.prevent="submit">
        <label class="field">
          <span>Email</span>
          <input v-model="email" type="email" required autocomplete="email" />
        </label>
        <label class="field">
          <span>Password</span>
          <input v-model="password" type="password" required autocomplete="current-password" />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="btn primary" type="submit" :disabled="loading">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
      <p class="muted">
        No account?
        <RouterLink to="/register">Register</RouterLink>
      </p>
    </div>
  </div>
</template>
