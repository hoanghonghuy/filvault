/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { ApiError } from '@/api/client'
import ForgotPasswordView from './ForgotPasswordView.vue'

const apiMock = vi.hoisted(() => vi.fn<(path: string, init?: RequestInit, options?: { auth?: boolean }) => Promise<void>>())

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return { ...actual, api: apiMock }
})

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/forgot-password', name: 'forgot-password', component: ForgotPasswordView },
      { path: '/login', name: 'login', component: { template: '<div>Login</div>' } },
    ],
  })
}

async function mountRecovery() {
  const router = makeRouter()
  await router.push('/forgot-password')
  await router.isReady()
  const wrapper = mount(ForgotPasswordView, { global: { plugins: [router] } })
  return { wrapper, router }
}

async function enterResetPhase(wrapper: ReturnType<typeof mount>) {
  await wrapper.get<HTMLInputElement>('#recovery-email').setValue(' user@example.com ')
  await wrapper.get('form').trigger('submit')
  await flushPromises()
}

describe('ForgotPasswordView', () => {
  beforeEach(() => {
    apiMock.mockReset()
    apiMock.mockResolvedValue(undefined)
  })

  it('shows the same non-enumerating success presentation after a reset request', async () => {
    const { wrapper } = await mountRecovery()

    await enterResetPhase(wrapper)

    expect(apiMock).toHaveBeenCalledWith(
      '/auth/password/forgot',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ email: 'user@example.com' }) }),
      { auth: false },
    )
    expect(wrapper.get('[role="status"]').text()).toContain('If an account exists for user@example.com')
    expect(wrapper.get('#recovery-token').exists()).toBe(true)
    wrapper.unmount()
  })

  it('clears a request-stage network error when entering an existing reset token', async () => {
    apiMock.mockRejectedValueOnce(new Error('offline'))
    const { wrapper } = await mountRecovery()

    await wrapper.get<HTMLInputElement>('#recovery-email').setValue('user@example.com')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('Could not request a password reset')

    await wrapper.get('button.token-ready-btn').trigger('click')

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('#recovery-token').exists()).toBe(true)
    wrapper.unmount()
  })

  it('validates password mismatch before calling reset API', async () => {
    const { wrapper } = await mountRecovery()
    await enterResetPhase(wrapper)
    apiMock.mockClear()

    await wrapper.get<HTMLInputElement>('#recovery-token').setValue('token-from-email')
    await wrapper.get<HTMLInputElement>('#recovery-password').setValue('password-123')
    await wrapper.get<HTMLInputElement>('#recovery-confirm-password').setValue('different-123')
    await wrapper.get('form').trigger('submit')

    expect(apiMock).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toContain('Passwords do not match')
    wrapper.unmount()
  })

  it('submits one reset request while pending and returns to Login with success feedback state', async () => {
    let resolveReset!: () => void
    const pending = new Promise<void>((resolve) => {
      resolveReset = resolve
    })
    apiMock.mockReturnValueOnce(pending)

    const { wrapper, router } = await mountRecovery()
    await wrapper.get('button.token-ready-btn').trigger('click')

    await wrapper.get<HTMLInputElement>('#recovery-token').setValue('valid-reset-token')
    await wrapper.get<HTMLInputElement>('#recovery-password').setValue('password-123')
    await wrapper.get<HTMLInputElement>('#recovery-confirm-password').setValue('password-123')
    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')

    expect(apiMock).toHaveBeenCalledTimes(1)
    expect(apiMock).toHaveBeenCalledWith(
      '/auth/password/reset',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ token: 'valid-reset-token', newPassword: 'password-123' }),
      }),
      { auth: false },
    )

    resolveReset()
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/login?reset=success')
    wrapper.unmount()
  })

  it('offers a request-again path after an invalid or expired token', async () => {
    apiMock.mockRejectedValueOnce(new ApiError('VALIDATION_ERROR', 'invalid', 400))
    const { wrapper } = await mountRecovery()
    await wrapper.get('button.token-ready-btn').trigger('click')

    await wrapper.get<HTMLInputElement>('#recovery-token').setValue('expired-token')
    await wrapper.get<HTMLInputElement>('#recovery-password').setValue('password-123')
    await wrapper.get<HTMLInputElement>('#recovery-confirm-password').setValue('password-123')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toContain('invalid or expired')
    expect(wrapper.findAll('button').some((button) => button.text().includes('Request a new reset token'))).toBe(true)
    wrapper.unmount()
  })
})
