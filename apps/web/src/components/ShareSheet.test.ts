/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { setLocale } from '@/lib/i18n'
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
  beforeEach(() => {
    setLocale('en')
  })

  it('uses a truthful copy icon for the copy-link action', () => {
    const wrapper = mountSheet({ url: '/s/token', expiresAt: null, createdAt: new Date().toISOString() })

    const copy = wrapper.get('button[aria-label="Copy link"]')
    expect(copy.find('[data-icon="copy"]').exists()).toBe(true)
    expect(copy.find('[data-icon="file"]').exists()).toBe(false)

    wrapper.unmount()
  })

  it('reacts to Vietnamese locale in create and existing-link states', async () => {
    const createWrapper = mountSheet()
    expect(createWrapper.get('[role="radiogroup"]').attributes('aria-label')).toBe('Link expiry')
    expect(createWrapper.text()).toContain('Forever')

    setLocale('vi')
    await createWrapper.vm.$nextTick()
    expect(createWrapper.get('[role="radiogroup"]').attributes('aria-label')).toBe('Thời hạn liên kết')
    expect(createWrapper.text()).toContain('Vĩnh viễn')
    expect(createWrapper.text()).toContain('Tạo liên kết')
    createWrapper.unmount()

    const existingWrapper = mountSheet({ url: '/s/token', expiresAt: null, createdAt: new Date().toISOString() })
    expect(existingWrapper.text()).toContain('Bất kỳ ai có liên kết này đều có thể xem và tải xuống')
    expect(existingWrapper.text()).toContain('Không hết hạn')
    expect(existingWrapper.get('button[aria-label="Sao chép liên kết"]').exists()).toBe(true)
    existingWrapper.unmount()
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
