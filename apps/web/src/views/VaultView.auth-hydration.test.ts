/**
 * @vitest-environment jsdom
 */
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/lib/i18n'
import VaultView from './VaultView.vue'

const loadFilesMock = vi.hoisted(() => vi.fn<() => Promise<unknown>>())
const fetchStatusMock = vi.hoisted(() => vi.fn<() => Promise<unknown>>())
const unlockMock = vi.hoisted(() => vi.fn<(pin: string) => Promise<unknown>>())
const showToastMock = vi.hoisted(() => vi.fn<(message: string, type?: string) => void>())
const routerBackMock = vi.hoisted(() => vi.fn<() => void>())
const routerPushMock = vi.hoisted(() => vi.fn<(path: string) => void>())
const vaultStore = vi.hoisted(() => ({
  status: { initialized: true, unlocked: false },
  files: [] as unknown[],
  loading: false,
  error: '',
  isInitialized: true,
  isUnlocked: false,
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ back: routerBackMock, push: routerPushMock }),
}))

vi.mock('@/stores/vault', () => ({
  useVaultStore: () => ({
    ...vaultStore,
    fetchStatus: fetchStatusMock,
    loadFiles: loadFilesMock,
    setup: vi.fn<(pin: string) => Promise<unknown>>(),
    unlock: unlockMock,
    lock: vi.fn<() => void>(),
    changePin: vi.fn<(currentPin: string, newPin: string) => Promise<void>>(),
    resetPin: vi.fn<(accountPassword: string, newPin: string) => Promise<unknown>>(),
    removeFromVault: vi.fn<(fileIds: string[]) => Promise<void>>(),
  }),
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({
    showToast: showToastMock,
    confirm: vi.fn<() => Promise<boolean>>().mockResolvedValue(true),
  }),
}))

describe('VaultView auth hydration separation', () => {
  beforeEach(() => {
    vaultStore.status.initialized = true
    vaultStore.status.unlocked = false
    vaultStore.isInitialized = true
    vaultStore.isUnlocked = false
    vaultStore.files = []
    loadFilesMock.mockReset()
    fetchStatusMock.mockReset().mockResolvedValue({ initialized: true, unlocked: false })
    unlockMock.mockReset().mockImplementation(async () => {
      vaultStore.status.unlocked = true
      vaultStore.isUnlocked = true
      return { token: 'vault-token' }
    })
    showToastMock.mockReset()
    routerBackMock.mockReset()
    routerPushMock.mockReset()
    setLocale('en')
  })

  it('keeps unlock successful when the subsequent file hydration fails', async () => {
    loadFilesMock.mockRejectedValue(new Error('vault files offline'))

    const wrapper = mount(VaultView, {
      global: {
        stubs: {
          AppIcon: true,
          BottomSheet: true,
          MediaLightbox: true,
        },
      },
    })
    await flushPromises()

    const input = wrapper.get('input[placeholder="Enter your PIN"]')
    await input.setValue('1234')
    await wrapper.get('form.vault-form').trigger('submit')
    await flushPromises()

    expect(unlockMock).toHaveBeenCalledWith('1234')
    expect(loadFilesMock).toHaveBeenCalledTimes(1)
    expect(showToastMock).toHaveBeenCalledWith(expect.any(String), 'success')
    expect(wrapper.text()).not.toContain('Failed to unlock Vault')
    expect(wrapper.get('[role="alert"]').text()).toContain('Failed to load files')
    expect(wrapper.text()).not.toContain('Personal Vault is empty')

    wrapper.unmount()
  })
})
