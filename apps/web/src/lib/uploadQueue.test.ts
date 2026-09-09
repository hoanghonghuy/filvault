/**
 * @vitest-environment jsdom
 */
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { ApiError } from '@/api/client'
import {
  FileUploadQueue,
  computeAggregateProgress,
  type UploadQueueItem,
  type UploadQueueDeps,
} from './uploadQueue'

function makeFile(name: string, size = 1000): File {
  return new File(['x'.repeat(size)], name, { type: 'text/plain' })
}

function makeItem(overrides: Partial<UploadQueueItem> & Pick<UploadQueueItem, 'id'>): UploadQueueItem {
  const file = overrides.file ?? makeFile('test.txt')
  return {
    id: overrides.id,
    file,
    displayName: overrides.displayName ?? file.name,
    folderParts: overrides.folderParts ?? [],
    rootFolderId: overrides.rootFolderId ?? null,
    status: overrides.status ?? 'queued',
    progress: overrides.progress ?? 0,
    error: overrides.error,
    controller: overrides.controller ?? new AbortController(),
    attempt: overrides.attempt ?? 1,
    resolvedFolderId: overrides.resolvedFolderId,
  }
}

function createDeps(overrides: Partial<UploadQueueDeps> = {}): UploadQueueDeps {
  return {
    createSession: vi.fn<UploadQueueDeps['createSession']>(async () => ({
      fileId: 'file-1',
      uploadUrl: 'https://example.com/upload',
    })),
    uploadBytes: vi.fn<UploadQueueDeps['uploadBytes']>(async (_url, _file, _type, onProgress) => {
      onProgress?.(1)
    }),
    completeUpload: vi.fn<UploadQueueDeps['completeUpload']>(async () => undefined),
    ensureFolderPath: vi.fn<UploadQueueDeps['ensureFolderPath']>(async () => 'folder-1'),
    resolveContentType: vi.fn<UploadQueueDeps['resolveContentType']>(() => 'text/plain'),
    formatError: vi.fn<UploadQueueDeps['formatError']>((e: unknown) =>
      e instanceof Error ? e.message : 'error',
    ),
    ...overrides,
  }
}

describe('computeAggregateProgress', () => {
  it('returns null when there are no active items', () => {
    expect(computeAggregateProgress([])).toBeNull()
    expect(
      computeAggregateProgress([makeItem({ id: '1', status: 'cancelled' })]),
    ).toBeNull()
  })

  it('uses byte-weighted progress and never reaches 100% with unresolved items', () => {
    const items = [
      makeItem({ id: '1', status: 'completed', file: makeFile('a.txt', 1000) }),
      makeItem({ id: '2', status: 'failed', file: makeFile('b.txt', 1000) }),
    ]
    expect(computeAggregateProgress(items)).toBe(0.5)
  })

  it('includes partial byte progress for uploading items', () => {
    const items = [
      makeItem({ id: '1', status: 'completed', file: makeFile('a.txt', 1000) }),
      makeItem({ id: '2', status: 'uploading', progress: 0.5, file: makeFile('b.txt', 1000) }),
      makeItem({ id: '3', status: 'queued', file: makeFile('c.txt', 1000) }),
    ]
    expect(computeAggregateProgress(items)).toBeCloseTo(0.5, 2)
  })

  it('reaches 100% only when every active item completed', () => {
    const items = [
      makeItem({ id: '1', status: 'completed', file: makeFile('a.txt', 500) }),
      makeItem({ id: '2', status: 'completed', file: makeFile('b.txt', 500) }),
    ]
    expect(computeAggregateProgress(items)).toBe(1)
  })
})

describe('FileUploadQueue', () => {
  let deps: UploadQueueDeps

  beforeEach(() => {
    deps = createDeps()
  })

  it('uploads all files successfully', async () => {
    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('one.txt'), makeFile('two.txt')], null)
    await queue.run()

    expect(deps.createSession).toHaveBeenCalledTimes(2)
    expect(deps.uploadBytes).toHaveBeenCalledTimes(2)
    expect(deps.completeUpload).toHaveBeenCalledTimes(2)
    expect(queue.items.every((item) => item.status === 'completed')).toBe(true)
    expect(computeAggregateProgress(queue.items)).toBe(1)
  })

  it('continues batch when one item fails', async () => {
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async (_url, file) => {
      if (file.name === 'bad.txt') {
        throw new ApiError('UPLOAD_FAILED', 'Upload network error', 0)
      }
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('bad.txt'), makeFile('good.txt')], null)
    await queue.run()

    const bad = queue.items.find((item) => item.file.name === 'bad.txt')
    const good = queue.items.find((item) => item.file.name === 'good.txt')
    expect(bad?.status).toBe('failed')
    expect(good?.status).toBe('completed')
    expect(deps.completeUpload).toHaveBeenCalledTimes(1)
  })

  it('retries only the failed item without re-uploading completed items', async () => {
    let failFirst = true
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async (_url, file) => {
      if (file.name === 'bad.txt' && failFirst) {
        failFirst = false
        throw new ApiError('UPLOAD_FAILED', 'Upload network error', 0)
      }
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('bad.txt'), makeFile('good.txt')], null)
    await queue.run()

    const failedId = queue.items.find((item) => item.file.name === 'bad.txt')!.id
    expect(deps.completeUpload).toHaveBeenCalledTimes(1)

    queue.retry(failedId)
    await queue.run()

    expect(deps.completeUpload).toHaveBeenCalledTimes(2)
    expect(queue.items.find((item) => item.file.name === 'bad.txt')?.status).toBe('completed')
    expect(queue.items.find((item) => item.file.name === 'good.txt')?.status).toBe('completed')
  })

  it('cancels queued items and aborts in-flight uploads', async () => {
    let uploadStarted: (() => void) | undefined
    let releaseUpload: (() => void) | undefined
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async (_url, _file, _type, _onProgress, signal) => {
      uploadStarted?.()
      await new Promise<void>((resolve, reject) => {
        releaseUpload = resolve
        signal?.addEventListener('abort', () => reject(new ApiError('UPLOAD_CANCELED', 'Upload canceled', 0)))
      })
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('active.txt'), makeFile('queued.txt')], null)

    const runPromise = queue.run()
    await new Promise<void>((resolve) => {
      uploadStarted = resolve
    })

    const active = queue.items.find((item) => item.status === 'uploading')
    const queued = queue.items.find((item) => item.status === 'queued')
    expect(active).toBeTruthy()
    expect(queued).toBeTruthy()

    queue.cancel(queued!.id)
    expect(queue.items.find((item) => item.id === queued!.id)?.status).toBe('cancelled')

    queue.cancel(active!.id)
    releaseUpload?.()
    await runPromise

    expect(queue.items.find((item) => item.id === active!.id)?.status).toBe('cancelled')
    expect(deps.completeUpload).not.toHaveBeenCalled()
  })

  it('ignores duplicate retry and cancel actions', async () => {
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async () => {
      throw new ApiError('UPLOAD_FAILED', 'Upload network error', 0)
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('bad.txt')], null)
    await queue.run()

    const item = queue.items[0]!
    expect(item.status).toBe('failed')

    expect(queue.retry(item.id)).toBe(true)
    expect(queue.retry(item.id)).toBe(false)
    expect(queue.items[0]?.status).toBe('queued')

    expect(queue.cancel(item.id)).toBe(true)
    expect(queue.cancel(item.id)).toBe(false)
    expect(queue.items[0]?.status).toBe('cancelled')
  })

  it('prevents duplicate completion on stale attempts after rapid retry', async () => {
    let attempt = 0
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async () => {
      attempt += 1
      if (attempt === 1) {
        throw new ApiError('UPLOAD_FAILED', 'Upload network error', 0)
      }
      await new Promise((resolve) => setTimeout(resolve, 20))
    })

    const completeUpload = vi.fn<UploadQueueDeps['completeUpload']>(async () => undefined)
    deps.completeUpload = completeUpload

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('file.txt')], null)
    await queue.run()
    expect(queue.items[0]?.status).toBe('failed')

    queue.retry(queue.items[0]!.id)
    queue.retry(queue.items[0]!.id)
    const runPromise = queue.run()
    await new Promise((resolve) => setTimeout(resolve, 5))
    queue.cancel(queue.items[0]!.id)
    await runPromise

    expect(completeUpload).not.toHaveBeenCalled()
  })

  it('preserves folder relative paths across retry and partial failure', async () => {
    const ensureFolderPath = vi.fn<UploadQueueDeps['ensureFolderPath']>(async (parts: string[]) =>
      `folder-${parts.join('/')}`,
    )
    deps.ensureFolderPath = ensureFolderPath
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async (_url, file) => {
      if (file.name === 'bad.txt') {
        throw new ApiError('UPLOAD_FAILED', 'Upload network error', 0)
      }
    })

    const queue = new FileUploadQueue(deps)
    const nested = makeFile('good.txt')
    Object.defineProperty(nested, 'webkitRelativePath', { value: 'photos/2024/good.txt' })
    const bad = makeFile('bad.txt')
    Object.defineProperty(bad, 'webkitRelativePath', { value: 'photos/2024/bad.txt' })

    queue.enqueueFiles([bad, nested], 'root-folder', { isFolderUpload: true })
    await queue.run()

    expect(ensureFolderPath).toHaveBeenCalledWith(['photos', '2024'], 'root-folder', expect.any(Map))

    const failedId = queue.items.find((item) => item.file.name === 'bad.txt')!.id
    queue.retry(failedId)
    await queue.run()

    expect(ensureFolderPath).toHaveBeenCalledWith(['photos', '2024'], 'root-folder', expect.any(Map))
    expect(deps.createSession).toHaveBeenCalledWith(
      expect.objectContaining({ folderId: 'folder-photos/2024' }),
    )
  })

  it('records resolvedName when conflict auto-rename succeeds', async () => {
    let attempts = 0
    deps.createSession = vi.fn<UploadQueueDeps['createSession']>(async (params) => {
      attempts += 1
      if (attempts === 1) {
        throw new ApiError('CONFLICT', 'Name already exists in this location.', 409)
      }
      expect(params.name).toBe('dup (1).pdf')
      return { fileId: 'file-1', uploadUrl: 'https://example.com/upload' }
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('dup.pdf')], null)
    await queue.run()

    expect(queue.items[0]?.resolvedName).toBe('dup (1).pdf')
    expect(queue.items[0]?.status).toBe('completed')
  })

  it('removes failed items when dismissFailed is called', async () => {
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async () => {
      throw new ApiError('UPLOAD_FAILED', 'Upload network error', 0)
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('bad.txt')], null)
    await queue.run()

    const failedId = queue.items[0]!.id
    expect(queue.dismissFailed(failedId)).toBe(true)
    expect(queue.items).toHaveLength(0)
    expect(queue.dismissFailed(failedId)).toBe(false)
  })

  it('appends and processes files enqueued while another upload is in flight', async () => {
    let releaseFirst: (() => void) | undefined
    let firstStarted: (() => void) | undefined
    deps.uploadBytes = vi.fn<UploadQueueDeps['uploadBytes']>(async (_url, file) => {
      if (file.name === 'first.txt') {
        firstStarted?.()
        await new Promise<void>((resolve) => {
          releaseFirst = resolve
        })
      }
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('first.txt', 1000)], null)
    const runPromise = queue.run()

    await new Promise<void>((resolve) => {
      firstStarted = resolve
    })
    expect(queue.items.find((item) => item.file.name === 'first.txt')?.status).toBe('uploading')

    queue.enqueueFiles([makeFile('second.txt', 2000), makeFile('third.txt', 3000)], null)
    queue.scheduleRun()

    expect(queue.items).toHaveLength(3)
    expect(queue.items.filter((item) => item.status === 'queued')).toHaveLength(2)
    expect(computeAggregateProgress(queue.items)).toBeLessThan(1)

    releaseFirst?.()
    await runPromise
    await queue.run()

    expect(deps.createSession).toHaveBeenCalledTimes(3)
    expect(queue.items.every((item) => item.status === 'completed')).toBe(true)
    expect(computeAggregateProgress(queue.items)).toBe(1)
  })

  it('surfaces actionable conflict errors without auto-renaming when rename budget is exhausted', async () => {
    deps.createSession = vi.fn<UploadQueueDeps['createSession']>(async () => {
      throw new ApiError('CONFLICT', 'Name already exists in this location.', 409)
    })

    const queue = new FileUploadQueue(deps)
    queue.enqueueFiles([makeFile('dup.txt')], null)
    await queue.run()

    expect(queue.items[0]?.status).toBe('failed')
    expect(queue.items[0]?.error).toContain('already exists')
  })
})
