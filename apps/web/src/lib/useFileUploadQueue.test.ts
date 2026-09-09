/**
 * @vitest-environment jsdom
 */
import { computed, defineComponent } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { uploadToPresigned } from '@/api/client'
import { useFileUploadQueue } from './useFileUploadQueue'

let releaseUpload: (() => void) | undefined
let releaseSettle: (() => void) | undefined

vi.mock('@/api/client', () => ({
  api: vi.fn<(path: string, init?: RequestInit) => Promise<unknown>>(async (path, init) => {
    if (path === '/files/upload-sessions') {
      return { fileId: 'file-1', uploadUrl: 'https://example.com/upload' }
    }
    if (path.startsWith('/files/') && init?.method === 'POST') {
      return undefined
    }
    if (path.startsWith('/files/') && init?.method === 'DELETE') {
      return undefined
    }
    throw new Error(`Unexpected API path: ${path}`)
  }),
  uploadToPresigned: vi.fn<
    (
      url: string,
      file: File,
      type: string,
      onProgress?: (ratio: number) => void,
      signal?: AbortSignal,
    ) => Promise<void>
  >(async (_url, file, _type, onProgress) => {
      if (file.name === 'in-flight.txt') {
        onProgress?.(0.25)
        await new Promise<void>((resolve) => {
          releaseUpload = resolve
        })
        onProgress?.(1)
        return
      }
      onProgress?.(1)
    }),
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

function makeFile(name: string, size = 1000): File {
  return new File(['x'.repeat(size)], name, { type: 'text/plain' })
}

async function waitFor(
  predicate: () => boolean,
  timeoutMs = 3000,
  intervalMs = 20,
): Promise<void> {
  const started = Date.now()
  while (!predicate()) {
    if (Date.now() - started > timeoutMs) {
      throw new Error('condition not met before timeout')
    }
    await flushPromises()
    await new Promise((resolve) => setTimeout(resolve, intervalMs))
  }
}

const Harness = defineComponent({
  setup() {
    const queue = useFileUploadQueue({
      getRootFolderId: () => null,
      onBatchSettled: async () => undefined,
      showToast: vi.fn<(message: string, tone?: 'success' | 'info') => void>(),
    })
    const summary = computed(() => ({
      count: queue.progressItems.value.length,
      queued: queue.progressItems.value.filter((item) => item.status === 'queued').length,
      uploading: queue.progressItems.value.some((item) => item.status === 'uploading'),
      progress: queue.aggregateProgress.value ?? 0,
    }))
    return { queue, summary }
  },
  template: `
    <div
      data-testid="summary"
      :data-count="summary.count"
      :data-queued="summary.queued"
      :data-uploading="summary.uploading"
      :data-progress="summary.progress"
    />
  `,
})

describe('useFileUploadQueue', () => {
  beforeEach(() => {
    releaseUpload = undefined
    releaseSettle = undefined
    vi.mocked(uploadToPresigned).mockClear()
  })

  it('appends to the same queue while an upload is active and keeps aggregate progress honest', async () => {
    const wrapper = mount(Harness)
    const { queue } = wrapper.vm as { queue: ReturnType<typeof useFileUploadQueue> }

    queue.enqueueFiles([makeFile('in-flight.txt', 4000)])
    await waitFor(() => wrapper.get('[data-testid="summary"]').attributes('data-uploading') === 'true')

    queue.enqueueFiles([makeFile('queued-a.txt', 2000), makeFile('queued-b.txt', 2000)])
    await waitFor(() => Number(wrapper.get('[data-testid="summary"]').attributes('data-count')) === 3)

    expect(Number(wrapper.get('[data-testid="summary"]').attributes('data-queued'))).toBe(2)
    expect(Number(wrapper.get('[data-testid="summary"]').attributes('data-progress'))).toBeLessThan(1)

    releaseUpload?.()
    await waitFor(() => vi.mocked(uploadToPresigned).mock.calls.length === 3)

    wrapper.unmount()
  })

  it('processes files enqueued while batch settlement is still awaiting', async () => {
    const SettleHarness = defineComponent({
      setup() {
        const queue = useFileUploadQueue({
          getRootFolderId: () => null,
          onBatchSettled: async () => {
            await new Promise<void>((resolve) => {
              releaseSettle = resolve
            })
          },
          showToast: vi.fn<(message: string, tone?: 'success' | 'info') => void>(),
        })
        const summary = computed(() => ({
          count: queue.progressItems.value.length,
          queued: queue.progressItems.value.filter((item) => item.status === 'queued').length,
          uploading: queue.progressItems.value.some((item) => item.status === 'uploading'),
          progress: queue.aggregateProgress.value ?? 0,
        }))
        return { queue, summary }
      },
      template: `
        <div
          data-testid="summary"
          :data-count="summary.count"
          :data-queued="summary.queued"
          :data-uploading="summary.uploading"
          :data-progress="summary.progress"
        />
      `,
    })

    const wrapper = mount(SettleHarness)
    const { queue } = wrapper.vm as { queue: ReturnType<typeof useFileUploadQueue> }

    queue.enqueueFiles([makeFile('first.txt')])
    await waitFor(() => vi.mocked(uploadToPresigned).mock.calls.length === 1)
    await waitFor(() => releaseSettle !== undefined)

    queue.enqueueFiles([makeFile('during-settle.txt')])
    await waitFor(() => Number(wrapper.get('[data-testid="summary"]').attributes('data-queued')) >= 1)

    releaseSettle?.()
    await waitFor(() => vi.mocked(uploadToPresigned).mock.calls.length === 2)
    expect(vi.mocked(uploadToPresigned).mock.calls.length).toBe(2)
    expect(Number(wrapper.get('[data-testid="summary"]').attributes('data-queued'))).toBe(0)

    wrapper.unmount()
  })
})
