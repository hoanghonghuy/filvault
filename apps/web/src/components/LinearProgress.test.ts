/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { setLocale } from '@/lib/i18n'
import LinearProgress from './LinearProgress.vue'

afterEach(() => {
  setLocale('vi')
})

describe('LinearProgress', () => {
  it('exposes a determinate progressbar with clamped percent and preserves explicit labels', () => {
    const wrapper = mount(LinearProgress, {
      props: { value: 0.42, label: 'Custom progress' },
    })

    const bar = wrapper.get('[role="progressbar"]')
    expect(bar.attributes('aria-valuemin')).toBe('0')
    expect(bar.attributes('aria-valuemax')).toBe('100')
    expect(bar.attributes('aria-valuenow')).toBe('42')
    expect(bar.attributes('aria-label')).toBe('Custom progress')
    expect(wrapper.text()).toContain('Custom progress')
    expect(wrapper.text()).toContain('42%')
    expect(wrapper.get('.linear-progress__fill').attributes('style')).toContain('width: 42%')
  })

  it('reactively localizes the default visible and accessible label', async () => {
    setLocale('vi')
    const wrapper = mount(LinearProgress, { props: { value: 0.25 } })

    expect(wrapper.get('.linear-progress__label').text()).toBe('Đang tải lên…')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('Đang tải lên…')

    setLocale('en')
    await nextTick()

    expect(wrapper.get('.linear-progress__label').text()).toBe('Uploading…')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('Uploading…')
  })

  it('keeps an explicitly empty label instead of replacing it with the localized default', () => {
    const wrapper = mount(LinearProgress, { props: { value: 0.1, label: '' } })

    expect(wrapper.get('.linear-progress__label').text()).toBe('')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('')
  })

  it('clamps out-of-range values to 0–100', () => {
    const over = mount(LinearProgress, { props: { value: 1.5 } })
    expect(over.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('100')

    const under = mount(LinearProgress, { props: { value: -0.2 } })
    expect(under.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('0')
  })
})
