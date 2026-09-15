/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { setLocale } from '@/lib/i18n'
import VerifyEmailView from './VerifyEmailView.vue'

const authMock = vi.hoisted(() => ({
  isAuthenticated: true,
  isVerified: false,
  user: { email: 'wrong@example.com' } as { email: string } | null,
  logout: vi.fn<() => Promise<void>>(),
  resendVerification: vi.fn<(email: string) => Promise<void>>(),
  verifyEmail: vi.fn<(email: string, code: string) => Promise<void>>(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authMock,
}))

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/verify-email', component: VerifyEmailView },
      { path: '/register', component: { template: '<div>Register</div>' } },
      { path: '/', component: { template: '<div>Home</div>' } },
    ],
  })
}

async function mountVerifyEmail() {
  const router = makeRouter()
  await router.push('/verify-email')
  await router.isReady()
  return { router, wrapper: mount(VerifyEmailView, { global: { plugins: [router] } }) }
}

describe('VerifyEmailView wrong-email recovery', () => {
  beforeEach(() => {
    setLocale('en')
    authMock.isAuthenticated = true
    authMock.isVerified = false
    authMock.user = { email: 'wrong@example.com' }
    authMock.logout.mockReset().mockResolvedValue(undefined)
    authMock.resendVerification.mockReset().mockResolvedValue(undefined)
    authMock.verifyEmail.mockReset().mockResolvedValue(undefined)
  })

  it('reacts to Vietnamese locale without remounting', async () => {
    const { wrapper } = await mountVerifyEmail()

    expect(wrapper.get('h1').text()).toBe('Verify email')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Verify email')
    expect(wrapper.text()).toContain('Wrong email?')

    setLocale('vi')
    await wrapper.vm.$nextTick()

    expect(wrapper.get('h1').text()).toBe('Xác minh email')
    expect(wrapper.text()).toContain('Nhập mã 6 chữ số đã được gửi tới wrong@example.com.')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Xác minh email')
    expect(wrapper.text()).toContain('Sai email?')
    expect(wrapper.get('button.btn.ghost').text()).toBe('Dùng email khác')
    wrapper.unmount()
  })

  it('signs out the unverified session and replaces to registration', async () => {
    const { router, wrapper } = await mountVerifyEmail()

    const button = wrapper.get('button.btn.ghost')
    expect(button.text()).toBe('Use a different email')
    expect(wrapper.text()).toContain('sign out of this unverified session and register again')

    await button.trigger('click')
    await vi.waitFor(() => expect(authMock.logout).toHaveBeenCalledOnce())
    await vi.waitFor(() => expect(router.currentRoute.value.path).toBe('/register'))

    expect(authMock.resendVerification).not.toHaveBeenCalled()
    expect(authMock.verifyEmail).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('prevents duplicate switching while logout is still in progress', async () => {
    let resolveLogout: (() => void) | undefined
    authMock.logout.mockImplementation(
      () => new Promise<void>((resolve) => {
        resolveLogout = resolve
      }),
    )

    const { router, wrapper } = await mountVerifyEmail()

    const button = wrapper.get('button.btn.ghost')
    await button.trigger('click')

    expect(authMock.logout).toHaveBeenCalledOnce()
    expect(button.attributes('disabled')).toBeDefined()
    expect(button.attributes('aria-busy')).toBe('true')
    expect(button.text()).toBe('Switching…')

    await button.trigger('click')
    expect(authMock.logout).toHaveBeenCalledOnce()

    resolveLogout?.()
    await vi.waitFor(() => expect(router.currentRoute.value.path).toBe('/register'))
    wrapper.unmount()
  })
})
