<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import { formatAuthError } from '@/api/errors'
import { safeInternalPath } from '@/lib/safeInternalPath'
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
    await auth.login(email.value.trim(), password.value)
    if (!auth.isVerified) {
      await router.push('/verify-email')
      return
    }
    await router.push(safeInternalPath(route.query.redirect, '/'))
  } catch (e) {
    error.value = formatAuthError(e, 'Sign in failed. Try again.')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthCard title="Sign in" subtitle="Access your files and photos from any device.">
    <template #default="{ titleId }">
      <form :aria-labelledby="titleId" :aria-describedby="error ? 'auth-error' : undefined" @submit.prevent="submit">
        <label class="field">
          <span class="field-label">Email</span>
          <input
            id="login-email"
            v-model="email"
            name="email"
            type="email"
            required
            autocomplete="username"
            inputmode="email"
            autocapitalize="none"
            autocorrect="off"
            spellcheck="false"
            :disabled="loading"
            :aria-invalid="error ? true : undefined"
          />
        </label>
        <label class="field">
          <span class="field-label">Password</span>
          <input
            id="login-password"
            v-model="password"
            name="password"
            type="password"
            required
            autocomplete="current-password"
            enterkeyhint="go"
            :disabled="loading"
            :aria-invalid="error ? true : undefined"
          />
        </label>

        <div v-if="error" id="auth-error" class="auth-alert" role="alert">{{ error }}</div>

        <button class="btn ink block auth-submit" type="submit" :disabled="loading" :aria-busy="loading">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
    </template>

    <template #footer>
      <p class="auth-switch">
        No account?
        <RouterLink class="link-accent" to="/register">Create account</RouterLink>
      </p>
    </template>
  </AuthCard>
</template>
