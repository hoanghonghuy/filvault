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

const rootShare = {
  id: 'share-root',
  resourceType: 'folder' as const,
  resourceId: 'root',
  resourceName: 'Root folder',
  owner: { id: 'owner', email: 'owner@example.com', displayName: 'Owner' },
  createdAt: '2026-09-10T00:00:00Z',
}

const rootBrowser = {
  folder: { id: 'root', name: 'Root folder' },
  folders: [{ id: 'child', name: 'Child folder' }],
  files: [],
}

const childBrowser = {
  folder: { id: 'child', name: 'Child folder' },
  folders: [],
  files: [{ id: 'deep-file', name: 'deep.txt', mimeType: 'text/plain', sizeBytes: 7 }],
}

function setupApi(options: { failChildOnce?: boolean } = {}) {
  let childAttempts = 0
  mockedApi.mockImplementation(async (path: string) => {
    if (path === '/shares/with-me') return { shares: [rootShare] } as never
    if (path === '/share-links') return { links: [] } as never
    if (path === '/shared/folders/root') return rootBrowser as never
    if (path === '/shared/folders/child') {
      childAttempts += 1
      if (options.failChildOnce && childAttempts === 1) throw new Error('network')
      return childBrowser as never
    }
    throw new Error(`Unexpected API call: ${path}`)
  })
}

async function mountWithMe() {
  const wrapper = mount(SharedWithMeView, {
    global: {
      stubs: {
        Icon: { template: '<span aria-hidden="true" />' },
        EmptyState: { template: '<div class="empty-stub" />' },
      },
    },
  })
  await flushPromises()
  const tabs = wrapper.findAll('[role="tab"]')
  await tabs[1]!.trigger('click')
  await flushPromises()
  return wrapper
}

describe('SharedWithMeView nested folder navigation', () => {
  beforeEach(() => {
    mockedApi.mockReset()
  })

  it('drills into a child folder and backs up one level before leaving the shared root', async () => {
    setupApi()
    const wrapper = await mountWithMe()

    await wrapper.get('.share-item-card').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Root folder')

    const child = wrapper.get('button.folder-card')
    expect(child.element.tagName).toBe('BUTTON')
    await child.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Child folder')
    expect(wrapper.text()).toContain('deep.txt')
    expect(wrapper.find('.browse-path').text()).toContain('Root folder')
    expect(wrapper.find('.browse-path').text()).toContain('Child folder')

    await wrapper.get('.back-btn').trigger('click')
    expect(wrapper.get('.folder-name').text()).toBe('Root folder')
    expect(wrapper.find('button.folder-card').exists()).toBe(true)

    await wrapper.get('.back-btn').trigger('click')
    expect(wrapper.find('.browse').exists()).toBe(false)
    expect(wrapper.text()).toContain('Root folder')
  })

  it('keeps the last valid folder visible after a child load failure and retries in place', async () => {
    setupApi({ failChildOnce: true })
    const wrapper = await mountWithMe()

    await wrapper.get('.share-item-card').trigger('click')
    await flushPromises()
    await wrapper.get('button.folder-card').trigger('click')
    await flushPromises()

    expect(wrapper.get('.folder-name').text()).toBe('Root folder')
    expect(wrapper.find('.browse-error').exists()).toBe(true)
    expect(wrapper.find('button.folder-card').exists()).toBe(true)

    await wrapper.get('.browse-error .retry-btn').trigger('click')
    await flushPromises()

    expect(wrapper.get('.folder-name').text()).toBe('Child folder')
    expect(wrapper.find('.browse-error').exists()).toBe(false)
  })
})
