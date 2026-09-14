/**
 * @vitest-environment jsdom
 */
import { nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/lib/i18n'
import FolderPickerSheet from './FolderPickerSheet.vue'

const { apiMock } = vi.hoisted(() => ({ apiMock: vi.fn<(path: string) => Promise<unknown>>() }))

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    api: apiMock,
  }
})

beforeEach(() => {
  apiMock.mockReset()
  apiMock.mockResolvedValue({ folder: null, breadcrumb: [], folders: [] })
})

afterEach(() => {
  setLocale('vi')
})

const BottomSheetStub = {
  props: ['open', 'title'],
  template: '<section><h1>{{ title }}</h1><slot /></section>',
}

describe('FolderPickerSheet', () => {
  it('reactively localizes folder navigation chrome and default confirm action', async () => {
    setLocale('vi')
    const wrapper = mount(FolderPickerSheet, {
      props: { open: false, title: 'Move' },
      global: { stubs: { BottomSheet: BottomSheetStub } },
    })

    await wrapper.setProps({ open: true })
    await flushPromises()

    expect(wrapper.find('.picker-breadcrumb').attributes('aria-label')).toBe('Duyệt thư mục')
    expect(wrapper.find('.crumb').text()).toBe('Gốc')
    expect(wrapper.find('.picker-empty').text()).toBe('Không có thư mục con.')
    expect(wrapper.find('.picker-confirm').text()).toBe('Di chuyển tới đây')

    setLocale('en')
    await nextTick()

    expect(wrapper.find('.picker-breadcrumb').attributes('aria-label')).toBe('Browse folders')
    expect(wrapper.find('.crumb').text()).toBe('Root')
    expect(wrapper.find('.picker-empty').text()).toBe('No subfolders here.')
    expect(wrapper.find('.picker-confirm').text()).toBe('Move here')

    wrapper.unmount()
  })

  it('keeps an explicit caller confirm label authoritative', () => {
    const wrapper = mount(FolderPickerSheet, {
      props: { open: false, title: 'Copy', confirmLabel: 'Copy here' },
      global: { stubs: { BottomSheet: BottomSheetStub } },
    })

    expect(wrapper.find('.picker-confirm').text()).toBe('Copy here')
    wrapper.unmount()
  })

  it('blocks stale selection during loading and exposes localized retry after failure', async () => {
    setLocale('en')
    let rejectLoad: ((reason?: unknown) => void) | undefined
    apiMock.mockImplementationOnce(
      () =>
        new Promise((_resolve, reject) => {
          rejectLoad = reject
        }),
    )

    const wrapper = mount(FolderPickerSheet, {
      props: { open: false, title: 'Move' },
      global: { stubs: { BottomSheet: BottomSheetStub } },
    })

    await wrapper.setProps({ open: true })
    await nextTick()

    expect(wrapper.find('.picker-confirm').attributes('disabled')).toBeDefined()
    expect(wrapper.find('.picker-list').exists()).toBe(false)

    rejectLoad?.(new Error('network down'))
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    expect(wrapper.find('.picker-list').exists()).toBe(false)
    expect(wrapper.find('.picker-confirm').attributes('disabled')).toBeDefined()
    expect(wrapper.find('.picker-retry').text()).toBe('Retry')

    apiMock.mockResolvedValueOnce({ folder: null, breadcrumb: [], folders: [] })
    await wrapper.find('.picker-retry').trigger('click')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.find('.picker-list').exists()).toBe(true)
    expect(wrapper.find('.picker-confirm').attributes('disabled')).toBeUndefined()

    wrapper.unmount()
  })
})
