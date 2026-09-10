/**
 * @vitest-environment jsdom
 */
import { computed, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TrashView from './TrashView.vue'

const mocks = vi.hoisted(() => ({
  api: vi.fn<(path: string, options?: RequestInit) => Promise<unknown>>(),
  confirm: vi.fn<(options: unknown) => Promise<boolean>>(),
  showToast: vi.fn<(message: string) => void>(),
  openActionSheet: vi.fn<(title: string, actions: unknown[]) => Promise<string | null>>(),
}))

vi.mock('@/api/client', () => ({
  api: mocks.api,
  formatBytes: (bytes: number) => `${bytes} B`,
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({
    confirm: mocks.confirm,
    showToast: mocks.showToast,
    openActionSheet: mocks.openActionSheet,
  }),
}))

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({
    locale: ref('en'),
    t: computed(() => ({
      trashTitle: 'Trash',
      trashRetentionNotice: 'Items are removed later.',
      trashEmpty: 'Trash is empty',
      trashEmptyDesc: 'Deleted items appear here.',
      emptyTrash: 'Empty trash',
      results: 'items',
      folders: 'Folders',
      files: 'Files',
      restore: 'Restore',
      deleteForever: 'Delete forever',
      fileRestored: 'File restored',
      folderRestored: 'Folder restored',
    })),
  }),
}))

vi.mock('@/lib/mimeIcon', () => ({ mimeIcon: () => 'file' }))

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

describe('TrashView Empty Trash', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.confirm.mockResolvedValue(true)
  })

  it('disables the purge action and prevents a duplicate batch while deletion is pending', async () => {
    const deletePending = deferred<unknown>()
    let trashLoads = 0
    mocks.api.mockImplementation((path, options) => {
      if (path === '/trash') {
        trashLoads += 1
        if (trashLoads === 1) {
          return Promise.resolve({
            folders: [{ id: 'folder-1', name: 'Archive', deletedAt: '2026-09-10T00:00:00Z' }],
            files: [],
          })
        }
        return Promise.resolve({ folders: [], files: [] })
      }
      if (path === '/trash/folders/folder-1' && options?.method === 'DELETE') {
        return deletePending.promise
      }
      return Promise.resolve({})
    })

    const wrapper = mount(TrashView, {
      global: {
        stubs: {
          Icon: { template: '<span />' },
          EmptyState: { template: '<div />' },
          LoadingSkeletonTrash: { template: '<div />' },
          TransitionGroup: { template: '<div><slot /></div>' },
        },
      },
    })
    await flushPromises()

    const button = wrapper.get('.empty-trash-btn')
    await button.trigger('click')
    await Promise.resolve()

    expect(button.attributes('disabled')).toBeDefined()
    expect(button.attributes('aria-busy')).toBe('true')
    expect(button.text()).toContain('Deleting 0/1')

    await button.trigger('click')
    expect(mocks.confirm).toHaveBeenCalledTimes(1)
    expect(mocks.api.mock.calls.filter(([path]) => path === '/trash/folders/folder-1')).toHaveLength(1)

    deletePending.resolve({})
    await flushPromises()

    expect(trashLoads).toBe(2)
    expect(mocks.showToast).toHaveBeenCalledWith('Delete forever')
  })
})
