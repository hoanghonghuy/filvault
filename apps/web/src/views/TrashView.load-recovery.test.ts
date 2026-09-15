/**
 * @vitest-environment jsdom
 */
import { computed, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TrashView from './TrashView.vue'

const mocks = vi.hoisted(() => ({
  api: vi.fn<(path: string, options?: RequestInit) => Promise<unknown>>(),
}))

vi.mock('@/api/client', () => ({
  api: mocks.api,
  formatBytes: (bytes: number) => `${bytes} B`,
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      trashAutoDeleteEnabled: true,
      trashRetentionDays: 30,
    },
  }),
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({
    confirm: vi.fn<(options: unknown) => Promise<boolean>>().mockResolvedValue(true),
    showToast: vi.fn<(message: string) => void>(),
    openActionSheet: vi.fn<(title: string, actions: unknown[]) => Promise<string | null>>().mockResolvedValue(null),
  }),
}))

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({
    locale: ref('en'),
    t: computed(() => ({
      trashTitle: 'Trash',
      trashEmpty: 'Trash is empty',
      trashEmptyDesc: 'Deleted items appear here.',
      emptyTrash: 'Empty trash',
      folders: 'Folders',
      files: 'Files',
      restore: 'Restore',
      deleteForever: 'Delete forever',
      fileRestored: 'File restored',
      folderRestored: 'Folder restored',
      retry: 'Retry',
    })),
  }),
}))

vi.mock('@/lib/mimeIcon', () => ({ mimeIcon: () => 'file' }))

function mountTrash() {
  return mount(TrashView, {
    global: {
      stubs: {
        Icon: { template: '<span />' },
        EmptyState: { template: '<div data-testid="trash-empty-state" />' },
        LoadingSkeletonTrash: { template: '<div data-testid="trash-loading" />' },
        TransitionGroup: { template: '<div><slot /></div>' },
      },
    },
  })
}

describe('TrashView initial load recovery', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not expose false-ready or true-empty controls when the initial load fails', async () => {
    mocks.api.mockRejectedValueOnce(new Error('offline'))

    const wrapper = mountTrash()
    await flushPromises()

    const recovery = wrapper.get('.trash-load-error')
    expect(recovery.attributes('role')).toBe('alert')
    expect(recovery.text()).toContain('Could not load trash')
    expect(recovery.get('.trash-retry-btn').text()).toBe('Retry')
    expect(wrapper.find('.trash-toolbar').exists()).toBe(false)
    expect(wrapper.find('.empty-trash-btn').exists()).toBe(false)
    expect(wrapper.find('[data-testid="trash-empty-state"]').exists()).toBe(false)
  })

  it('recovers from an initial failure without reloading the page', async () => {
    mocks.api
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({
        folders: [{ id: 'folder-1', name: 'Archive', deletedAt: '2026-09-10T00:00:00Z' }],
        files: [],
      })

    const wrapper = mountTrash()
    await flushPromises()

    await wrapper.get('.trash-retry-btn').trigger('click')
    await flushPromises()

    expect(mocks.api.mock.calls.filter(([path]) => path === '/trash')).toHaveLength(2)
    expect(wrapper.find('.trash-load-error').exists()).toBe(false)
    expect(wrapper.get('.trash-toolbar').text()).toContain('1 result')
    expect(wrapper.text()).toContain('Archive')
  })
})
