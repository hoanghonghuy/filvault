import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { api, clearTokens, setTokens } from '@/api/client'
import type { Session, User } from '@/api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)

  const isAuthenticated = computed(() => user.value !== null)
  const isVerified = computed(() => user.value?.emailVerified ?? false)

  async function loadMe(): Promise<void> {
    user.value = await api<User>('/users/me')
  }

  async function register(payload: {
    email: string
    password: string
    displayName: string
    inviteCode: string
  }): Promise<void> {
    const session = await api<Session>(
      '/auth/register',
      { method: 'POST', body: JSON.stringify(payload) },
      { auth: false },
    )
    setTokens(session.accessToken, session.refreshToken)
    user.value = session.user
  }

  async function login(email: string, password: string): Promise<void> {
    const session = await api<Session>(
      '/auth/login',
      { method: 'POST', body: JSON.stringify({ email, password }) },
      { auth: false },
    )
    setTokens(session.accessToken, session.refreshToken)
    user.value = session.user
  }

  async function logout(): Promise<void> {
    if (user.value) {
      try {
        await api('/auth/logout', {
          method: 'POST',
          body: JSON.stringify({ refreshToken: localStorage.getItem('filvault.refreshToken') }),
        })
      } catch {
        // ignore logout errors locally
      }
    }
    clearTokens()
    user.value = null
  }

  async function resendVerification(email: string): Promise<void> {
    await api(
      '/auth/resend-verification',
      { method: 'POST', body: JSON.stringify({ email }) },
      { auth: false },
    )
  }

  async function verifyEmail(email: string, code: string): Promise<void> {
    await api(
      '/auth/verify-email',
      { method: 'POST', body: JSON.stringify({ email, code }) },
      { auth: false },
    )
    if (user.value && user.value.email === email) {
      user.value = { ...user.value, emailVerified: true }
    }
    await loadMe()
  }

  async function bootstrap(): Promise<void> {
    if (!localStorage.getItem('filvault.accessToken')) {
      return
    }
    loading.value = true
    try {
      await loadMe()
    } catch {
      clearTokens()
      user.value = null
    } finally {
      loading.value = false
    }
  }

  return {
    user,
    loading,
    isAuthenticated,
    isVerified,
    loadMe,
    register,
    login,
    logout,
    resendVerification,
    verifyEmail,
    bootstrap,
  }
})
