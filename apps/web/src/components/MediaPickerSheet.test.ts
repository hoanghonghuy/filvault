/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/lib/i18n'
import MediaPickerSheet from './MediaPickerSheet.vue'

const apiMock = vi.hoisted(() => vi.fn<(path: string) => Promise<unknown>>())

vi.mock('@/api/client', () => ({
  api: apiMock,
}))

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
    apiMock
      .mockRejectedValueOnce(new Error('Network unavailable'))
      .mockResolvedValueOnce({
        groups: [
          {
            date: '2026-09-14',
            items: [
              {
                id: 'media-1',
                name: 'photo.jpg',
                mimeType: 'image/jpeg',
                sizeBytes: 123,
                createdAt: '2026-09-14T08:00:00Z',
              },
            ],
          },
        ],
        nextBefore: undefined,
      })

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
})
