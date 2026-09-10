/**
 * @vitest-environment jsdom
 */
import { computed, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileView from './ProfileView.vue'

const mocks = vi.hoisted(() => ({
  api: vi.fn<(path: string, options?: RequestInit) => Promise<unknown>>(),
  setTokens: vi.fn<(accessToken: string, refreshToken: string) => void>(),
  loadMe: vi.fn<() => Promise<void>>(),
  logout: vi.fn<() => Promise<void>>(),
  updateAvatar: vi.fn<(avatarUrl: string) => Promise<void>>(),
  showToast: vi.fn<(message: string, variant?: string) => void>(),
}))

vi.mock('@/api/client', () => ({
  api: mocks.api,
  setTokens: mocks.setTokens,
  ApiError: class ApiError extends Error {},
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      id: 'user-1',
      email: 'huy@example.com',
      displayName: 'Huy',
      avatarUrl: '',
    },
    loadMe: mocks.loadMe,
    logout: mocks.logout,
    updateAvatar: mocks.updateAvatar,
  }),
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({ showToast: mocks.showToast }),
}))

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({
    locale: ref('en'),
    t: computed(() => ({
      profile: 'Profile',
      displayName: 'Display name',
      saveProfile: 'Save profile',
      saving: 'Saving…',
      password: 'Password',
      currentPassword: 'Current password',
      newPassword: 'New password',
      changePassword: 'Change password',
      changing: 'Changing…',
      logOut: 'Log out',
    })),
  }),
}))

vi.mock('@/lib/heic', () => ({
  isHeic: () => false,
  convertHeicBlobToJpeg: vi.fn<(blob: Blob, quality?: number) => Promise<Blob>>(),
  checkIsHeicBlob: vi.fn<(blob: Blob) => Promise<boolean>>().mockResolvedValue(false),
}))

function mountProfile() {
  return mount(ProfileView, {
    global: {
      stubs: {
        Icon: { template: '<span />' },
      },
    },
  })
}

async function fillPasswords(
  wrapper: ReturnType<typeof mountProfile>,
  current = 'current-password',
  next = 'new-password-123',
  confirm = next,
) {
  await wrapper.get('#profile-current-password').setValue(current)
  await wrapper.get('#profile-new-password').setValue(next)
  await wrapper.get('#profile-confirm-password').setValue(confirm)
}

function changeButton(wrapper: ReturnType<typeof mountProfile>) {
  return wrapper.findAll('button').find((button) => button.text().includes('Change password'))!
}

describe('Profile password change', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.loadMe.mockResolvedValue(undefined)
  })

  it('blocks a mismatched confirmation before making the password request', async () => {
    const wrapper = mountProfile()
    await fillPasswords(wrapper, 'current-password', 'new-password-123', 'different-password')

    await changeButton(wrapper).trigger('click')

    expect(mocks.api).not.toHaveBeenCalled()
    expect(wrapper.get('#profile-password-error').text()).toContain('does not match')
  })

  it('preserves all entered values after a recoverable API failure', async () => {
    mocks.api.mockRejectedValueOnce(new Error('network'))
    const wrapper = mountProfile()
    await fillPasswords(wrapper)

    await changeButton(wrapper).trigger('click')
    await flushPromises()

    expect((wrapper.get('#profile-current-password').element as HTMLInputElement).value).toBe('current-password')
    expect((wrapper.get('#profile-new-password').element as HTMLInputElement).value).toBe('new-password-123')
    expect((wrapper.get('#profile-confirm-password').element as HTMLInputElement).value).toBe('new-password-123')
    expect(wrapper.get('#profile-password-error').text()).toBe('Password change failed')
  })

  it('clears password fields only after success and refreshes tokens/account state', async () => {
    mocks.api.mockResolvedValueOnce({ accessToken: 'access-2', refreshToken: 'refresh-2' })
    const wrapper = mountProfile()
    await fillPasswords(wrapper)

    await changeButton(wrapper).trigger('click')
    await flushPromises()

    expect(mocks.api).toHaveBeenCalledWith('/users/me/password', {
      method: 'POST',
      body: JSON.stringify({
        currentPassword: 'current-password',
        newPassword: 'new-password-123',
      }),
    })
    expect(mocks.setTokens).toHaveBeenCalledWith('access-2', 'refresh-2')
    expect(mocks.loadMe).toHaveBeenCalledTimes(1)
    expect((wrapper.get('#profile-current-password').element as HTMLInputElement).value).toBe('')
    expect((wrapper.get('#profile-new-password').element as HTMLInputElement).value).toBe('')
    expect((wrapper.get('#profile-confirm-password').element as HTMLInputElement).value).toBe('')
  })

  it('prevents duplicate password requests while one change is pending', async () => {
    let resolveRequest!: (value: { accessToken: string; refreshToken: string }) => void
    mocks.api.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveRequest = resolve
        }),
    )
    const wrapper = mountProfile()
    await fillPasswords(wrapper)

    const button = changeButton(wrapper)
    await button.trigger('click')
    await button.trigger('click')

    expect(mocks.api).toHaveBeenCalledTimes(1)
    expect(button.attributes('disabled')).toBeDefined()
    expect(button.attributes('aria-busy')).toBe('true')

    resolveRequest({ accessToken: 'access-2', refreshToken: 'refresh-2' })
    await flushPromises()
  })
})
