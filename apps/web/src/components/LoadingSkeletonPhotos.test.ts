/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/lib/i18n'
import LoadingSkeletonPhotos from './LoadingSkeletonPhotos.vue'

afterEach(() => {
  setLocale('vi')
})

describe('LoadingSkeletonPhotos', () => {
  it('reactively localizes the initial Photos loading chrome', async () => {
    setLocale('vi')
    const wrapper = mount(LoadingSkeletonPhotos)

    expect(wrapper.find('.sk-title').text()).toBe('Ảnh')
    expect(wrapper.find('.sk-subtitle').text()).toBe('Đang tải…')
    expect(wrapper.find('.section-title').text()).toBe('Album')

    setLocale('en')
    await nextTick()

    expect(wrapper.find('.sk-title').text()).toBe('Photos')
    expect(wrapper.find('.sk-subtitle').text()).toBe('Loading…')
    expect(wrapper.find('.section-title').text()).toBe('Albums')

    wrapper.unmount()
  })

  it('keeps the more variant focused on incremental skeleton rows', () => {
    const wrapper = mount(LoadingSkeletonPhotos, { props: { variant: 'more' } })

    expect(wrapper.find('.sk-title').exists()).toBe(false)
    expect(wrapper.find('.sk-subtitle').exists()).toBe(false)
    expect(wrapper.find('.toolbar').exists()).toBe(false)
    expect(wrapper.attributes('aria-busy')).toBe('true')
    expect(wrapper.attributes('aria-live')).toBe('polite')

    wrapper.unmount()
  })
})
