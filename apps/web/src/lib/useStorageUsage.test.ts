import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { StorageUsage } from '@/api/types'
import { useStorageUsage } from './useStorageUsage'

const apiMock = vi.hoisted(() => vi.fn<(path: string) => Promise<unknown>>())

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return { ...actual, api: apiMock }
})

vi.mock('@/lib/i18n', async () => {
  const { ref } = await import('vue')
  return {
    useI18n: () => ({ locale: ref<'vi' | 'en'>('en') }),
  }
})

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

const formatBytes = (bytes: number) => `${bytes} B`

const olderUsage: StorageUsage = { usedBytes: 100, quotaBytes: 1000 }
const newerUsage: StorageUsage = { usedBytes: 700, quotaBytes: 1000 }

describe('useStorageUsage reload ordering', () => {
  beforeEach(() => {
    apiMock.mockReset()
  })

  it('ignores an older success that resolves after a newer successful reload', async () => {
    const older = deferred<StorageUsage>()
    const newer = deferred<StorageUsage>()
    apiMock.mockImplementationOnce(() => older.promise).mockImplementationOnce(() => newer.promise)

    const storage = useStorageUsage(formatBytes)
    const olderReload = storage.reload()
    const newerReload = storage.reload()

    newer.resolve(newerUsage)
    await newerReload
    expect(storage.usage.value).toEqual(newerUsage)
    expect(storage.loading.value).toBe(false)

    older.resolve(olderUsage)
    await olderReload
    expect(storage.usage.value).toEqual(newerUsage)
    expect(storage.unavailable.value).toBe(false)
    expect(storage.loading.value).toBe(false)
  })

  it('ignores an older failure after a newer successful reload', async () => {
    const older = deferred<StorageUsage>()
    const newer = deferred<StorageUsage>()
    apiMock.mockImplementationOnce(() => older.promise).mockImplementationOnce(() => newer.promise)

    const storage = useStorageUsage(formatBytes)
    const olderReload = storage.reload()
    const newerReload = storage.reload()

    newer.resolve(newerUsage)
    await newerReload
    expect(storage.usage.value).toEqual(newerUsage)
    expect(storage.state.value).not.toBe('usage-unavailable')

    older.reject(new Error('stale request failed'))
    await olderReload
    expect(storage.usage.value).toEqual(newerUsage)
    expect(storage.unavailable.value).toBe(false)
    expect(storage.state.value).not.toBe('usage-unavailable')
    expect(storage.loading.value).toBe(false)
  })
})
