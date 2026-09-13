/**
 * @vitest-environment jsdom
 */
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import { setLocale } from '@/lib/i18n'
import PublicShareView from './PublicShareView.vue'

const apiMock = vi.fn<(path: string) => Promise<unknown>>()

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { token: 'public-token' } }),
}))

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    api: (path: string) => apiMock(path),
  }
})

const meta = {
  name: 'quarterly-report.pdf',
  mimeType: 'application/pdf',
  sizeBytes: 1024,
  expiresAt: '2026-09-12T08:30:00.000Z',
}

async function mountView() {
  const wrapper = mount(PublicShareView, {
    global: {
      stubs: {
        Icon: true,
        MediaLightbox: {
          props: ['open', 'name', 'mimeType', 'url'],
          emits: ['close', 'download'],
          template:
            '<div v-if="open" class="media-lightbox-stub" :data-url="url" :data-mime="mimeType">{{ name }}</div>',
        },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('PublicShareView', () => {
  beforeEach(() => {
    setLocale('en')
    apiMock.mockReset()
  })

  afterEach(() => {
    setLocale('vi')
  })

  it('recovers from a transient metadata failure without a page reload', async () => {
    apiMock.mockRejectedValueOnce(new TypeError('Failed to fetch')).mockResolvedValueOnce(meta)
    const wrapper = await mountView()

    expect(wrapper.text()).toContain('could not be loaded right now')
    expect(wrapper.text()).toContain('Retry')

    await wrapper.get('.retry-btn').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('quarterly-report.pdf')
    expect(wrapper.find('.retry-btn').exists()).toBe(false)
    expect(apiMock).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('keeps a terminal not-found state distinct from transient failure', async () => {
    apiMock.mockRejectedValueOnce(new ApiError('NOT_FOUND', 'internal detail', 404))
    const wrapper = await mountView()

    expect(wrapper.text()).toContain('This link is not available')
    expect(wrapper.text()).not.toContain('internal detail')
    expect(wrapper.find('.retry-btn').exists()).toBe(false)
    wrapper.unmount()
  })

  it('keeps valid metadata visible when preparing a download fails', async () => {
    apiMock.mockResolvedValueOnce(meta).mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = await mountView()

    await wrapper.get('.download-btn').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('quarterly-report.pdf')
    expect(wrapper.text()).toContain('link is still available')
    expect(wrapper.text()).not.toContain('This link is not available')
    expect(wrapper.get('.download-btn').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('prevents duplicate download activation while a request is pending', async () => {
    let rejectDownload!: (reason?: unknown) => void
    apiMock.mockResolvedValueOnce(meta).mockImplementationOnce(
      () =>
        new Promise((_resolve, reject) => {
          rejectDownload = reject
        }),
    )
    const wrapper = await mountView()

    await wrapper.get('.download-btn').trigger('click')
    await Promise.resolve()
    const button = wrapper.get('.download-btn')
    expect(button.attributes('disabled')).toBeDefined()
    expect(button.attributes('aria-busy')).toBe('true')

    await button.trigger('click')
    expect(apiMock).toHaveBeenCalledTimes(2)

    rejectDownload(new TypeError('Failed to fetch'))
    await flushPromises()
    expect(wrapper.get('.download-btn').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('shows Preview only for supported media types and opens the shared lightbox lazily', async () => {
    apiMock
      .mockResolvedValueOnce(meta)
      .mockResolvedValueOnce({ downloadUrl: 'https://objects.example.test/report.pdf' })
    const wrapper = await mountView()

    expect(wrapper.find('.preview-btn').exists()).toBe(true)
    expect(wrapper.find('.media-lightbox-stub').exists()).toBe(false)
    expect(apiMock).toHaveBeenCalledTimes(1)

    await wrapper.get('.preview-btn').trigger('click')
    await flushPromises()

    const lightbox = wrapper.get('.media-lightbox-stub')
    expect(lightbox.attributes('data-url')).toBe('https://objects.example.test/report.pdf')
    expect(lightbox.attributes('data-mime')).toBe('application/pdf')
    expect(apiMock).toHaveBeenCalledWith('/public/shares/public-token/download')
    expect(apiMock).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('keeps unsupported public files download-first', async () => {
    apiMock.mockResolvedValueOnce({ ...meta, name: 'archive.zip', mimeType: 'application/zip' })
    const wrapper = await mountView()

    expect(wrapper.find('.preview-btn').exists()).toBe(false)
    expect(wrapper.find('.download-btn').exists()).toBe(true)
    wrapper.unmount()
  })

  it('recovers from preview preparation failure without invalidating share metadata', async () => {
    apiMock.mockResolvedValueOnce(meta).mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = await mountView()

    await wrapper.get('.preview-btn').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('quarterly-report.pdf')
    expect(wrapper.text()).toContain('Could not prepare the preview')
    expect(wrapper.text()).not.toContain('This link is not available')
    expect(wrapper.get('.preview-btn').attributes('disabled')).toBeUndefined()
    expect(wrapper.find('.media-lightbox-stub').exists()).toBe(false)
    wrapper.unmount()
  })

  it('reacts to locale changes and formats expiry using the active locale', async () => {
    apiMock.mockResolvedValueOnce(meta)
    const wrapper = await mountView()
    expect(wrapper.text()).toContain('Expires')
    expect(wrapper.text()).toContain('Preview')

    setLocale('vi')
    await flushPromises()

    expect(wrapper.text()).toContain('Hết hạn')
    expect(wrapper.text()).toContain('Xem trước')
    expect(wrapper.text()).toContain('Được chia sẻ qua Filvault')
    wrapper.unmount()
  })
})
