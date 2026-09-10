/**
 * @vitest-environment jsdom
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import MediaLightbox from './MediaLightbox.vue'
import { setLocale } from '@/lib/i18n'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

function mockMatchMedia(matchesByQuery: Record<string, boolean> = {}) {
  window.matchMedia = vi.fn<(query: string) => MediaQueryList>().mockImplementation((query: string) => ({
    matches: matchesByQuery[query] ?? false,
    media: query,
    addEventListener: vi.fn<(type: string, listener: EventListener) => void>(),
    removeEventListener: vi.fn<(type: string, listener: EventListener) => void>(),
  })) as typeof window.matchMedia
}

function mockDialogElement() {
  HTMLDialogElement.prototype.showModal = vi.fn<() => void>(function showModal(this: HTMLDialogElement) {
    this.setAttribute('open', '')
  })
  HTMLDialogElement.prototype.close = vi.fn<() => void>(function close(this: HTMLDialogElement) {
    this.removeAttribute('open')
  })
}

describe('MediaLightbox', () => {
  beforeEach(() => {
    setLocale('en')
    document.body.innerHTML = ''
    mockMatchMedia()
    mockDialogElement()
  })

  afterEach(() => {
    document.body.style.overflow = ''
    vi.restoreAllMocks()
  })

  it('localizes navigation and PDF preview labels', () => {
    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'vacation.heic',
        mimeType: 'image/heic',
        url: 'https://example.com/vacation.heic',
        hasPrev: true,
        hasNext: true,
      },
      attachTo: document.body,
    })

    expect(document.body.querySelector('button[aria-label="Previous item"]')).toBeTruthy()
    expect(document.body.querySelector('button[aria-label="Next item"]')).toBeTruthy()

    wrapper.unmount()
  })

  it('localizes the PDF preview iframe title', () => {
    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'report.pdf',
        mimeType: 'application/pdf',
        url: 'https://example.com/report.pdf',
      },
      attachTo: document.body,
    })

    const frame = document.body.querySelector('iframe.pdf-frame')
    expect(frame?.getAttribute('title')).toBe('PDF preview')
    wrapper.unmount()
  })

  it('keeps core toolbar actions discoverable without overflow at narrow widths', async () => {
    mockMatchMedia({ '(max-width: 639px)': true })

    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'a-very-long-filename-that-should-truncate-cleanly.heic',
        mimeType: 'image/heic',
        url: 'https://example.com/file.heic',
        hasPrev: true,
        hasNext: true,
      },
      attachTo: document.body,
    })

    await nextTick()

    expect(document.body.querySelector('.lightbox-title')).toBeTruthy()
    expect(document.body.querySelector('button[aria-label="Close preview"]')).toBeTruthy()
    expect(document.body.querySelector('button[aria-label="Download"]')).toBeTruthy()
    expect(document.body.querySelector('button[aria-label="Previous item"]')).toBeTruthy()
    expect(document.body.querySelector('button[aria-label="Next item"]')).toBeTruthy()
    expect(document.body.querySelector('.heic-overflow')).toBeTruthy()
    expect(document.body.querySelector('.heic-desktop')).toBeFalsy()

    wrapper.unmount()
  })

  it('does not inject a fake captions track when captions are unavailable', () => {
    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'clip.mp4',
        mimeType: 'video/mp4',
        url: 'https://example.com/clip.mp4',
      },
      attachTo: document.body,
    })

    expect(document.body.querySelector('track')).toBeFalsy()
    wrapper.unmount()
  })

  it('renders a captions track only when real caption data is provided', () => {
    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'clip.mp4',
        mimeType: 'video/mp4',
        url: 'https://example.com/clip.mp4',
        captionsUrl: 'https://example.com/clip.vtt',
      },
      attachTo: document.body,
    })

    const track = document.body.querySelector('track')
    expect(track).toBeTruthy()
    expect(track?.getAttribute('src')).toBe('https://example.com/clip.vtt')
    expect(track?.getAttribute('kind')).toBe('captions')
    wrapper.unmount()
  })

  it('returns focus to the invoking element after close', async () => {
    const trigger = document.createElement('button')
    trigger.textContent = 'Open preview'
    document.body.append(trigger)
    trigger.focus()
    expect(document.activeElement).toBe(trigger)

    const wrapper = mount(MediaLightbox, {
      props: {
        open: false,
        name: 'photo.jpg',
        mimeType: 'image/jpeg',
        url: 'https://example.com/photo.jpg',
      },
      attachTo: document.body,
    })

    trigger.focus()
    await wrapper.setProps({ open: true })
    await flushPromises()
    await nextTick()

    expect(document.activeElement).not.toBe(trigger)

    await wrapper.setProps({ open: false })
    await flushPromises()
    await nextTick()

    expect(document.activeElement).toBe(trigger)
    wrapper.unmount()
  })

  it('closes on Escape via the native dialog', async () => {
    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'photo.jpg',
        mimeType: 'image/jpeg',
        url: 'https://example.com/photo.jpg',
      },
      attachTo: document.body,
    })

    const dialog = document.body.querySelector('dialog') as HTMLDialogElement
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await nextTick()

    expect(wrapper.emitted('close')).toBeTruthy()
    wrapper.unmount()
  })

  it('does not navigate with arrow keys while native media controls are focused', async () => {
    const wrapper = mount(MediaLightbox, {
      props: {
        open: true,
        name: 'clip.mp4',
        mimeType: 'video/mp4',
        url: 'https://example.com/clip.mp4',
        hasPrev: true,
        hasNext: true,
      },
      attachTo: document.body,
    })

    const video = document.body.querySelector('video') as HTMLVideoElement
    const videoFocusSpy = vi.spyOn(document, 'activeElement', 'get').mockReturnValue(video)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowLeft', bubbles: true }))
    await nextTick()

    expect(wrapper.emitted('next')).toBeFalsy()
    expect(wrapper.emitted('prev')).toBeFalsy()
    videoFocusSpy.mockRestore()

    const closeButton = document.body.querySelector('button[aria-label="Close preview"]') as HTMLButtonElement
    const activeElementSpy = vi.spyOn(document, 'activeElement', 'get').mockReturnValue(closeButton)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }))
    await nextTick()
    expect(wrapper.emitted('next')).toBeTruthy()
    activeElementSpy.mockRestore()

    wrapper.unmount()
  })

  it('honors prefers-reduced-motion for transition contracts', () => {
    const source = readSrc('./MediaLightbox.vue')
    expect(source).toMatch(/@media\s*\(\s*prefers-reduced-motion:\s*reduce\s*\)/)
    expect(source).toMatch(/prefersReducedMotion/)
  })
})
