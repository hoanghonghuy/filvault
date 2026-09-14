import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { api, clearTokens, setTokens } from '@/api/client'
import type { Session, User } from '@/api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)
  let sessionGeneration = 0

  const isAuthenticated = computed(() => user.value !== null)
  const isVerified = computed(() => user.value?.emailVerified ?? false)

  function beginSessionTransition(): number {
    sessionGeneration += 1
    loading.value = false
    return sessionGeneration
  }

  function isCurrentSession(generation: number): boolean {
    return generation === sessionGeneration
  }

  async function loadMe(): Promise<void> {
    const generation = sessionGeneration
    const nextUser = await api<User>('/users/me')
    if (isCurrentSession(generation)) user.value = nextUser
  }

  async function register(payload: {
    email: string
    password: string
    displayName: string
    inviteCode: string
  }): Promise<void> {
    const generation = beginSessionTransition()
    const session = await api<Session>(
      '/auth/register',
      { method: 'POST', body: JSON.stringify(payload) },
      { auth: false },
    )
    if (!isCurrentSession(generation)) return
    setTokens(session.accessToken, session.refreshToken)
    user.value = session.user
  }

  async function login(email: string, password: string): Promise<void> {
    const generation = beginSessionTransition()
    const session = await api<Session>(
      '/auth/login',
      { method: 'POST', body: JSON.stringify({ email, password }) },
      { auth: false },
    )
    if (!isCurrentSession(generation)) return
    setTokens(session.accessToken, session.refreshToken)
    user.value = session.user
  }

  async function logout(): Promise<void> {
    const generation = beginSessionTransition()
    const hadUser = Boolean(user.value)
    const refreshToken = localStorage.getItem('filvault.refreshToken')
    user.value = null

    if (hadUser) {
      try {
        await api('/auth/logout', {
          method: 'POST',
          body: JSON.stringify({ refreshToken }),
        })
      } catch {
        // ignore logout errors locally
      }
    }

    if (!isCurrentSession(generation)) return
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
    const generation = sessionGeneration
    await api(
      '/auth/verify-email',
      { method: 'POST', body: JSON.stringify({ email, code }) },
      { auth: false },
    )
    if (!isCurrentSession(generation)) return
    if (user.value && user.value.email === email) {
      user.value = { ...user.value, emailVerified: true }
    }
    await loadMe()
  }

  async function bootstrap(): Promise<void> {
    if (!localStorage.getItem('filvault.accessToken')) {
      return
    }
    const generation = beginSessionTransition()
    loading.value = true
    try {
      const nextUser = await api<User>('/users/me')
      if (isCurrentSession(generation)) user.value = nextUser
    } catch {
      if (isCurrentSession(generation)) {
        clearTokens()
        user.value = null
      }
    } finally {
      if (isCurrentSession(generation)) loading.value = false
    }
  }

  async function updateActiveStatus(enabled: boolean): Promise<void> {
    const generation = sessionGeneration
    const updated = await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({ activeStatusEnabled: enabled }),
    })
    if (isCurrentSession(generation)) user.value = updated
  }

  async function updateAvatar(avatarUrl: string): Promise<void> {
    const generation = sessionGeneration
    const updated = await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({ avatarUrl }),
    })
    if (isCurrentSession(generation)) user.value = updated
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
    updateActiveStatus,
    updateAvatar,
    bootstrap,
  }
})
