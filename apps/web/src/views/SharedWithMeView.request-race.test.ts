/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import SharedWithMeView from './SharedWithMeView.vue'
import { api } from '@/api/client'

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    api: vi.fn<typeof actual.api>(),
    formatBytes: (bytes: number) => `${bytes} B`,
  }
})

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({
    showToast: vi.fn<() => void>(),
    confirm: vi.fn<() => void>(),
    openActionSheet: vi.fn<() => void>(),
  }),
}))

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({
    locale: ref<'vi' | 'en'>('en'),
    t: ref({
      navShared: 'Shared',
      myShares: 'My shares',
      sharedWithMeTab: 'Shared with me',
      viewGrid: 'Grid view',
      viewList: 'List view',
      allSharedItems: 'All shared items',
      back: 'Back',
      folder: 'Folder',
      folderEmpty: 'Folder is empty',
      retry: 'Retry',
      loading: 'Loading…',
      download: 'Download',
      open: 'Open',
      nothingShared: 'Nothing shared',
      nothingSharedDesc: 'Nothing shared with you yet',
      activeForever: 'Active forever',
    }),
  }),
}))

const mockedApi = vi.mocked(api)

const folderShare = {
  id: 'share-root',
  resourceType: 'folder' as const,
  resourceId: 'root',
  resourceName: 'Root folder',
  owner: { id: 'owner', email: 'owner@example.com', displayName: 'Owner' },
  createdAt: '2026-09-10T00:00:00Z',
}

const rootBrowser = {
  folder: { id: 'root', name: 'Root folder' },
  folders: [],
  files: [],
}

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function setupDeferredRootFolder() {
  const rootRequest = deferred<typeof rootBrowser>()
  mockedApi.mockImplementation(async (path: string) => {
    if (path === '/shares/with-me') return { shares: [folderShare] } as never
    if (path === '/share-links') return { links: [] } as never
    if (path === '/shared/folders/root') return rootRequest.promise as never
    throw new Error(`Unexpected API call: ${path}`)
  })
  return rootRequest
}

async function mountView() {
  const wrapper = mount(SharedWithMeView, {
    global: {
      stubs: {
        Icon: { template: '<span aria-hidden="true" />' },
        EmptyState: { template: '<div class="empty-stub" />' },
        Transition: { template: '<slot />' },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('SharedWithMeView folder navigation request ownership', () => {
  beforeEach(() => {
    mockedApi.mockReset()
  })

  it('ignores a stale root-folder success after switching to My Shares', async () => {
    const rootRequest = setupDeferredRootFolder()
    const wrapper = await mountView()
    const tabs = wrapper.findAll('[role="tab"]')

    await tabs[1]!.trigger('click')
    await flushPromises()
    await wrapper.get('.share-item-card').trigger('click')

    await tabs[0]!.trigger('click')
    await flushPromises()

    expect(wrapper.find('.browse').exists()).toBe(false)
    expect(tabs[0]!.attributes('aria-selected')).toBe('true')

    rootRequest.resolve(rootBrowser)
    await flushPromises()

    expect(wrapper.find('.browse').exists()).toBe(false)
    expect(wrapper.find('.folder-name').exists()).toBe(false)
    expect(wrapper.get('#shared-panel-my-shares').attributes('hidden')).toBeUndefined()
  })

  it('ignores stale root-folder failure and finally after switching to My Shares', async () => {
    const rootRequest = setupDeferredRootFolder()
    const wrapper = await mountView()
    const tabs = wrapper.findAll('[role="tab"]')

    await tabs[1]!.trigger('click')
    await flushPromises()
    await wrapper.get('.share-item-card').trigger('click')

    await tabs[0]!.trigger('click')
    await flushPromises()

    rootRequest.reject(new Error('network'))
    await flushPromises()

    expect(wrapper.find('.browse').exists()).toBe(false)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('network')
  })

  it('keeps normal root-folder navigation functional after request ownership checks', async () => {
    const rootRequest = setupDeferredRootFolder()
    const wrapper = await mountView()
    const tabs = wrapper.findAll('[role="tab"]')

    await tabs[1]!.trigger('click')
    await flushPromises()
    await wrapper.get('.share-item-card').trigger('click')

    rootRequest.resolve(rootBrowser)
    await flushPromises()

    expect(wrapper.find('.browse').exists()).toBe(true)
    expect(wrapper.get('.folder-name').text()).toBe('Root folder')
    expect(wrapper.find('.browse-status').exists()).toBe(false)
  })
})
