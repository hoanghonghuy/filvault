/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/lib/i18n'
import PhotoPlaceholder from './PhotoPlaceholder.vue'

afterEach(() => {
  setLocale('vi')
})

describe('PhotoPlaceholder', () => {
  it('reactively localizes the photo badge while preserving concise video wording', async () => {
    setLocale('vi')
    const image = mount(PhotoPlaceholder, { props: { mimeType: 'image/jpeg', name: 'photo.jpg' } })
    const video = mount(PhotoPlaceholder, { props: { mimeType: 'video/mp4', name: 'clip.mp4' } })

    expect(image.find('.badge').text()).toBe('Ảnh')
    expect(video.find('.badge').text()).toBe('Video')

    setLocale('en')
    await nextTick()

    expect(image.find('.badge').text()).toBe('Photo')
    expect(video.find('.badge').text()).toBe('Video')

    image.unmount()
    video.unmount()
  })

  it('includes the localized media type in the accessible name while preserving filename identity and click behavior', async () => {
    setLocale('vi')
    const wrapper = mount(PhotoPlaceholder, { props: { mimeType: 'image/png', name: 'cover.png' } })

    expect(wrapper.attributes('aria-label')).toBe('Ảnh: cover.png')
    expect(wrapper.attributes('title')).toBe('cover.png')

    setLocale('en')
    await nextTick()
    expect(wrapper.attributes('aria-label')).toBe('Photo: cover.png')

    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)

    wrapper.unmount()
  })
})
