/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import ToastHost from './ToastHost.vue'
import { useUiStore } from '@/stores/ui'

describe('ToastHost', () => {
  it('animates with a named Vue Transition', async () => {
    const pinia = createPinia()
    const wrapper = mount(ToastHost, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })
    useUiStore(pinia).showToast('Saved')
    await wrapper.vm.$nextTick()
    const transition = wrapper.findComponent({ name: 'Transition' })
    expect(transition.exists()).toBe(true)
    expect(transition.props('name')).toBe('toast')
    wrapper.unmount()
  })
})
