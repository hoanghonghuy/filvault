/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UploadProgress from './UploadProgress.vue'

describe('UploadProgress', () => {
  it('renders aggregate and per-file progress with accessible labels', () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.5,
        items: [
          { id: '1', name: 'done.txt', status: 'completed', progress: 1 },
          { id: '2', name: 'bad.txt', status: 'failed', progress: 0, error: "Can't reach the server. Try again in a moment." },
          { id: '3', name: 'next.txt', status: 'queued', progress: 0 },
        ],
      },
    })

    const bar = wrapper.get('[role="progressbar"][aria-label="Tiến trình tải lên tổng"]')
    expect(bar.attributes('aria-valuenow')).toBe('50')
    expect(wrapper.text()).toContain('done.txt')
    expect(wrapper.text()).toContain('bad.txt')
    expect(wrapper.text()).toContain("Can't reach the server")
    expect(wrapper.get('[aria-label="Danh sách tệp đang tải lên"]')).toBeTruthy()
  })

  it('emits retry and cancel actions', async () => {
    const wrapper = mount(UploadProgress, {
      props: {
        aggregateProgress: 0.2,
        items: [
          { id: 'failed-1', name: 'bad.txt', status: 'failed', progress: 0, error: 'Upload failed' },
          { id: 'queued-1', name: 'wait.txt', status: 'queued', progress: 0 },
        ],
      },
    })

    await wrapper.get('[data-status="failed"] .upload-action-btn').trigger('click')
    expect(wrapper.emitted('retry')?.[0]).toEqual(['failed-1'])

    await wrapper.get('[data-status="queued"] .upload-action-btn').trigger('click')
    expect(wrapper.emitted('cancel')?.[0]).toEqual(['queued-1'])
  })
})
