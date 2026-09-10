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

function setupApi() {
  mockedApi.mockImplementation(async (path: string) => {
    if (path === '/shares/with-me') return { shares: [folderShare] } as never
    if (path === '/share-links') return { links: [] } as never
    if (path === '/shared/folders/root') {
      return { folder: { id: 'root', name: 'Root folder' }, folders: [], files: [] } as never
    }
    throw new Error(`Unexpected API call: ${path}`)
  })
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

describe('SharedWithMeView tabpanel contract', () => {
  beforeEach(() => {
    mockedApi.mockReset()
    setupApi()
  })

  it('associates each tab with a labelled panel and hides the inactive panel', async () => {
    const wrapper = await mountView()
    const tabs = wrapper.findAll('[role="tab"]')
    const myPanel = wrapper.get('#shared-panel-my-shares')
    const withMePanel = wrapper.get('#shared-panel-with-me')

    expect(tabs[0]!.attributes('id')).toBe('shared-tab-my-shares')
    expect(tabs[0]!.attributes('aria-controls')).toBe('shared-panel-my-shares')
    expect(myPanel.attributes('role')).toBe('tabpanel')
    expect(myPanel.attributes('aria-labelledby')).toBe('shared-tab-my-shares')
    expect(myPanel.attributes('hidden')).toBeUndefined()

    expect(tabs[1]!.attributes('aria-controls')).toBe('shared-panel-with-me')
    expect(withMePanel.attributes('aria-labelledby')).toBe('shared-tab-with-me')
    expect(withMePanel.attributes('hidden')).toBeDefined()

    await tabs[1]!.trigger('click')
    await flushPromises()

    expect(wrapper.get('#shared-panel-my-shares').attributes('hidden')).toBeDefined()
    expect(wrapper.get('#shared-panel-with-me').attributes('hidden')).toBeUndefined()
  })

  it('keeps selected tab and visible panel aligned when leaving nested browsing', async () => {
    const wrapper = await mountView()
    const tabs = wrapper.findAll('[role="tab"]')
    await tabs[1]!.trigger('click')
    await flushPromises()

    await wrapper.get('.share-item-card').trigger('click')
    await flushPromises()
    expect(wrapper.find('.browse').exists()).toBe(true)

    await tabs[0]!.trigger('click')
    await flushPromises()

    expect(wrapper.find('.browse').exists()).toBe(false)
    expect(tabs[0]!.attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#shared-panel-my-shares').attributes('hidden')).toBeUndefined()
    expect(wrapper.get('#shared-panel-with-me').attributes('hidden')).toBeDefined()
  })
})
