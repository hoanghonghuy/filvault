/**
 * @vitest-environment jsdom
 */
import { mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ChatBubblePickerModal from './ChatBubblePickerModal.vue'

vi.mock('@/lib/i18n', () => ({
  useI18n: () => ({ locale: ref('en') }),
}))

vi.mock('@/lib/chatBubblePickerCopy', () => ({
  chatBubblePickerCopy: () => ({
    cancel: 'Cancel',
    save: 'Save',
    title: 'Chat bubbles',
    preview: 'Preview',
    suggestions: 'Styles',
    help: 'Choose a style',
    stylesAria: 'Bubble styles',
    savedToast: () => 'Saved',
  }),
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({ showToast: vi.fn() }),
}))

const flush = async () => {
  await nextTick()
  await Promise.resolve()
}

afterEach(() => {
  document.body.innerHTML = ''
})

describe('ChatBubblePickerModal focus lifecycle', () => {
  it('contains keyboard and programmatic focus, closes on Escape, and restores the opener', async () => {
    const opener = document.createElement('button')
    opener.textContent = 'Open picker'
    document.body.append(opener)
    opener.focus()

    const wrapper = mount(ChatBubblePickerModal, {
      props: { open: false },
      attachTo: document.body,
    })

    await wrapper.setProps({ open: true })
    await flush()

    const dialog = document.querySelector<HTMLElement>('[role="dialog"]')!
    const cancel = dialog.querySelector<HTMLButtonElement>('.cancel-btn')!
    const buttons = dialog.querySelectorAll<HTMLButtonElement>('button')
    const last = buttons.item(buttons.length - 1)

    expect(document.activeElement).toBe(cancel)

    opener.focus()
    expect(document.activeElement).toBe(cancel)

    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(cancel)

    cancel.focus()
    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true }),
    )
    expect(document.activeElement).toBe(last)

    const escape = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true })
    document.dispatchEvent(escape)
    expect(escape.defaultPrevented).toBe(true)
    expect(wrapper.emitted('close')).toHaveLength(1)

    await wrapper.setProps({ open: false })
    await flush()
    expect(document.activeElement).toBe(opener)

    wrapper.unmount()
  })
})
