/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import OverviewView from './OverviewView.vue'
import { setLocale } from '@/lib/i18n'

const apiMock = vi.hoisted(() => vi.fn<(path: string) => Promise<unknown>>())

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return { ...actual, api: apiMock }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { displayName: 'Huy' } }),
}))

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: OverviewView },
      { path: '/files', component: { template: '<div>Files</div>' } },
      { path: '/photos', component: { template: '<div>Photos</div>' } },
      { path: '/photos/preview/:id', component: { template: '<div>Preview</div>' } },
      { path: '/shared', component: { template: '<div>Shared</div>' } },
      { path: '/vault', component: { template: '<div>Vault</div>' } },
      { path: '/trash', component: { template: '<div>Trash</div>' } },
      { path: '/settings', component: { template: '<div>Settings</div>' } },
    ],
  })
}

async function mountOverview() {
  const router = makeRouter()
  await router.push('/')
  await router.isReady()
  const wrapper = mount(OverviewView, {
    global: {
      plugins: [router],
      stubs: {
        AppIcon: true,
        OverviewStorageCard: true,
        PhotoThumb: true,
      },
    },
  })
  await flushPromises()
  return { wrapper, router }
}

describe('OverviewView recovery', () => {
  beforeEach(() => {
    apiMock.mockReset()
    setLocale('en')
  })

  it('offers one localized retry for a combined failure and recovers through the existing load path', async () => {
    apiMock
      .mockRejectedValueOnce(new Error('files offline'))
      .mockRejectedValueOnce(new Error('photos offline'))
      .mockResolvedValueOnce({ files: [] })
      .mockResolvedValueOnce({ files: [] })
      .mockResolvedValueOnce({ groups: [] })
      .mockResolvedValueOnce({ files: [] })

    const { wrapper } = await mountOverview()

    const retry = wrapper.get('button.overview-retry')
    expect(retry.text()).toBe('Retry')
    expect(apiMock).toHaveBeenCalledTimes(3)

    const firstRetry = retry.trigger('click')
    const duplicateRetry = retry.trigger('click')
    await Promise.all([firstRetry, duplicateRetry])
    expect(apiMock).toHaveBeenCalledTimes(6)

    await flushPromises()
    expect(wrapper.find('.overview-error').exists()).toBe(false)
    expect(wrapper.find('.content-tabs-section').exists()).toBe(true)
    wrapper.unmount()
  })

  it('preserves partial success instead of replacing it with the global recovery state', async () => {
    apiMock
      .mockResolvedValueOnce({ files: [] })
      .mockRejectedValueOnce(new Error('photos offline'))
      .mockResolvedValueOnce({ files: [] })

    const { wrapper } = await mountOverview()

    expect(wrapper.find('.overview-error').exists()).toBe(false)
    expect(wrapper.find('.content-tabs-section').exists()).toBe(true)
    expect(wrapper.find('button.overview-retry').exists()).toBe(false)
    wrapper.unmount()
  })
})
