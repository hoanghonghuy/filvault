<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import { formatAuthError } from '@/api/errors'
import { safeInternalPath } from '@/lib/safeInternalPath'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(() => {
  if (auth.isAuthenticated) {
    void router.replace(auth.isVerified ? '/' : '/verify-email')
  }
})

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const resetSucceeded = computed(() => route.query.reset === 'success')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value.trim(), password.value)
    if (!auth.isVerified) {
      await router.replace('/verify-email')
      return
    }
    await router.replace(safeInternalPath(route.query.redirect, '/'))
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
        <div v-if="resetSucceeded" class="auth-success" role="status">
          Password reset complete. Sign in with your new password.
        </div>

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
        <div class="field">
          <div class="password-label-row">
            <label class="field-label" for="login-password">Password</label>
            <RouterLink class="forgot-link" to="/forgot-password">Forgot password?</RouterLink>
          </div>
          <PasswordInput
            id="login-password"
            v-model="password"
            name="password"
            autocomplete="current-password"
            enterkeyhint="go"
            required
            :disabled="loading"
            :aria-invalid="Boolean(error)"
            :aria-describedby="error ? 'auth-error' : undefined"
          />
        </div>

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

<style scoped>
.password-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  margin-bottom: var(--space-xxs);
}

.password-label-row .field-label {
  margin: 0;
}

.forgot-link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  justify-content: flex-end;
  color: var(--accent);
  font-size: 0.8125rem;
  font-weight: 600;
  text-decoration: none;
  text-align: right;
}

.forgot-link:hover {
  text-decoration: underline;
}

.forgot-link:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.auth-success {
  margin-bottom: var(--space-md);
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--success, #22863a);
  border-radius: var(--radius-md);
  color: var(--ink);
  font-size: 0.875rem;
  line-height: 1.45;
}

@media (max-width: 420px) {
  .password-label-row {
    align-items: flex-start;
  }

  .forgot-link {
    max-width: 52%;
  }
}
</style>
