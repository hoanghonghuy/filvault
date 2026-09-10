/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import RegisterView from './RegisterView.vue'

const authMock = vi.hoisted(() => ({
  isAuthenticated: false,
  isVerified: false,
  register: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authMock,
}))

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/register', component: RegisterView },
      { path: '/verify-email', component: { template: '<div>Verify</div>' } },
      { path: '/', component: { template: '<div>Home</div>' } },
      { path: '/login', component: { template: '<div>Login</div>' } },
    ],
  })
}

async function fillRegistration(wrapper: ReturnType<typeof mount>, confirmation: string) {
  await wrapper.get<HTMLInputElement>('#register-email').setValue('  user@example.com  ')
  await wrapper.get<HTMLInputElement>('#register-display-name').setValue('  User Name  ')
  await wrapper.get<HTMLInputElement>('#register-password').setValue('password-123')
  await wrapper.get<HTMLInputElement>('#register-confirm-password').setValue(confirmation)
  await wrapper.get<HTMLInputElement>('#register-invite').setValue('  invite-code  ')
}

describe('RegisterView password confirmation', () => {
  beforeEach(() => {
    authMock.isAuthenticated = false
    authMock.isVerified = false
    authMock.register.mockReset()
    authMock.register.mockResolvedValue(undefined)
  })

  it('blocks registration locally when passwords do not match and preserves input', async () => {
    const router = makeRouter()
    await router.push('/register')
    await router.isReady()
    const wrapper = mount(RegisterView, { global: { plugins: [router] } })

    await fillRegistration(wrapper, 'different-password')
    await wrapper.get('form').trigger('submit')

    expect(authMock.register).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toContain('Passwords do not match')
    expect(wrapper.get<HTMLInputElement>('#register-email').element.value).toBe('  user@example.com  ')
    expect(wrapper.get<HTMLInputElement>('#register-password').element.value).toBe('password-123')
    expect(wrapper.get<HTMLInputElement>('#register-confirm-password').element.value).toBe('different-password')

    wrapper.unmount()
  })

  it('submits the existing normalized payload and redirects when passwords match', async () => {
    const router = makeRouter()
    await router.push('/register')
    await router.isReady()
    const wrapper = mount(RegisterView, { global: { plugins: [router] } })

    await fillRegistration(wrapper, 'password-123')
    await wrapper.get('form').trigger('submit')
    await vi.waitFor(() => expect(authMock.register).toHaveBeenCalledOnce())

    expect(authMock.register).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'password-123',
      displayName: 'User Name',
      inviteCode: 'invite-code',
    })
    await vi.waitFor(() => expect(router.currentRoute.value.path).toBe('/verify-email'))

    wrapper.unmount()
  })

  it('uses new-password autofill semantics for password and confirmation', async () => {
    const router = makeRouter()
    await router.push('/register')
    await router.isReady()
    const wrapper = mount(RegisterView, { global: { plugins: [router] } })

    expect(wrapper.get('#register-password').attributes('autocomplete')).toBe('new-password')
    expect(wrapper.get('#register-confirm-password').attributes('autocomplete')).toBe('new-password')

    wrapper.unmount()
  })
})
