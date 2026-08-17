<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import { formatAuthError } from '@/api/errors'
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
      email: email.value.trim(),
      password: password.value,
      displayName: displayName.value.trim(),
      inviteCode: inviteCode.value.trim(),
    })
    await router.push('/verify-email')
  } catch (e) {
    error.value = formatAuthError(e, 'Registration failed. Check your details and try again.')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthCard title="Create account" subtitle="Registration requires a valid invite code.">
    <template #default="{ titleId }">
      <form :aria-labelledby="titleId" :aria-describedby="error ? 'auth-error' : undefined" @submit.prevent="submit">
        <label class="field">
          <span class="field-label">Email</span>
          <input
            id="register-email"
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
          <span class="field-label">Display name</span>
          <input
            id="register-display-name"
            v-model="displayName"
            name="displayName"
            required
            autocomplete="nickname"
            maxlength="80"
            :disabled="loading"
            :aria-invalid="error ? true : undefined"
          />
        </label>
        <label class="field">
          <span class="field-label">Password</span>
          <input
            id="register-password"
            v-model="password"
            name="password"
            type="password"
            required
            minlength="8"
            autocomplete="new-password"
            :disabled="loading"
            :aria-invalid="error ? true : undefined"
          />
          <span class="field-hint">At least 8 characters.</span>
        </label>
        <label class="field">
          <span class="field-label">Invite code</span>
          <input
            id="register-invite"
            v-model="inviteCode"
            name="inviteCode"
            required
            autocomplete="off"
            autocapitalize="none"
            autocorrect="off"
            spellcheck="false"
            :disabled="loading"
            :aria-invalid="error ? true : undefined"
          />
          <span class="field-hint">Ask the person who invited you if you do not have a code.</span>
        </label>

        <div v-if="error" id="auth-error" class="auth-alert" role="alert">{{ error }}</div>

        <button class="btn ink block auth-submit" type="submit" :disabled="loading" :aria-busy="loading">
          {{ loading ? 'Creating account…' : 'Create account' }}
        </button>
      </form>
    </template>

    <template #footer>
      <p class="auth-switch">
        Already have an account?
        <RouterLink class="link-accent" to="/login">Sign in</RouterLink>
      </p>
    </template>
  </AuthCard>
</template>
