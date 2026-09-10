/**
 * @vitest-environment jsdom
 */
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import ShareUserSheet from './ShareUserSheet.vue'
import { useAuthStore } from '@/stores/auth'
import { setLocale } from '@/lib/i18n'

const apiMock = vi.fn<(path: string, options?: RequestInit) => Promise<unknown>>()

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    api: (path: string, options?: RequestInit) => apiMock(path, options),
  }
})

function mountSheet(overrides: { open?: boolean; resourceId?: string | null } = {}) {
  return mount(ShareUserSheet, {
    props: {
      open: overrides.open ?? true,
      name: 'Budget.xlsx',
      resourceId: overrides.resourceId ?? '01KFILE0000000000000000001',
    },
    attachTo: document.body,
    global: {
      stubs: { Teleport: true },
    },
  })
}

async function fillEmail(wrapper: ReturnType<typeof mountSheet>, value: string) {
  await wrapper.get('input[type="email"]').setValue(value)
}

async function submitShare(wrapper: ReturnType<typeof mountSheet>) {
  await wrapper.get('button[type="submit"]').trigger('click')
  await flushPromises()
}

describe('ShareUserSheet', () => {
  beforeEach(() => {
    setLocale('en')
    setActivePinia(createPinia())
    const auth = useAuthStore()
    auth.user = {
      id: '01KUSER00000000000000000001',
      email: 'owner@example.com',
      displayName: 'Owner',
      emailVerified: true,
      storageUsed: 0,
      storageQuota: 1_073_741_824,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-01-01T00:00:00.000Z',
    }
    apiMock.mockReset()
  })

  afterEach(() => {
    setLocale('vi')
    document.body.innerHTML = ''
  })

  it('states read-access permission before confirmation', () => {
    const wrapper = mountSheet()
    expect(wrapper.text()).toContain("They'll get read access to this item")
    expect(wrapper.text()).toContain('Unknown emails receive an invite to sign up')
    wrapper.unmount()
  })

  it('reacts to the active locale for permission and actions', async () => {
    const wrapper = mountSheet()
    setLocale('vi')
    await flushPromises()
    expect(wrapper.text()).toContain('Người nhận sẽ có quyền đọc mục này')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Chia sẻ')
    wrapper.unmount()
  })

  it('blocks invalid email locally without calling the API', async () => {
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'user@domain')
    await submitShare(wrapper)
    expect(apiMock).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toContain('Enter a valid email address')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Share')
    expect(wrapper.get('input[type="email"]').element.value).toBe('user@domain')
    wrapper.unmount()
  })

  it('blocks self-share locally and keeps the email for correction', async () => {
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'owner@example.com')
    await submitShare(wrapper)
    expect(apiMock).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toContain("You can't share with yourself")
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('shows success for an existing user and prevents duplicate submission', async () => {
    apiMock.mockResolvedValueOnce({
      id: '01KSHARE000000000000000001',
      recipient: { id: '01KUSER2', email: 'peer@example.com', displayName: 'Peer' },
    })
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'peer@example.com')
    await submitShare(wrapper)
    expect(wrapper.text()).toContain('Shared with Peer')
    expect(wrapper.text()).toContain('read access')
    expect(wrapper.find('button[type="submit"]').exists()).toBe(false)
    await wrapper.get('.btn.ink').trigger('click')
    expect(wrapper.emitted('close')).toEqual([[]])
    wrapper.unmount()
  })

  it('shows invite success when the recipient is not registered yet', async () => {
    apiMock.mockResolvedValueOnce({ invited: true })
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'new@example.com')
    await submitShare(wrapper)
    expect(wrapper.text()).toContain('Invitation sent to new@example.com')
    expect(wrapper.text()).toContain('sign up')
    wrapper.unmount()
  })

  it('recovers from API failure with email preserved and CTA re-enabled', async () => {
    apiMock.mockRejectedValueOnce(new ApiError('CONFLICT', 'Conflict', 409))
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'peer@example.com')
    await submitShare(wrapper)
    expect(wrapper.get('[role="alert"]').text()).toContain('already shared')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Share')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.get('input[type="email"]').element.value).toBe('peer@example.com')
    wrapper.unmount()
  })

  it('localizes server failures in Vietnamese', async () => {
    apiMock.mockRejectedValueOnce(new ApiError('CONFLICT', 'Conflict', 409))
    const wrapper = mountSheet()
    setLocale('vi')
    await fillEmail(wrapper, 'peer@example.com')
    await submitShare(wrapper)
    expect(wrapper.get('[role="alert"]').text()).toContain('đã được chia sẻ')
    expect(wrapper.get('input[type="email"]').element.value).toBe('peer@example.com')
    wrapper.unmount()
  })

  it('allows retry after failure and share another after success', async () => {
    apiMock
      .mockRejectedValueOnce(new ApiError('VALIDATION_ERROR', 'Invalid request', 400))
      .mockResolvedValueOnce({ invited: true })
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'retry@example.com')
    await submitShare(wrapper)
    expect(wrapper.get('[role="alert"]').exists()).toBe(true)
    await submitShare(wrapper)
    expect(wrapper.text()).toContain('Invitation sent to retry@example.com')
    await wrapper.get('.btn.ghost').trigger('click')
    await flushPromises()
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
    expect(wrapper.get('input[type="email"]').element.value).toBe('')
    wrapper.unmount()
  })

  it('does not leave the submit button stuck while the request is pending', async () => {
    let resolveRequest!: (value: unknown) => void
    apiMock.mockImplementationOnce(() => new Promise((resolve) => { resolveRequest = resolve }))
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'pending@example.com')
    const promise = wrapper.get('button[type="submit"]').trigger('click')
    await Promise.resolve()
    const submit = wrapper.get('button[type="submit"]')
    expect(submit.text()).toBe('Sharing…')
    expect(submit.attributes('aria-busy')).toBe('true')
    resolveRequest({ invited: true })
    await promise
    await flushPromises()
    expect(wrapper.find('button[type="submit"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('moves keyboard focus to the error alert on failure', async () => {
    apiMock.mockRejectedValueOnce(new ApiError('NOT_FOUND', 'Not found', 404))
    const wrapper = mountSheet()
    await fillEmail(wrapper, 'missing@example.com')
    await submitShare(wrapper)
    expect(document.activeElement).toBe(wrapper.get('[role="alert"]').element)
    wrapper.unmount()
  })
})
