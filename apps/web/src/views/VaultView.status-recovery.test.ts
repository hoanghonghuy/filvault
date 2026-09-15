/**
 * @vitest-environment jsdom
 */
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/lib/i18n'
import VaultView from './VaultView.vue'

const fetchStatusMock = vi.hoisted(() =>
  vi.fn<() => Promise<{ initialized: boolean; unlocked: boolean }>>(),
)
const loadFilesMock = vi.hoisted(() => vi.fn<() => Promise<unknown>>())
const routerBackMock = vi.hoisted(() => vi.fn<() => void>())
const routerPushMock = vi.hoisted(() => vi.fn<(path: string) => void>())

vi.mock('vue-router', () => ({
  useRouter: () => ({
    back: routerBackMock,
    push: routerPushMock,
  }),
}))

vi.mock('@/stores/vault', () => ({
  useVaultStore: () => ({
    status: null,
    files: [],
    loading: false,
    error: '',
    isInitialized: true,
    isUnlocked: false,
    fetchStatus: fetchStatusMock,
    loadFiles: loadFilesMock,
    setup: vi.fn<(pin: string) => Promise<unknown>>(),
    unlock: vi.fn<(pin: string) => Promise<unknown>>(),
    lock: vi.fn<() => void>(),
    changePin: vi.fn<(currentPin: string, newPin: string) => Promise<void>>(),
    resetPin: vi.fn<(accountPassword: string, newPin: string) => Promise<unknown>>(),
    removeFromVault: vi.fn<(fileIds: string[]) => Promise<void>>(),
  }),
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({
    showToast: vi.fn<(message: string, type?: string) => void>(),
    confirm: vi.fn<() => Promise<boolean>>().mockResolvedValue(true),
  }),
}))

function mountVault() {
  return mount(VaultView, {
    global: {
      stubs: {
        AppIcon: true,
        BottomSheet: true,
        MediaLightbox: true,
      },
    },
  })
}

describe('VaultView status recovery', () => {
  beforeEach(() => {
    fetchStatusMock.mockReset()
    loadFilesMock.mockReset()
    routerBackMock.mockReset()
    routerPushMock.mockReset()
    setLocale('en')
  })

  it('never presents first-time setup while status is pending or failed, then recovers to the authoritative locked state', async () => {
    let rejectInitial: ((reason?: unknown) => void) | undefined
    const initialRequest = new Promise<{ initialized: boolean; unlocked: boolean }>((_resolve, reject) => {
      rejectInitial = reject
    })

    fetchStatusMock
      .mockImplementationOnce(() => initialRequest)
      .mockResolvedValueOnce({ initialized: true, unlocked: false })

    const wrapper = mountVault()
    await nextTick()

    expect(wrapper.text()).toContain('Loading Personal Vault…')
    expect(wrapper.text()).not.toContain('Set Up Personal Vault Password')

    expect(rejectInitial).toBeTypeOf('function')
    rejectInitial?.(new Error('status offline'))
    await flushPromises()

    const alert = wrapper.get('[role="alert"]')
    expect(alert.text()).toContain('Could not load Personal Vault status. Please try again.')
    expect(wrapper.text()).not.toContain('Set Up Personal Vault Password')
    expect(wrapper.text()).not.toContain('Personal Vault is Locked')

    await alert.get('button').trigger('click')
    await flushPromises()

    expect(fetchStatusMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Personal Vault is Locked')
    expect(loadFilesMock).not.toHaveBeenCalled()

    wrapper.unmount()
  })

  it('continues into the existing file load path when authoritative status is unlocked', async () => {
    fetchStatusMock.mockResolvedValueOnce({ initialized: true, unlocked: true })
    loadFilesMock.mockResolvedValueOnce([])

    const wrapper = mountVault()
    await flushPromises()

    expect(fetchStatusMock).toHaveBeenCalledTimes(1)
    expect(loadFilesMock).toHaveBeenCalledTimes(1)

    wrapper.unmount()
  })
})
