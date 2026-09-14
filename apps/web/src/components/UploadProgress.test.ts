/**
 * @vitest-environment jsdom
 */
import { describe, expect, it, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UploadProgress from './UploadProgress.vue'
import { setLocale } from '@/lib/i18n'

const baseItems = [
  { id: '1', name: 'done.txt', status: 'completed' as const, progress: 1 },
  { id: '2', name: 'bad.txt', status: 'failed' as const, progress: 0, error: "Can't reach the server. Try again in a moment." },
  { id: '3', name: 'next.txt', status: 'queued' as const, progress: 0 },
]

describe('UploadProgress', () => {
  beforeEach(() => {
    setLocale('en')
  })

  it('renders aggregate progress without aria-live on the progress section', () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.5,
        items: baseItems,
      },
    })

    expect(wrapper.get('.upload-progress').attributes('aria-live')).toBeUndefined()
    const bar = wrapper.get('[role="progressbar"][aria-label="Overall upload progress"]')
    expect(bar.attributes('aria-valuenow')).toBe('50')
    expect(wrapper.get('[aria-label="Files uploading"]')).toBeTruthy()
  })

  it('preserves explicit labels, including an intentionally empty label', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        progress: 0.5,
        label: 'Custom upload label',
      },
    })

    expect(wrapper.get('.upload-progress-title span').text()).toBe('Custom upload label')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('Custom upload label')

    await wrapper.setProps({ label: '' })

    expect(wrapper.get('.upload-progress-title span').text()).toBe('')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('')
  })

  it('keeps omitted labels locale-reactive', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        progress: 0.5,
      },
    })

    expect(wrapper.get('.upload-progress-title span').text()).toBe('Uploading…')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('Overall upload progress')

    setLocale('vi')
    await flushPromises()

    expect(wrapper.get('.upload-progress-title span').text()).toBe('Đang tải lên…')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-label')).toBe('Tiến trình tải lên tổng')
  })

  it('exposes file-specific action labels and preserves action events', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.2,
        items: [
          { id: 'failed-1', name: 'bad.txt', status: 'failed', progress: 0, error: 'Upload failed' },
          { id: 'queued-1', name: 'wait.txt', status: 'queued', progress: 0 },
          { id: 'uploading-1', name: 'big.bin', status: 'uploading', progress: 0.2 },
        ],
      },
    })

    const failedButtons = wrapper.get('[data-status="failed"] .upload-item-actions').findAll('.upload-action-btn')
    expect(failedButtons[0]!.attributes('aria-label')).toBe('Retry: bad.txt')
    expect(failedButtons[1]!.attributes('aria-label')).toBe('Remove failed upload: bad.txt')
    expect(wrapper.get('[data-status="queued"] .upload-action-btn').attributes('aria-label')).toBe('Cancel: wait.txt')
    expect(wrapper.get('[data-status="uploading"] .upload-action-btn').attributes('aria-label')).toBe('Cancel: big.bin')

    await failedButtons[0]!.trigger('click')
    expect(wrapper.emitted('retry')?.[0]).toEqual(['failed-1'])

    await failedButtons[1]!.trigger('click')
    expect(wrapper.emitted('dismissFailed')?.[0]).toEqual(['failed-1'])

    await wrapper.get('[data-status="queued"] .upload-action-btn').trigger('click')
    expect(wrapper.emitted('cancel')?.[0]).toEqual(['queued-1'])

    setLocale('vi')
    await flushPromises()

    expect(failedButtons[0]!.attributes('aria-label')).toBe('Thử lại: bad.txt')
    expect(failedButtons[1]!.attributes('aria-label')).toBe('Xóa mục tải lên thất bại: bad.txt')
    expect(wrapper.get('[data-status="queued"] .upload-action-btn').attributes('aria-label')).toBe('Hủy: wait.txt')
  })

  it('shows clear-settled after active transfers finish', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.2,
        items: [
          { id: 'failed-1', name: 'bad.txt', status: 'failed', progress: 0, error: 'Upload failed' },
          { id: 'queued-1', name: 'wait.txt', status: 'queued', progress: 0 },
        ],
      },
    })

    expect(wrapper.find('.upload-dismiss-btn').exists()).toBe(false)

    await wrapper.setProps({
      items: [{ id: 'failed-1', name: 'bad.txt', status: 'failed', progress: 0, error: 'Upload failed' }],
    })
    await wrapper.get('.upload-dismiss-btn').trigger('click')
    expect(wrapper.emitted('dismiss')?.length).toBeGreaterThan(0)
  })

  it('surfaces resolved conflict names', () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.8,
        items: [
          {
            id: 'renamed',
            name: 'folder/foo.pdf',
            resolvedName: 'folder/foo (1).pdf',
            status: 'uploading',
            progress: 0.8,
          },
        ],
      },
    })

    expect(wrapper.text()).toContain('folder/foo.pdf')
    expect(wrapper.text()).toContain('Uploaded as folder/foo (1).pdf')
  })

  it('switches queue labels with locale', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.1,
        items: [{ id: 'q1', name: 'wait.txt', status: 'queued', progress: 0 }],
      },
    })

    expect(wrapper.text()).toContain('Waiting')
    expect(wrapper.text()).toContain('Cancel')

    setLocale('vi')
    await flushPromises()

    expect(wrapper.text()).toContain('Đang chờ')
    expect(wrapper.text()).toContain('Hủy')
  })

  it('shows clear-settled control when only failed items remain', () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.5,
        items: [{ id: 'failed-1', name: 'bad.txt', status: 'failed', progress: 0, error: 'Upload failed' }],
      },
    })

    expect(wrapper.text()).toContain('Clear')
  })

  it('announces status transitions in the live region without progress-only updates', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0,
        items: [{ id: 'u1', name: 'big.bin', status: 'uploading', progress: 0.1 }],
      },
    })

    const live = wrapper.get('.upload-live-status')
    expect(live.text()).toBe('')

    await wrapper.setProps({
      aggregateProgress: 0.45,
      items: [{ id: 'u1', name: 'big.bin', status: 'uploading', progress: 0.45 }],
    })
    await flushPromises()
    expect(live.text()).toBe('')

    await wrapper.setProps({
      aggregateProgress: 1,
      items: [{ id: 'u1', name: 'big.bin', status: 'completed', progress: 1 }],
    })
    await flushPromises()
    expect(live.text()).toContain('big.bin uploaded')
  })
})
