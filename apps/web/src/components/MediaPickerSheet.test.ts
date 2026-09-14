/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/lib/i18n'
import MediaPickerSheet from './MediaPickerSheet.vue'

const apiMock = vi.hoisted(() => vi.fn<(path: string) => Promise<unknown>>())

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    api: apiMock,
  }
})

beforeEach(() => {
  apiMock.mockReset()
})

afterEach(() => {
  setLocale('vi')
})

const globalStubs = {
  BottomSheet: {
    props: ['open', 'title'],
    template: '<section><h1 class="sheet-title">{{ title }}</h1><slot /></section>',
  },
  EmptyState: {
    props: ['title', 'description'],
    template: '<div class="empty-state"><h2>{{ title }}</h2><p>{{ description }}</p></div>',
  },
  PhotoThumb: {
    props: ['name'],
    template: '<button class="photo-thumb">{{ name }}</button>',
  },
}

function timeline(name: string, id = name, nextBefore?: string) {
  return {
    groups: [
      {
        date: '2026-09-14',
        items: [
          {
            id,
            name,
            mimeType: 'image/jpeg',
            sizeBytes: 123,
            createdAt: '2026-09-14T08:00:00Z',
          },
        ],
      },
    ],
    nextBefore,
  }
}

describe('MediaPickerSheet', () => {
  it('reactively updates mounted picker chrome and empty state with the locale', async () => {
    setLocale('vi')
    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: { stubs: globalStubs },
    })

    expect(wrapper.find('.sheet-title').text()).toBe('Thêm vào album')
    expect(wrapper.find('.hint').text()).toBe('Chạm vào ảnh hoặc video để thêm.')
    expect(wrapper.find('.empty-state h2').text()).toBe('Không có ảnh hoặc video khả dụng')
    expect(wrapper.find('.empty-state p').text()).toBe('Hãy tải ảnh hoặc video lên Tệp của tôi trước.')

    setLocale('en')
    await nextTick()

    expect(wrapper.find('.sheet-title').text()).toBe('Add to album')
    expect(wrapper.find('.hint').text()).toBe('Tap a photo or video to add it.')
    expect(wrapper.find('.empty-state h2').text()).toBe('No media available')
    expect(wrapper.find('.empty-state p').text()).toBe('Upload photos or videos in My Files first.')

    wrapper.unmount()
  })

  it('shows a distinct retryable error state and recovers after a successful retry', async () => {
    setLocale('en')
    apiMock.mockRejectedValueOnce(new Error('Network unavailable')).mockResolvedValueOnce(timeline('photo.jpg', 'media-1'))

    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: { stubs: globalStubs },
    })
    await wrapper.setProps({ open: true })
    await flushPromises()

    expect(apiMock).toHaveBeenCalledTimes(1)
    expect(wrapper.find('.picker-error').exists()).toBe(true)
    expect(wrapper.find('.empty-state').exists()).toBe(false)
    expect(wrapper.find('.picker-error button').text()).toBe('Retry')

    await wrapper.find('.picker-error button').trigger('click')
    await flushPromises()

    expect(apiMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('.picker-error').exists()).toBe(false)
    expect(wrapper.find('.empty-state').exists()).toBe(false)
    expect(wrapper.find('.photo-thumb').text()).toBe('photo.jpg')

    wrapper.unmount()
  })

  it('ignores stale initial success after close and reopen', async () => {
    setLocale('en')
    let resolveOld: ((value: unknown) => void) | undefined
    let resolveCurrent: ((value: unknown) => void) | undefined
    apiMock
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveOld = resolve
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveCurrent = resolve
          }),
      )

    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: { stubs: globalStubs },
    })

    await wrapper.setProps({ open: true })
    await nextTick()
    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })
    await nextTick()

    resolveOld?.(timeline('old.jpg', 'old'))
    await flushPromises()

    expect(wrapper.find('.photo-thumb').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('old.jpg')
    expect(wrapper.text()).toContain('Loading')

    resolveCurrent?.(timeline('current.jpg', 'current'))
    await flushPromises()

    expect(wrapper.find('.photo-thumb').text()).toBe('current.jpg')
    expect(wrapper.text()).not.toContain('old.jpg')

    wrapper.unmount()
  })

  it('ignores stale initial failure and finally after close and reopen', async () => {
    setLocale('en')
    let rejectOld: ((reason?: unknown) => void) | undefined
    let resolveCurrent: ((value: unknown) => void) | undefined
    apiMock
      .mockImplementationOnce(
        () =>
          new Promise((_resolve, reject) => {
            rejectOld = reject
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveCurrent = resolve
          }),
      )

    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: { stubs: globalStubs },
    })

    await wrapper.setProps({ open: true })
    await nextTick()
    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })
    await nextTick()

    rejectOld?.(new Error('old request failed'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Loading')
    expect(wrapper.find('.empty-state').exists()).toBe(false)

    resolveCurrent?.(timeline('current.jpg', 'current'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.find('.photo-thumb').text()).toBe('current.jpg')

    wrapper.unmount()
  })

  it('ignores stale pagination success after a new session starts', async () => {
    setLocale('en')
    let resolveOldMore: ((value: unknown) => void) | undefined
    let resolveCurrent: ((value: unknown) => void) | undefined
    apiMock
      .mockResolvedValueOnce(timeline('first.jpg', 'first', 'cursor-1'))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveOldMore = resolve
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveCurrent = resolve
          }),
      )

    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: { stubs: globalStubs },
    })

    await wrapper.setProps({ open: true })
    await flushPromises()
    await wrapper.find('button.btn.block').trigger('click')
    await nextTick()
    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })
    await nextTick()

    resolveOldMore?.(timeline('stale-more.jpg', 'stale-more'))
    await flushPromises()

    expect(wrapper.find('.photo-thumb').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('stale-more.jpg')
    expect(wrapper.text()).toContain('Loading')

    resolveCurrent?.(timeline('current.jpg', 'current'))
    await flushPromises()

    expect(wrapper.find('.photo-thumb').text()).toBe('current.jpg')
    expect(wrapper.text()).not.toContain('stale-more.jpg')

    wrapper.unmount()
  })

  it('ignores stale pagination failure and finally after a new session starts', async () => {
    setLocale('en')
    let rejectOldMore: ((reason?: unknown) => void) | undefined
    let resolveCurrent: ((value: unknown) => void) | undefined
    apiMock
      .mockResolvedValueOnce(timeline('first.jpg', 'first', 'cursor-1'))
      .mockImplementationOnce(
        () =>
          new Promise((_resolve, reject) => {
            rejectOldMore = reject
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveCurrent = resolve
          }),
      )

    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: { stubs: globalStubs },
    })

    await wrapper.setProps({ open: true })
    await flushPromises()
    await wrapper.find('button.btn.block').trigger('click')
    await nextTick()
    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })
    await nextTick()

    rejectOldMore?.(new Error('old pagination failed'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Loading')
    expect(wrapper.find('.empty-state').exists()).toBe(false)

    resolveCurrent?.(timeline('current.jpg', 'current'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.find('.photo-thumb').text()).toBe('current.jpg')

    wrapper.unmount()
  })
})
