<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import { formatAuthError } from '@/api/errors'
import { safeInternalPath } from '@/lib/safeInternalPath'
import { useI18n } from '@/lib/i18n'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { locale } = useI18n()

const copy = computed(() => locale.value === 'en' ? {
  title: 'Sign in',
  subtitle: 'Access your files and photos from any device.',
  resetSuccess: 'Password reset complete. Sign in with your new password.',
  email: 'Email',
  password: 'Password',
  forgotPassword: 'Forgot password?',
  failed: 'Sign in failed. Try again.',
  incorrectCredentials: 'Incorrect email or password.',
  signingIn: 'Signing in…',
  signIn: 'Sign in',
  noAccount: 'No account?',
  createAccount: 'Create account',
} : {
  title: 'Đăng nhập',
  subtitle: 'Truy cập tệp và ảnh của bạn trên mọi thiết bị.',
  resetSuccess: 'Đặt lại mật khẩu thành công. Hãy đăng nhập bằng mật khẩu mới.',
  email: 'Email',
  password: 'Mật khẩu',
  forgotPassword: 'Quên mật khẩu?',
  failed: 'Đăng nhập không thành công. Vui lòng thử lại.',
  incorrectCredentials: 'Email hoặc mật khẩu không đúng.',
  signingIn: 'Đang đăng nhập…',
  signIn: 'Đăng nhập',
  noAccount: 'Chưa có tài khoản?',
  createAccount: 'Tạo tài khoản',
})

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
    error.value = formatAuthError(e, copy.value.failed, {
      unauthorized: copy.value.incorrectCredentials,
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthCard :title="copy.title" :subtitle="copy.subtitle">
    <template #default="{ titleId }">
      <form :aria-labelledby="titleId" :aria-describedby="error ? 'auth-error' : undefined" @submit.prevent="submit">
        <div v-if="resetSucceeded" class="auth-success" role="status">
          {{ copy.resetSuccess }}
        </div>

        <label class="field">
          <span class="field-label">{{ copy.email }}</span>
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
            <label class="field-label" for="login-password">{{ copy.password }}</label>
            <RouterLink class="forgot-link" to="/forgot-password">{{ copy.forgotPassword }}</RouterLink>
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
          {{ loading ? copy.signingIn : copy.signIn }}
        </button>
      </form>
    </template>

    <template #footer>
      <p class="auth-switch">
        {{ copy.noAccount }}
        <RouterLink class="link-accent" to="/register">{{ copy.createAccount }}</RouterLink>
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
