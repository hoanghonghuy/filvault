/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import LoginView from './LoginView.vue'

const authMock = vi.hoisted(() => ({
  isAuthenticated: false,
  isVerified: false,
  login: vi.fn<(email: string, password: string) => Promise<void>>(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authMock,
}))

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', name: 'login', component: LoginView },
      { path: '/forgot-password', name: 'forgot-password', component: { template: '<div>Recovery</div>' } },
      { path: '/register', name: 'register', component: { template: '<div>Register</div>' } },
      { path: '/verify-email', name: 'verify-email', component: { template: '<div>Verify</div>' } },
      { path: '/', name: 'home', component: { template: '<div>Home</div>' } },
    ],
  })
}

describe('LoginView password recovery', () => {
  beforeEach(() => {
    authMock.isAuthenticated = false
    authMock.isVerified = false
    authMock.login.mockReset()
    authMock.login.mockResolvedValue(undefined)
  })

  it('exposes a keyboard-accessible forgot-password route beside the password field', async () => {
    const router = makeRouter()
    await router.push('/login')
    await router.isReady()
    const wrapper = mount(LoginView, { global: { plugins: [router] } })

    const link = wrapper.get('a.forgot-link')
    expect(link.text()).toBe('Forgot password?')
    expect(link.attributes('href')).toBe('/forgot-password')

    wrapper.unmount()
  })

  it('shows reset completion feedback when returning from recovery', async () => {
    const router = makeRouter()
    await router.push('/login?reset=success')
    await router.isReady()
    const wrapper = mount(LoginView, { global: { plugins: [router] } })

    expect(wrapper.get('[role="status"]').text()).toContain('Password reset complete')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Sign in')

    wrapper.unmount()
  })
})
