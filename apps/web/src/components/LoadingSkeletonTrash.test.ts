/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/lib/i18n'
import LoadingSkeletonTrash from './LoadingSkeletonTrash.vue'

afterEach(() => {
  setLocale('vi')
})

describe('LoadingSkeletonTrash', () => {
  it('reactively localizes the Trash loading heading', async () => {
    setLocale('vi')
    const wrapper = mount(LoadingSkeletonTrash)

    expect(wrapper.find('.page-title').text()).toBe('Thùng rác')
    expect(wrapper.attributes('aria-busy')).toBe('true')
    expect(wrapper.attributes('aria-live')).toBe('polite')

    setLocale('en')
    await nextTick()

    expect(wrapper.find('.page-title').text()).toBe('Trash')
    expect(wrapper.attributes('aria-busy')).toBe('true')
    expect(wrapper.attributes('aria-live')).toBe('polite')

    wrapper.unmount()
  })
})
