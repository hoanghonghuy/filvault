/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ShareSheet from './ShareSheet.vue'

const BottomSheetStub = {
  props: ['open', 'title'],
  template: '<div><slot /></div>',
}

const IconStub = {
  props: ['name', 'size'],
  template: '<span :data-icon="name" />',
}

function mountSheet(existing: { url: string; expiresAt: string | null; createdAt: string } | null = null) {
  return mount(ShareSheet, {
    props: {
      open: true,
      name: 'example.txt',
      existing,
    },
    attachTo: document.body,
    global: {
      stubs: {
        BottomSheet: BottomSheetStub,
        Icon: IconStub,
      },
    },
  })
}

describe('ShareSheet', () => {
  it('uses a truthful copy icon for the copy-link action', () => {
    const wrapper = mountSheet({ url: '/s/token', expiresAt: null, createdAt: new Date().toISOString() })

    const copy = wrapper.get('button[aria-label="Copy link"]')
    expect(copy.find('[data-icon="copy"]').exists()).toBe(true)
    expect(copy.find('[data-icon="file"]').exists()).toBe(false)

    wrapper.unmount()
  })

  it('moves selection with arrows, wraps, and keeps roving tabindex in sync', async () => {
    const wrapper = mountSheet()
    const radios = wrapper.findAll<HTMLButtonElement>('[role="radio"]')

    expect(radios).toHaveLength(4)
    expect(radios[0]?.attributes('aria-checked')).toBe('true')
    expect(radios[0]?.attributes('tabindex')).toBe('0')

    await radios[0]?.trigger('keydown', { key: 'ArrowRight' })
    expect(radios[1]?.attributes('aria-checked')).toBe('true')
    expect(radios[1]?.attributes('tabindex')).toBe('0')
    expect(document.activeElement).toBe(radios[1]?.element)

    await radios[1]?.trigger('keydown', { key: 'ArrowLeft' })
    expect(radios[0]?.attributes('aria-checked')).toBe('true')

    await radios[0]?.trigger('keydown', { key: 'ArrowLeft' })
    expect(radios[3]?.attributes('aria-checked')).toBe('true')
    expect(document.activeElement).toBe(radios[3]?.element)

    wrapper.unmount()
  })

  it('supports Home/End and pointer selection', async () => {
    const wrapper = mountSheet()
    const radios = wrapper.findAll<HTMLButtonElement>('[role="radio"]')

    await radios[0]?.trigger('keydown', { key: 'End' })
    expect(radios[3]?.attributes('aria-checked')).toBe('true')
    expect(document.activeElement).toBe(radios[3]?.element)

    await radios[3]?.trigger('keydown', { key: 'Home' })
    expect(radios[0]?.attributes('aria-checked')).toBe('true')
    expect(document.activeElement).toBe(radios[0]?.element)

    await radios[2]?.trigger('click')
    expect(radios[2]?.attributes('aria-checked')).toBe('true')
    expect(radios[2]?.attributes('tabindex')).toBe('0')

    wrapper.unmount()
  })
})
