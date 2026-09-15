/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { useResendCooldown } from './useResendCooldown'

const CooldownHost = defineComponent({
  setup() {
    return useResendCooldown(3)
  },
  template: '<button type="button" @click="start">{{ remainingSeconds }}|{{ active }}</button>',
})

describe('useResendCooldown', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('counts down to zero and becomes inactive again', async () => {
    const wrapper = mount(CooldownHost)

    await wrapper.get('button').trigger('click')
    expect(wrapper.text()).toBe('3|true')

    await vi.advanceTimersByTimeAsync(2000)
    expect(wrapper.text()).toBe('1|true')

    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.text()).toBe('0|false')
    expect(vi.getTimerCount()).toBe(0)

    wrapper.unmount()
  })

  it('restarts with one interval and clears it on unmount', async () => {
    const wrapper = mount(CooldownHost)
    const button = wrapper.get('button')

    await button.trigger('click')
    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.text()).toBe('2|true')

    await button.trigger('click')
    expect(wrapper.text()).toBe('3|true')
    expect(vi.getTimerCount()).toBe(1)

    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })
})
