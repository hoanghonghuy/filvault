/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import BottomSheet from './BottomSheet.vue'

afterEach(() => {
  document.body.innerHTML = ''
  document.body.style.overflow = ''
})

describe('BottomSheet focus containment', () => {
  it('immediately recaptures programmatic focus outside the dialog', async () => {
    const behind = document.createElement('button')
    document.body.appendChild(behind)
    const wrapper = mount(BottomSheet, {
      props: { open: true, title: 'Move' },
      slots: { default: '<button id="inside">Inside</button>' },
      attachTo: document.body,
    })
    await nextTick()

    behind.focus()

    expect(document.activeElement).toBe(document.body.querySelector('#inside'))
    wrapper.unmount()
  })

  it('lets only the topmost stacked sheet recapture outside focus', async () => {
    const behind = document.createElement('button')
    document.body.appendChild(behind)
    const first = mount(BottomSheet, {
      props: { open: true, title: 'First' },
      slots: { default: '<button id="first-action">First</button>' },
      attachTo: document.body,
    })
    const second = mount(BottomSheet, {
      props: { open: true, title: 'Second' },
      slots: { default: '<button id="second-action">Second</button>' },
      attachTo: document.body,
    })
    await nextTick()

    behind.focus()

    expect(document.activeElement).toBe(document.body.querySelector('#second-action'))
    second.unmount()
    first.unmount()
  })

  it('falls back to the dialog panel when it has no focusable descendants', async () => {
    const behind = document.createElement('button')
    document.body.appendChild(behind)
    const wrapper = mount(BottomSheet, {
      props: { open: true, title: 'Info' },
      slots: { default: '<p>Nothing actionable</p>' },
      attachTo: document.body,
    })
    await nextTick()

    behind.focus()

    expect(document.activeElement).toBe(document.body.querySelector('[role="dialog"]'))
    wrapper.unmount()
  })
})
