/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { Session, User } from '@/api/types'
import { useAuthStore } from '@/stores/auth'

const { apiMock, clearTokensMock, setTokensMock } = vi.hoisted(() => ({
  apiMock: vi.fn<(path: string, options?: unknown, config?: unknown) => Promise<unknown>>(),
  clearTokensMock: vi.fn(),
  setTokensMock: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  api: apiMock,
  clearTokens: clearTokensMock,
  setTokens: setTokensMock,
}))

function makeUser(id: string, email: string): User {
  return {
    id,
    email,
    displayName: email.split('@')[0] ?? email,
    emailVerified: true,
    storageUsed: 0,
    storageQuota: 1_073_741_824,
    imageThumbnailsEnabled: true,
    videoThumbnailsEnabled: true,
    trashAutoDeleteEnabled: false,
    trashRetentionDays: 30,
    createdAt: '2026-01-01T00:00:00.000Z',
  }
}

function makeSession(id: string, email: string): Session {
  return {
    user: makeUser(id, email),
    accessToken: `access-${id}`,
    refreshToken: `refresh-${id}`,
  }
}

describe('useAuthStore session ownership', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    apiMock.mockReset()
    clearTokensMock.mockReset()
    setTokensMock.mockReset()
  })

  it('does not let a pending loadMe restore user state after logout', async () => {
    const oldUser = makeUser('old-user', 'old@example.com')
    let resolveLoadMe!: (value: unknown) => void
    apiMock.mockImplementation((path) => {
      if (path === '/users/me') {
        return new Promise((resolve) => {
          resolveLoadMe = resolve
        })
      }
      if (path === '/auth/logout') return Promise.resolve({})
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const auth = useAuthStore()
    auth.user = oldUser
    localStorage.setItem('filvault.refreshToken', 'refresh-old')

    const pendingLoad = auth.loadMe()
    await Promise.resolve()
    await auth.logout()

    expect(auth.user).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
    expect(clearTokensMock).toHaveBeenCalledTimes(1)

    resolveLoadMe(oldUser)
    await pendingLoad

    expect(auth.user).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
  })

  it('keeps the newer login authoritative when an older login resolves later', async () => {
    const older = makeSession('older', 'older@example.com')
    const newer = makeSession('newer', 'newer@example.com')
    let resolveOlder!: (value: unknown) => void
    apiMock
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveOlder = resolve
          }),
      )
      .mockResolvedValueOnce(newer)

    const auth = useAuthStore()
    const olderLogin = auth.login('older@example.com', 'password')
    await Promise.resolve()
    await auth.login('newer@example.com', 'password')

    expect(auth.user?.id).toBe('newer')
    expect(setTokensMock).toHaveBeenCalledTimes(1)
    expect(setTokensMock).toHaveBeenLastCalledWith(newer.accessToken, newer.refreshToken)

    resolveOlder(older)
    await olderLogin

    expect(auth.user?.id).toBe('newer')
    expect(setTokensMock).toHaveBeenCalledTimes(1)
  })

  it('does not let stale bootstrap overwrite a newer login and clears loading ownership', async () => {
    const bootUser = makeUser('boot-user', 'boot@example.com')
    const loginSession = makeSession('login-user', 'login@example.com')
    let resolveBootstrap!: (value: unknown) => void
    apiMock
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveBootstrap = resolve
          }),
      )
      .mockResolvedValueOnce(loginSession)

    localStorage.setItem('filvault.accessToken', 'existing-access')
    const auth = useAuthStore()
    const bootstrap = auth.bootstrap()
    await Promise.resolve()

    expect(auth.loading).toBe(true)
    await auth.login('login@example.com', 'password')

    expect(auth.user?.id).toBe('login-user')
    expect(auth.loading).toBe(false)

    resolveBootstrap(bootUser)
    await bootstrap

    expect(auth.user?.id).toBe('login-user')
    expect(auth.loading).toBe(false)
    expect(setTokensMock).toHaveBeenCalledTimes(1)
  })

  it('still clears the current local session when server logout fails', async () => {
    apiMock.mockRejectedValueOnce(new Error('network down'))
    const auth = useAuthStore()
    auth.user = makeUser('current', 'current@example.com')
    localStorage.setItem('filvault.refreshToken', 'refresh-current')

    await expect(auth.logout()).resolves.toBeUndefined()

    expect(auth.user).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
    expect(clearTokensMock).toHaveBeenCalledTimes(1)
  })

  it('ignores a stale avatar update after logout', async () => {
    const currentUser = makeUser('current', 'current@example.com')
    let resolveAvatar!: (value: unknown) => void
    apiMock.mockImplementation((path, options) => {
      if (path === '/users/me' && (options as RequestInit | undefined)?.method === 'PATCH') {
        return new Promise((resolve) => {
          resolveAvatar = resolve
        })
      }
      if (path === '/auth/logout') return Promise.resolve({})
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const auth = useAuthStore()
    auth.user = currentUser
    const pendingAvatar = auth.updateAvatar('https://example.com/avatar.png')
    await Promise.resolve()
    await auth.logout()

    resolveAvatar({ ...currentUser, avatarUrl: 'https://example.com/avatar.png' })
    await pendingAvatar

    expect(auth.user).toBeNull()
  })
})
