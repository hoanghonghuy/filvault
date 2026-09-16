/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CallModal from './CallModal.vue'
import { useAuthStore } from '@/stores/auth'
import { useCallStore } from '@/stores/call'
import { setLocale } from '@/lib/i18n'

vi.mock('@/lib/callAudio', () => ({
  callAudio: {
    playIncomingRing: vi.fn<() => void>(),
    playOutgoingRing: vi.fn<() => void>(),
    playEndCall: vi.fn<() => void>(),
    stop: vi.fn<() => void>(),
  },
}))

vi.mock('livekit-client', () => ({
  Track: {
    Kind: {
      Video: 'video',
      Audio: 'audio',
    },
    Source: {
      Camera: 'camera',
    },
  },
  RoomEvent: {
    LocalTrackPublished: 'localTrackPublished',
    LocalTrackUnpublished: 'localTrackUnpublished',
    TrackSubscribed: 'trackSubscribed',
    TrackUnsubscribed: 'trackUnsubscribed',
    TrackMuted: 'trackMuted',
    TrackUnmuted: 'trackUnmuted',
  },
}))

const iconStub = { template: '<span data-testid="icon-stub" />' }

describe('CallModal locale reactivity', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    document.body.innerHTML = ''
    setLocale('vi')
    pinia = createPinia()
    setActivePinia(pinia)
    const auth = useAuthStore(pinia)
    auth.user = {
      id: '01KCHATAUTH00000000000000001',
      email: 'sender@example.com',
      displayName: 'Sender',
      emailVerified: true,
      storageUsed: 0,
      storageQuota: 1024,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-08-28T12:00:00.000Z',
    }
  })

  afterEach(() => {
    document.body.innerHTML = ''
    vi.restoreAllMocks()
  })

  function findInBody(selector: string) {
    const element = document.body.querySelector(selector)
    expect(element).toBeTruthy()
    return element as HTMLElement
  }

  function mountIncomingVideoCall() {
    const callStore = useCallStore(pinia)
    callStore.state = 'incoming'
    callStore.isVideo = true
    callStore.callerName = 'Alice'

    return mount(CallModal, {
      global: {
        plugins: [pinia],
        stubs: { Icon: iconStub },
      },
      attachTo: document.body,
    })
  }

  function mountConnectedAudioCall() {
    const callStore = useCallStore(pinia)
    callStore.state = 'connected'
    callStore.isVideo = false
    callStore.callerName = 'Alice'
    callStore.isMicEnabled = true

    return mount(CallModal, {
      global: {
        plugins: [pinia],
        stubs: { Icon: iconStub },
      },
      attachTo: document.body,
    })
  }

  it('updates incoming video dialog title and decline/answer controls from Vietnamese to English without remounting', async () => {
    const wrapper = mountIncomingVideoCall()
    await flushPromises()

    const dialog = findInBody('[role="dialog"]')
    expect(dialog.getAttribute('aria-label')).toBe('Cuộc gọi video đến')
    expect(findInBody('.call-subtitle').textContent).toBe('Cuộc gọi video đến')

    const declineBtn = findInBody('.btn-decline')
    const answerBtn = findInBody('.btn-accept')
    expect(declineBtn.getAttribute('aria-label')).toBe('Từ chối cuộc gọi')
    expect(declineBtn.getAttribute('title')).toBe('Từ chối')
    expect(answerBtn.getAttribute('aria-label')).toBe('Trả lời cuộc gọi')
    expect(answerBtn.getAttribute('title')).toBe('Trả lời')

    setLocale('en')
    await flushPromises()

    expect(dialog.getAttribute('aria-label')).toBe('Incoming video call')
    expect(findInBody('.call-subtitle').textContent).toBe('Incoming video call')
    expect(declineBtn.getAttribute('aria-label')).toBe('Decline call')
    expect(declineBtn.getAttribute('title')).toBe('Decline')
    expect(answerBtn.getAttribute('aria-label')).toBe('Answer call')
    expect(answerBtn.getAttribute('title')).toBe('Answer')

    wrapper.unmount()
  })

  it('updates connected call mute control labels from Vietnamese to English without remounting', async () => {
    const wrapper = mountConnectedAudioCall()
    await flushPromises()

    const muteBtn = findInBody('.control-btn')
    expect(muteBtn.getAttribute('aria-label')).toBe('Tắt mic')
    expect(muteBtn.getAttribute('title')).toBe('Tắt mic')

    setLocale('en')
    await flushPromises()

    expect(muteBtn.getAttribute('aria-label')).toBe('Mute microphone')
    expect(muteBtn.getAttribute('title')).toBe('Mute microphone')

    wrapper.unmount()
  })
})
