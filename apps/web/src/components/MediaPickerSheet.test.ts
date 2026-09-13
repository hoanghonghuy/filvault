/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/lib/i18n'
import MediaPickerSheet from './MediaPickerSheet.vue'

afterEach(() => {
  setLocale('vi')
})

describe('MediaPickerSheet', () => {
  it('reactively updates mounted picker chrome and empty state with the locale', async () => {
    setLocale('vi')
    const wrapper = mount(MediaPickerSheet, {
      props: { open: false },
      global: {
        stubs: {
          BottomSheet: {
            props: ['open', 'title'],
            template: '<section><h1 class="sheet-title">{{ title }}</h1><slot /></section>',
          },
          EmptyState: {
            props: ['title', 'description'],
            template: '<div class="empty-state"><h2>{{ title }}</h2><p>{{ description }}</p></div>',
          },
          PhotoThumb: { template: '<button />' },
        },
      },
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
})
