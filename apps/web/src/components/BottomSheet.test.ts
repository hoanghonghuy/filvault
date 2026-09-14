/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BottomSheet from './BottomSheet.vue'

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
})
