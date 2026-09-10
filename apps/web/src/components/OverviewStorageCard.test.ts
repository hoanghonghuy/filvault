/**
 * @vitest-environment jsdom
 */
import { computed } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import OverviewStorageCard from './OverviewStorageCard.vue'

const { apiMock } = vi.hoisted(() => ({
  apiMock: vi.fn<(...args: unknown[]) => Promise<unknown>>(),
}))

vi.mock('@/api/client', () => ({
  api: apiMock,
  formatBytes: (n: number) => `${n} B`,
}))

const messages = {
  storageLabel: 'Storage',
  storageLoading: 'Loading storage…',
  storageUnavailable: 'Storage usage unavailable',
  storageNearQuota: 'Almost full',
  storageFullQuota: 'Storage full',
  storageFreeUp: 'Free up space',
  retry: 'Retry',
  myCloud: 'My cloud',
  manageStorage: 'Manage storage',
}

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({ t: computed(() => messages) }),
}))

const mountCard = () =>
  mount(OverviewStorageCard, {
    global: {
      stubs: {
        RouterLink: {
          template: '<a :href="to"><slot /></a>',
          props: ['to'],
        },
      },
    },
  })

describe('OverviewStorageCard', () => {
  beforeEach(() => apiMock.mockReset())

  it('shows live normal usage from /storage rather than an auth snapshot', async () => {
    apiMock.mockResolvedValueOnce({ usedBytes: 250, quotaBytes: 1000 })
    const wrapper = mountCard()
    await flushPromises()

    expect(apiMock).toHaveBeenCalledWith('/storage')
    expect(wrapper.text()).toContain('250 B / 1000 B')
    expect(wrapper.text()).toContain('25%')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('25')
  })

  it('shows unavailable + retry without fabricated stale usage', async () => {
    apiMock.mockRejectedValueOnce(new Error('network'))
    const wrapper = mountCard()
    await flushPromises()

    expect(wrapper.text()).toContain('Storage usage unavailable')
    expect(wrapper.find('[role="progressbar"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('0 B / 0 B')

    apiMock.mockResolvedValueOnce({ usedBytes: 400, quotaBytes: 1000 })
    await wrapper.get('[data-testid="overview-storage-retry"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('400 B / 1000 B')
  })

  it('uses the shared near-quota threshold and non-color status copy', async () => {
    apiMock.mockResolvedValueOnce({ usedBytes: 900, quotaBytes: 1000 })
    const wrapper = mountCard()
    await flushPromises()

    expect(wrapper.attributes('data-state')).toBe('near-quota')
    expect(wrapper.text()).toContain('Almost full')
    expect(wrapper.get('.storage-progress-fill').classes()).toContain('warning')
  })

  it('shows full quota with a recovery route to Trash', async () => {
    apiMock.mockResolvedValueOnce({ usedBytes: 1000, quotaBytes: 1000 })
    const wrapper = mountCard()
    await flushPromises()

    expect(wrapper.attributes('data-state')).toBe('full-quota')
    expect(wrapper.text()).toContain('Storage full')
    expect(wrapper.get('.storage-progress-fill').classes()).toContain('danger')
    expect(wrapper.get('.storage-free-up').attributes('href')).toBe('/trash')
  })
})
