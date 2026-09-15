/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import type { DownloadURL } from '@/api/types'
import PhotoPreviewView from './PhotoPreviewView.vue'

interface OwnedFile {
  id: string
  name: string
  mimeType: string
  sizeBytes: number
  status: string
  createdAt: string
}

const { apiMock, routerReplaceMock } = vi.hoisted(() => ({
  apiMock: vi.fn<(path: string, options?: unknown) => Promise<unknown>>(),
  routerReplaceMock: vi.fn<(to: unknown) => void>(),
}))
const routeMock = reactive({ params: { id: 'media-a' } as { id: string } })

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return { ...actual, api: apiMock }
})

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ replace: routerReplaceMock }),
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

function media(id: string): OwnedFile {
  return {
    id,
    name: `${id}.jpg`,
    mimeType: 'image/jpeg',
    sizeBytes: 100,
    status: 'READY',
    createdAt: '2026-01-01T00:00:00.000Z',
  }
}

function download(id: string): DownloadURL {
  return {
    downloadUrl: `https://example.com/${id}`,
    expiresAt: '2026-01-01T00:05:00.000Z',
  }
}

function mountView() {
  return mount(PhotoPreviewView, {
    global: {
      stubs: {
        RouterLink: { template: '<a><slot /></a>' },
        AppIcon: true,
        MediaLightbox: {
          props: ['open', 'name', 'mimeType', 'url'],
          emits: ['download', 'close'],
          template:
            '<div data-test="lightbox" :data-open="String(open)" :data-name="name" :data-url="url"><button data-test="close-lightbox" @click="$emit(\'close\')" /></div>',
        },
      },
    },
  })
}

describe('PhotoPreviewView request ownership', () => {
  beforeEach(() => {
    routeMock.params.id = 'media-a'
    apiMock.mockReset()
    routerReplaceMock.mockReset()
  })

  it('reloads a reused preview route and keeps the newer download authoritative', async () => {
    const downloadA = deferred<DownloadURL>()
    const downloadB = deferred<DownloadURL>()
    apiMock.mockImplementation((path) => {
      if (path === '/files/media-a') return Promise.resolve(media('media-a'))
      if (path === '/files/media-a/download') return downloadA.promise
      if (path === '/files/media-b') return Promise.resolve(media('media-b'))
      if (path === '/files/media-b/download') return downloadB.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const wrapper = mountView()
    await flushPromises()

    routeMock.params.id = 'media-b'
    await nextTick()
    await flushPromises()

    downloadB.resolve(download('media-b'))
    await flushPromises()

    let lightbox = wrapper.find('[data-test="lightbox"]')
    expect(lightbox.attributes('data-open')).toBe('true')
    expect(lightbox.attributes('data-name')).toBe('media-b.jpg')
    expect(lightbox.attributes('data-url')).toBe('https://example.com/media-b')

    downloadA.resolve(download('media-a'))
    await flushPromises()

    lightbox = wrapper.find('[data-test="lightbox"]')
    expect(lightbox.attributes('data-open')).toBe('true')
    expect(lightbox.attributes('data-name')).toBe('media-b.jpg')
    expect(lightbox.attributes('data-url')).toBe('https://example.com/media-b')
  })

  it('ignores stale failure and finally while the newer route is still loading', async () => {
    const fileA = deferred<OwnedFile>()
    const fileB = deferred<OwnedFile>()
    const downloadB = deferred<DownloadURL>()
    apiMock.mockImplementation((path) => {
      if (path === '/files/media-a') return fileA.promise
      if (path === '/files/media-b') return fileB.promise
      if (path === '/files/media-b/download') return downloadB.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const wrapper = mountView()
    await nextTick()

    routeMock.params.id = 'media-b'
    await nextTick()

    fileA.reject(new Error('stale preview failure'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.find('[aria-busy="true"]').exists()).toBe(true)

    fileB.resolve(media('media-b'))
    await flushPromises()
    downloadB.resolve(download('media-b'))
    await flushPromises()

    expect(wrapper.find('[aria-busy="true"]').exists()).toBe(false)
    const lightbox = wrapper.find('[data-test="lightbox"]')
    expect(lightbox.attributes('data-open')).toBe('true')
    expect(lightbox.attributes('data-name')).toBe('media-b.jpg')
  })

  it('invalidates an in-flight download when preview closes', async () => {
    const downloadA = deferred<DownloadURL>()
    apiMock.mockImplementation((path) => {
      if (path === '/files/media-a') return Promise.resolve(media('media-a'))
      if (path === '/files/media-a/download') return downloadA.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-test="close-lightbox"]').trigger('click')
    downloadA.resolve(download('media-a'))
    await flushPromises()

    expect(routerReplaceMock).toHaveBeenCalledWith('/photos')
    const lightbox = wrapper.find('[data-test="lightbox"]')
    expect(lightbox.attributes('data-open')).toBe('false')
    expect(lightbox.attributes('data-name')).toBe('')
    expect(lightbox.attributes('data-url')).toBe('')
  })
})
