/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/lib/i18n'
import BottomSheet from './BottomSheet.vue'

afterEach(() => {
  setLocale('vi')
  document.body.innerHTML = ''
  document.body.style.overflow = ''
})

describe('BottomSheet', () => {
  it('animates open and close with a named Vue Transition', () => {
    const wrapper = mount(BottomSheet, {
      props: { open: true, title: 'Move' },
      slots: { default: 'Body' },
      attachTo: document.body,
    })
    const transition = wrapper.findComponent({ name: 'Transition' })
    expect(transition.exists()).toBe(true)
    expect(transition.props('name')).toBe('sheet')
    wrapper.unmount()
  })

  it('reactively localizes the fallback accessible name for untitled sheets', async () => {
    setLocale('vi')
    const wrapper = mount(BottomSheet, {
      props: { open: true },
      attachTo: document.body,
    })

    let dialog = document.body.querySelector<HTMLElement>('[role="dialog"]')
    expect(dialog?.getAttribute('aria-label')).toBe('Hộp thoại')
    expect(dialog?.hasAttribute('aria-labelledby')).toBe(false)

    setLocale('en')
    await nextTick()

    dialog = document.body.querySelector<HTMLElement>('[role="dialog"]')
    expect(dialog?.getAttribute('aria-label')).toBe('Dialog')

    wrapper.unmount()
  })

  it('keeps visible titles authoritative through aria-labelledby', () => {
    setLocale('en')
    const wrapper = mount(BottomSheet, {
      props: { open: true, title: 'Move to' },
      attachTo: document.body,
    })

    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]')
    const heading = document.body.querySelector<HTMLElement>('.sheet-title')

    expect(dialog?.hasAttribute('aria-label')).toBe(false)
    expect(dialog?.getAttribute('aria-labelledby')).toBe(heading?.id)
    expect(heading?.textContent).toBe('Move to')

    wrapper.unmount()
  })
})
