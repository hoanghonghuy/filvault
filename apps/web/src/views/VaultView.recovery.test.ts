/**
 * @vitest-environment jsdom
 */
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/lib/i18n'
import VaultView from './VaultView.vue'

const loadFilesMock = vi.hoisted(() => vi.fn<() => Promise<unknown>>())
const fetchStatusMock = vi.hoisted(() => vi.fn<() => Promise<unknown>>())
const routerBackMock = vi.hoisted(() => vi.fn())
const routerPushMock = vi.hoisted(() => vi.fn())

vi.mock('vue-router', () => ({
  useRouter: () => ({
    back: routerBackMock,
    push: routerPushMock,
  }),
}))

vi.mock('@/stores/vault', () => ({
  useVaultStore: () => ({
    status: { initialized: true, unlocked: true },
    files: [],
    loading: false,
    error: '',
    isInitialized: true,
    isUnlocked: true,
    fetchStatus: fetchStatusMock,
    loadFiles: loadFilesMock,
    setup: vi.fn(),
    unlock: vi.fn(),
    lock: vi.fn(),
    changePin: vi.fn(),
    resetPin: vi.fn(),
    removeFromVault: vi.fn(),
  }),
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({
    showToast: vi.fn(),
    confirm: vi.fn().mockResolvedValue(true),
  }),
}))

async function mountVault() {
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
  return wrapper
}

describe('VaultView load recovery', () => {
  beforeEach(() => {
    loadFilesMock.mockReset()
    fetchStatusMock.mockReset().mockResolvedValue({ initialized: true, unlocked: true })
    routerBackMock.mockReset()
    routerPushMock.mockReset()
    setLocale('en')
  })

  it('shows load failure instead of the empty state, then recovers to the genuine empty state on retry', async () => {
    loadFilesMock
      .mockRejectedValueOnce(new Error('vault files offline'))
      .mockResolvedValueOnce([])

    const wrapper = await mountVault()

    const alert = wrapper.get('[role="alert"]')
    expect(alert.text()).toContain('Failed to load files')
    expect(wrapper.text()).not.toContain('Personal Vault is empty')
    expect(loadFilesMock).toHaveBeenCalledTimes(1)

    const retry = alert.get('button')
    expect(retry.text()).toBe('Retry')
    await retry.trigger('click')
    await flushPromises()

    expect(loadFilesMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Personal Vault is empty')

    wrapper.unmount()
  })
})
