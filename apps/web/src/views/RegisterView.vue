<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AuthCard from '@/components/AuthCard.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import { formatAuthError } from '@/api/errors'
import { useI18n } from '@/lib/i18n'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const { locale } = useI18n()

const copy = computed(() => locale.value === 'en' ? {
  title: 'Create account',
  subtitle: 'Registration requires a valid invite code.',
  email: 'Email',
  displayName: 'Display name',
  password: 'Password',
  passwordHint: 'At least 8 characters.',
  confirmPassword: 'Confirm password',
  inviteCode: 'Invite code',
  inviteHint: 'Ask the person who invited you if you do not have a code.',
  mismatch: 'Passwords do not match. Re-enter the confirmation and try again.',
  failed: 'Registration failed. Check your details and try again.',
  conflict: 'An account with this email already exists.',
  invalidInvite: 'Invalid invite code.',
  disabled: 'Registration is currently disabled.',
  creating: 'Creating account…',
  create: 'Create account',
  existingAccount: 'Already have an account?',
  signIn: 'Sign in',
} : {
  title: 'Tạo tài khoản',
  subtitle: 'Đăng ký cần có mã mời hợp lệ.',
  email: 'Email',
  displayName: 'Tên hiển thị',
  password: 'Mật khẩu',
  passwordHint: 'Ít nhất 8 ký tự.',
  confirmPassword: 'Xác nhận mật khẩu',
  inviteCode: 'Mã mời',
  inviteHint: 'Hãy hỏi người đã mời bạn nếu bạn chưa có mã.',
  mismatch: 'Mật khẩu xác nhận không khớp. Hãy nhập lại và thử lại.',
  failed: 'Đăng ký không thành công. Hãy kiểm tra thông tin và thử lại.',
  conflict: 'Đã có tài khoản sử dụng email này.',
  invalidInvite: 'Mã mời không hợp lệ.',
  disabled: 'Tính năng đăng ký hiện đang tạm tắt.',
  creating: 'Đang tạo tài khoản…',
  create: 'Tạo tài khoản',
  existingAccount: 'Đã có tài khoản?',
  signIn: 'Đăng nhập',
})

onMounted(() => {
  if (auth.isAuthenticated) {
    void router.replace(auth.isVerified ? '/' : '/verify-email')
  }
})

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const displayName = ref('')
const inviteCode = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''

  if (password.value !== confirmPassword.value) {
    error.value = copy.value.mismatch
    return
  }

  loading.value = true
  try {
    await auth.register({
      email: email.value.trim(),
      password: password.value,
      displayName: displayName.value.trim(),
      inviteCode: inviteCode.value.trim(),
    })
    await router.replace('/verify-email')
  } catch (e) {
    error.value = formatAuthError(e, copy.value.failed, {
      conflict: copy.value.conflict,
      forbidden: copy.value.invalidInvite,
      registerDisabled: copy.value.disabled,
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
        <label class="field">
          <span class="field-label">{{ copy.email }}</span>
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
          <span class="field-label">{{ copy.displayName }}</span>
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
        <div class="field">
          <label class="field-label" for="register-password">{{ copy.password }}</label>
          <PasswordInput
            id="register-password"
            v-model="password"
            name="password"
            autocomplete="new-password"
            required
            :minlength="8"
            :disabled="loading"
            :aria-invalid="Boolean(error)"
            :aria-describedby="error ? 'auth-error' : 'register-password-hint'"
          />
          <span id="register-password-hint" class="field-hint">{{ copy.passwordHint }}</span>
        </div>
        <div class="field">
          <label class="field-label" for="register-confirm-password">{{ copy.confirmPassword }}</label>
          <PasswordInput
            id="register-confirm-password"
            v-model="confirmPassword"
            name="confirmPassword"
            autocomplete="new-password"
            required
            :minlength="8"
            :disabled="loading"
            :aria-invalid="Boolean(error)"
            :aria-describedby="error ? 'auth-error' : undefined"
          />
        </div>
        <label class="field">
          <span class="field-label">{{ copy.inviteCode }}</span>
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
          <span class="field-hint">{{ copy.inviteHint }}</span>
        </label>

        <div v-if="error" id="auth-error" class="auth-alert" role="alert">{{ error }}</div>

        <button class="btn ink block auth-submit" type="submit" :disabled="loading" :aria-busy="loading">
          {{ loading ? copy.creating : copy.create }}
        </button>
      </form>
    </template>

    <template #footer>
      <p class="auth-switch">
        {{ copy.existingAccount }}
        <RouterLink class="link-accent" to="/login">{{ copy.signIn }}</RouterLink>
      </p>
    </template>
  </AuthCard>
</template>
