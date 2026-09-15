/**
 * @vitest-environment jsdom
 */
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PhotoThumb from './PhotoThumb.vue'

const getHeicDisplayUrlMock = vi.hoisted(() => vi.fn<(source: string) => Promise<string>>())

vi.mock('@/lib/heic', () => ({
  isHeic: (name?: string | null, mime?: string | null) =>
    Boolean(name?.toLowerCase().endsWith('.heic') || mime?.toLowerCase() === 'image/heic'),
  getHeicDisplayUrl: getHeicDisplayUrlMock,
}))

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

describe('PhotoThumb media identity', () => {
  beforeEach(() => {
    getHeicDisplayUrlMock.mockReset()
  })

  it('ignores a stale HEIC conversion after media props change', async () => {
    const first = deferred<string>()
    const second = deferred<string>()
    getHeicDisplayUrlMock.mockImplementationOnce(() => first.promise).mockImplementationOnce(() => second.promise)

    const wrapper = mount(PhotoThumb, {
      props: {
        name: 'first.heic',
        mimeType: 'image/heic',
        thumbnailUrl: 'https://files/first.heic',
      },
      global: { stubs: { AppIcon: true } },
    })

    await wrapper.setProps({
      name: 'second.heic',
      mimeType: 'image/heic',
      thumbnailUrl: 'https://files/second.heic',
    })

    first.resolve('blob:first')
    await first.promise
    await nextTick()
    expect(wrapper.find('img').attributes('src')).toBe('https://files/second.heic')

    second.resolve('blob:second')
    await second.promise
    await nextTick()
    expect(wrapper.find('img').attributes('src')).toBe('blob:second')
  })

  it('clears HEIC display state when switching to non-HEIC media', async () => {
    getHeicDisplayUrlMock.mockResolvedValueOnce('blob:first')

    const wrapper = mount(PhotoThumb, {
      props: {
        name: 'first.heic',
        mimeType: 'image/heic',
        thumbnailUrl: 'https://files/first.heic',
      },
      global: { stubs: { AppIcon: true } },
    })
    await Promise.resolve()
    await nextTick()
    expect(wrapper.find('img').attributes('src')).toBe('blob:first')

    await wrapper.setProps({
      name: 'second.jpg',
      mimeType: 'image/jpeg',
      thumbnailUrl: 'https://files/second.jpg',
    })
    await nextTick()

    expect(wrapper.find('img').attributes('src')).toBe('https://files/second.jpg')
    expect(wrapper.find('.heic-badge').exists()).toBe(false)
  })

  it('resets a previous media load failure when new media is assigned', async () => {
    const wrapper = mount(PhotoThumb, {
      props: {
        name: 'broken.jpg',
        mimeType: 'image/jpeg',
        thumbnailUrl: 'https://files/broken.jpg',
      },
      global: { stubs: { AppIcon: true } },
    })

    await wrapper.find('img').trigger('error')
    expect(wrapper.find('.fallback').exists()).toBe(true)

    await wrapper.setProps({
      name: 'healthy.jpg',
      mimeType: 'image/jpeg',
      thumbnailUrl: 'https://files/healthy.jpg',
    })
    await nextTick()

    expect(wrapper.find('.fallback').exists()).toBe(false)
    expect(wrapper.find('img').attributes('src')).toBe('https://files/healthy.jpg')
  })
})
