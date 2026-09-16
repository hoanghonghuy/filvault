import { describe, expect, it } from 'vitest'
import type { DownloadURL } from '@/api/types'
import {
  createFilesMediaPreviewSession,
  runFilesMediaPreviewDownload,
} from './filesMediaPreview'

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function file(id: string) {
  return { id, name: `${id}.jpg`, mimeType: 'image/jpeg' }
}

function download(id: string): DownloadURL {
  return {
    downloadUrl: `https://example.com/${id}`,
    expiresAt: '2026-01-01T00:05:00.000Z',
  }
}

describe('filesMediaPreview request ownership', () => {
  it('keeps the newer preview authoritative when an older download resolves late', async () => {
    const session = createFilesMediaPreviewSession()
    const downloadA = deferred<DownloadURL>()
    const downloadB = deferred<DownloadURL>()

    const requestA = session.beginPreview()
    const requestB = session.beginPreview()

    const promiseA = runFilesMediaPreviewDownload({
      file: file('media-a'),
      requestSequence: requestA,
      isCurrentPreview: session.isCurrentPreview.bind(session),
      fetchDownload: (fileId) => (fileId === 'media-a' ? downloadA.promise : Promise.reject(new Error('unexpected'))),
    })
    const promiseB = runFilesMediaPreviewDownload({
      file: file('media-b'),
      requestSequence: requestB,
      isCurrentPreview: session.isCurrentPreview.bind(session),
      fetchDownload: (fileId) => (fileId === 'media-b' ? downloadB.promise : Promise.reject(new Error('unexpected'))),
    })

    downloadB.resolve(download('media-b'))
    const resultB = await promiseB
    expect(resultB).toEqual({
      status: 'success',
      preview: {
        id: 'media-b',
        name: 'media-b.jpg',
        mimeType: 'image/jpeg',
        url: 'https://example.com/media-b',
      },
    })

    downloadA.resolve(download('media-a'))
    const resultA = await promiseA
    expect(resultA).toEqual({ status: 'stale' })
  })

  it('ignores stale preview failure while a newer preview is authoritative', async () => {
    const session = createFilesMediaPreviewSession()
    const downloadA = deferred<DownloadURL>()
    const downloadB = deferred<DownloadURL>()

    const requestA = session.beginPreview()
    const requestB = session.beginPreview()

    const promiseA = runFilesMediaPreviewDownload({
      file: file('media-a'),
      requestSequence: requestA,
      isCurrentPreview: session.isCurrentPreview.bind(session),
      fetchDownload: () => downloadA.promise,
    })
    const promiseB = runFilesMediaPreviewDownload({
      file: file('media-b'),
      requestSequence: requestB,
      isCurrentPreview: session.isCurrentPreview.bind(session),
      fetchDownload: () => downloadB.promise,
    })

    downloadA.reject(new Error('stale preview failure'))
    const resultA = await promiseA
    expect(resultA).toEqual({ status: 'stale' })

    downloadB.resolve(download('media-b'))
    const resultB = await promiseB
    expect(resultB).toEqual({
      status: 'success',
      preview: {
        id: 'media-b',
        name: 'media-b.jpg',
        mimeType: 'image/jpeg',
        url: 'https://example.com/media-b',
      },
    })
  })

  it('invalidates an in-flight preview when the session is explicitly closed', async () => {
    const session = createFilesMediaPreviewSession()
    const pendingDownload = deferred<DownloadURL>()
    const requestSequence = session.beginPreview()

    const promise = runFilesMediaPreviewDownload({
      file: file('media-a'),
      requestSequence,
      isCurrentPreview: session.isCurrentPreview.bind(session),
      fetchDownload: () => pendingDownload.promise,
    })

    session.invalidatePreview()
    pendingDownload.resolve(download('media-a'))

    const result = await promise
    expect(result).toEqual({ status: 'stale' })
  })
})
