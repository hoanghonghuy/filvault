/**
 * @vitest-environment jsdom
 */
import { computed } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import StorageBar from './StorageBar.vue'

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
  storageUnavailableHint: 'Usage may be outdated until it loads again.',
  storageNearQuota: 'Almost full',
  storageNearQuotaHint: 'Free up space in Trash or delete files.',
  storageFullQuota: 'Storage full',
  storageFullQuotaHint: 'Uploads may fail until you free space.',
  storageFreeUp: 'Free up space',
  retry: 'Retry',
  navTrash: 'Trash',
}

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({
    t: computed(() => messages),
  }),
}))

describe('StorageBar', () => {
  beforeEach(() => {
    apiMock.mockReset()
  })

  it('shows a compact unavailable state with retry when /storage fails', async () => {
    apiMock.mockRejectedValueOnce(new Error('network'))

    const wrapper = mount(StorageBar, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a :href="to"><slot /></a>',
            props: ['to'],
          },
        },
      },
    })
    await flushPromises()

    expect(wrapper.find('.storage-bar').exists()).toBe(true)
    expect(wrapper.text()).toContain('Storage usage unavailable')
    expect(wrapper.get('[data-testid="storage-retry"]').text()).toContain('Retry')
    expect(wrapper.find('[role="progressbar"]').exists()).toBe(false)
  })

  it('exposes warning text and tone for near-quota usage', async () => {
    apiMock.mockResolvedValueOnce({ usedBytes: 950, quotaBytes: 1000 })

    const wrapper = mount(StorageBar, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Almost full')
    expect(wrapper.get('.fill').classes()).toContain('warning')
    const bar = wrapper.get('[role="progressbar"]')
    expect(bar.attributes('aria-valuenow')).toBe('95')
  })

  it('surfaces full-quota impact and a recovery action', async () => {
    apiMock.mockResolvedValueOnce({ usedBytes: 1000, quotaBytes: 1000 })

    const wrapper = mount(StorageBar, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a :href="to"><slot /></a>',
            props: ['to'],
          },
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Storage full')
    expect(wrapper.text()).toContain('Uploads may fail')
    expect(wrapper.get('.fill').classes()).toContain('danger')
    expect(wrapper.get('[data-testid="storage-action"]').attributes('href')).toBe('/trash')
  })

  it('keeps progressbar aria values valid for zero quota', async () => {
    apiMock.mockResolvedValueOnce({ usedBytes: 100, quotaBytes: 0 })

    const wrapper = mount(StorageBar, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()

    const bar = wrapper.get('[role="progressbar"]')
    expect(bar.attributes('aria-valuenow')).toBe('0')
    expect(bar.attributes('aria-valuemin')).toBe('0')
    expect(bar.attributes('aria-valuemax')).toBe('100')
    expect(Number.isFinite(Number(bar.attributes('aria-valuenow')))).toBe(true)
  })
})
