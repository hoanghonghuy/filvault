/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import LinearProgress from './LinearProgress.vue'

describe('LinearProgress', () => {
  it('exposes a determinate progressbar with clamped percent', () => {
    const wrapper = mount(LinearProgress, {
      props: { value: 0.42, label: 'Uploading…' },
    })

    const bar = wrapper.get('[role="progressbar"]')
    expect(bar.attributes('aria-valuemin')).toBe('0')
    expect(bar.attributes('aria-valuemax')).toBe('100')
    expect(bar.attributes('aria-valuenow')).toBe('42')
    expect(bar.attributes('aria-label')).toBe('Uploading…')
    expect(wrapper.text()).toContain('42%')
    expect(wrapper.get('.linear-progress__fill').attributes('style')).toContain('width: 42%')
  })

  it('clamps out-of-range values to 0–100', () => {
    const over = mount(LinearProgress, { props: { value: 1.5 } })
    expect(over.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('100')

    const under = mount(LinearProgress, { props: { value: -0.2 } })
    expect(under.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('0')
  })
})
