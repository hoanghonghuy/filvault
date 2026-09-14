import { mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ChatAttachment } from '@/api/types'
import ChatInlineImage from './ChatInlineImage.vue'

const getHeicDisplayUrlMock = vi.hoisted(() => vi.fn<(source: string) => Promise<string>>())

vi.mock('@/lib/heic', () => ({
  isHeic: (name?: string | null, mime?: string | null) =>
    Boolean(name?.toLowerCase().endsWith('.heic') || mime?.toLowerCase() === 'image/heic'),
  getHeicDisplayUrl: getHeicDisplayUrlMock,
}))

vi.mock('@/api/client', () => ({
  normalizePresignedUrl: (url: string) => url,
}))

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({ locale: ref<'vi' | 'en'>('en') }),
}))

vi.mock('@/lib/chatAttachmentCopy', () => ({
  chatAttachmentCopy: () => ({ viewImageAria: (name: string) => `View image: ${name}` }),
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

function attachment(id: string, name: string, thumbnailUrl: string, mimeType = 'image/heic') {
  return { id, name, thumbnailUrl, mimeType } as unknown as ChatAttachment
}

describe('ChatInlineImage HEIC resolution', () => {
  beforeEach(() => {
    getHeicDisplayUrlMock.mockReset()
  })

  it('ignores a stale HEIC conversion after the attachment changes', async () => {
    const first = deferred<string>()
    const second = deferred<string>()
    getHeicDisplayUrlMock.mockImplementationOnce(() => first.promise).mockImplementationOnce(() => second.promise)

    const wrapper = mount(ChatInlineImage, {
      props: { attachment: attachment('a', 'first.heic', 'https://files/first.heic') },
    })
    await nextTick()
    expect(wrapper.find('.heic-loading-box').exists()).toBe(true)

    await wrapper.setProps({ attachment: attachment('b', 'second.heic', 'https://files/second.heic') })
    expect(wrapper.find('.heic-loading-box').exists()).toBe(true)

    first.resolve('blob:first')
    await first.promise
    await nextTick()
    expect(wrapper.find('.heic-loading-box').exists()).toBe(true)
    expect(wrapper.find('img').exists()).toBe(false)

    second.resolve('blob:second')
    await second.promise
    await nextTick()
    expect(wrapper.find('.heic-loading-box').exists()).toBe(false)
    expect(wrapper.find('img').attributes('src')).toBe('blob:second')
  })

  it('clears HEIC display state when switching to a non-HEIC attachment', async () => {
    getHeicDisplayUrlMock.mockResolvedValueOnce('blob:first')
    const wrapper = mount(ChatInlineImage, {
      props: { attachment: attachment('a', 'first.heic', 'https://files/first.heic') },
    })
    await Promise.resolve()
    await nextTick()
    expect(wrapper.find('img').attributes('src')).toBe('blob:first')

    await wrapper.setProps({
      attachment: attachment('b', 'second.jpg', 'https://files/second.jpg', 'image/jpeg'),
    })
    await nextTick()
    expect(wrapper.find('.heic-loading-box').exists()).toBe(false)
    expect(wrapper.find('img').attributes('src')).toBe('https://files/second.jpg')
  })
})
