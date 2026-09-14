/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick, reactive } from 'vue'
import type { AlbumDetail } from '@/api/types'
import AlbumView from './AlbumView.vue'

const { apiMock, routerPushMock } = vi.hoisted(() => ({
  apiMock: vi.fn<(path: string, options?: unknown) => Promise<unknown>>(),
  routerPushMock: vi.fn(),
}))
const routeMock = reactive({ params: { id: 'album-a' } as { id: string } })

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return { ...actual, api: apiMock }
})

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ push: routerPushMock }),
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

function album(id: string, itemIds: string[] = []): AlbumDetail {
  return {
    id,
    name: `Album ${id}`,
    itemCount: itemIds.length,
    createdAt: '2026-01-01T00:00:00.000Z',
    updatedAt: '2026-01-01T00:00:00.000Z',
    items: itemIds.map((itemId) => ({
      id: itemId,
      name: `${itemId}.jpg`,
      mimeType: 'image/jpeg',
      sizeBytes: 100,
      createdAt: '2026-01-01T00:00:00.000Z',
    })),
  }
}

function mountView() {
  return mount(AlbumView, {
    global: {
      plugins: [createPinia()],
      stubs: {
        RouterLink: { template: '<a><slot /></a>' },
        AppIcon: true,
        EmptyState: { template: '<div data-test="empty-state" />' },
        LoadingSkeletonAlbum: { template: '<div data-test="album-loading" />' },
        MediaPickerSheet: true,
        PhotoThumb: {
          props: ['name'],
          emits: ['click', 'contextmenu'],
          template: '<button type="button" data-test="photo-thumb" @click="$emit(\'click\')">{{ name }}</button>',
        },
        MediaLightbox: {
          props: ['open', 'name', 'url'],
          emits: ['next', 'prev', 'download', 'close'],
          template: '<div data-test="lightbox" :data-open="String(open)" :data-name="name" :data-url="url"><button data-test="close-lightbox" @click="$emit(\'close\')" /></div>',
        },
      },
    },
  })
}

describe('AlbumView request ownership', () => {
  beforeEach(() => {
    routeMock.params.id = 'album-a'
    apiMock.mockReset()
    routerPushMock.mockReset()
  })

  it('keeps the newer route load authoritative across stale failure, finally and success', async () => {
    const albumA = deferred<AlbumDetail>()
    const albumB = deferred<AlbumDetail>()
    apiMock.mockImplementation((path) => {
      if (path === '/photos/albums/album-a') return albumA.promise
      if (path === '/photos/albums/album-b') return albumB.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const wrapper = mountView()
    await nextTick()

    routeMock.params.id = 'album-b'
    await nextTick()
    expect(wrapper.find('[data-test="album-loading"]').exists()).toBe(true)

    albumA.reject(new Error('stale album failure'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="album-loading"]').exists()).toBe(true)

    albumB.resolve(album('album-b'))
    await flushPromises()

    expect(wrapper.text()).toContain('Album album-b')
    expect(wrapper.find('[data-test="album-loading"]').exists()).toBe(false)

    const staleSuccess = deferred<AlbumDetail>()
    const currentReload = deferred<AlbumDetail>()
    apiMock.mockImplementation((path) => {
      if (path === '/photos/albums/album-a') return staleSuccess.promise
      if (path === '/photos/albums/album-b') return currentReload.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    routeMock.params.id = 'album-a'
    await nextTick()
    routeMock.params.id = 'album-b'
    await nextTick()

    currentReload.resolve(album('album-b'))
    await flushPromises()
    staleSuccess.resolve(album('album-a'))
    await flushPromises()

    expect(wrapper.text()).toContain('Album album-b')
    expect(wrapper.text()).not.toContain('Album album-a')
  })

  it('keeps the latest preview authoritative and ignores stale preview failure', async () => {
    const previewA = deferred<{ downloadUrl: string; expiresAt: string }>()
    const previewB = deferred<{ downloadUrl: string; expiresAt: string }>()
    apiMock.mockImplementation((path) => {
      if (path === '/photos/albums/album-a') return Promise.resolve(album('album-a', ['media-a', 'media-b']))
      if (path === '/files/media-a/download') return previewA.promise
      if (path === '/files/media-b/download') return previewB.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const wrapper = mountView()
    await flushPromises()

    const thumbs = wrapper.findAll('[data-test="photo-thumb"]')
    await thumbs[0]!.trigger('click')
    await thumbs[1]!.trigger('click')

    previewA.reject(new Error('stale preview failure'))
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)

    previewB.resolve({ downloadUrl: 'https://example.com/media-b', expiresAt: '2026-01-01T00:05:00.000Z' })
    await flushPromises()

    const lightbox = wrapper.find('[data-test="lightbox"]')
    expect(lightbox.attributes('data-open')).toBe('true')
    expect(lightbox.attributes('data-name')).toBe('media-b.jpg')
    expect(lightbox.attributes('data-url')).toBe('https://example.com/media-b')
  })

  it('invalidates an in-flight preview when the lightbox is closed', async () => {
    const preview = deferred<{ downloadUrl: string; expiresAt: string }>()
    apiMock.mockImplementation((path) => {
      if (path === '/photos/albums/album-a') return Promise.resolve(album('album-a', ['media-a']))
      if (path === '/files/media-a/download') return preview.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-test="photo-thumb"]').trigger('click')
    await wrapper.find('[data-test="close-lightbox"]').trigger('click')

    preview.resolve({ downloadUrl: 'https://example.com/stale', expiresAt: '2026-01-01T00:05:00.000Z' })
    await flushPromises()

    const lightbox = wrapper.find('[data-test="lightbox"]')
    expect(lightbox.attributes('data-open')).toBe('false')
    expect(lightbox.attributes('data-name')).toBe('')
    expect(lightbox.attributes('data-url')).toBe('')
  })
})
