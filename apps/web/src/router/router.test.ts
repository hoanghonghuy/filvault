/**
 * @vitest-environment jsdom
 */
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import router from './index'
import { useAuthStore } from '@/stores/auth'

describe('router navigation guards for auth and verification', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(async () => {
    pinia = createPinia()
    setActivePinia(pinia)
    localStorage.clear()
    await router.push('/')
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('redirects verified authenticated user away from /verify-email to /', async () => {
    const auth = useAuthStore(pinia)
    auth.user = {
      id: 'user-1',
      email: 'verified@example.com',
      displayName: 'Verified User',
      emailVerified: true,
      storageUsed: 0,
      storageQuota: 1024,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-08-28T12:00:00.000Z',
    }

    await router.push('/verify-email')
    expect(router.currentRoute.value.path).toBe('/')
  })

  it('redirects verified authenticated user away from /login and /register to /', async () => {
    const auth = useAuthStore(pinia)
    auth.user = {
      id: 'user-1',
      email: 'verified@example.com',
      displayName: 'Verified User',
      emailVerified: true,
      storageUsed: 0,
      storageQuota: 1024,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-08-28T12:00:00.000Z',
    }

    await router.push('/login')
    expect(router.currentRoute.value.path).toBe('/')

    await router.push('/register')
    expect(router.currentRoute.value.path).toBe('/')
  })

  it('allows unverified authenticated user to access /verify-email', async () => {
    const auth = useAuthStore(pinia)
    auth.user = {
      id: 'user-2',
      email: 'unverified@example.com',
      displayName: 'Unverified User',
      emailVerified: false,
      storageUsed: 0,
      storageQuota: 1024,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-08-28T12:00:00.000Z',
    }

    await router.push('/verify-email')
    expect(router.currentRoute.value.path).toBe('/verify-email')
  })

  it('redirects unverified authenticated user away from protected routes to /verify-email', async () => {
    const auth = useAuthStore(pinia)
    auth.user = {
      id: 'user-2',
      email: 'unverified@example.com',
      displayName: 'Unverified User',
      emailVerified: false,
      storageUsed: 0,
      storageQuota: 1024,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-08-28T12:00:00.000Z',
    }

    await router.push('/chat')
    expect(router.currentRoute.value.path).toBe('/verify-email')

    await router.push('/files')
    expect(router.currentRoute.value.path).toBe('/verify-email')
  })

  it('redirects unauthenticated guest away from /verify-email to /login', async () => {
    const auth = useAuthStore(pinia)
    auth.user = null

    await router.push('/verify-email')
    expect(router.currentRoute.value.path).toBe('/login')
  })
})
