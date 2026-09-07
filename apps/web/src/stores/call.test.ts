/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useCallStore, resolveLiveKitUrl } from '@/stores/call'

// Mock livekit-client
vi.mock('livekit-client', () => ({
  Room: class MockRoom {
    connect = vi.fn<() => Promise<void>>().mockResolvedValue(undefined)
    disconnect = vi.fn<() => Promise<void>>().mockResolvedValue(undefined)
    on = vi.fn<() => void>()
    localParticipant = {
      setMicrophoneEnabled: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
      setCameraEnabled: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
    }
  },
  RoomEvent: {
    ParticipantConnected: 'participantConnected',
    ParticipantDisconnected: 'participantDisconnected',
    TrackSubscribed: 'trackSubscribed',
    TrackUnsubscribed: 'trackUnsubscribed',
    ActiveSpeakersChanged: 'activeSpeakersChanged',
    Disconnected: 'disconnected',
  },
}))

// Mock callAudio
vi.mock('@/lib/callAudio', () => ({
  callAudio: {
    playIncomingRing: vi.fn<() => void>(),
    playOutgoingRing: vi.fn<() => void>(),
    playEndCall: vi.fn<() => void>(),
    stop: vi.fn<() => void>(),
  },
}))

// Mock auth store
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { id: 'user-my-id', displayName: 'My Name' },
    isAuthenticated: true,
  }),
}))

// Mock api client
vi.mock('@/api/client', () => ({
  API_BASE: 'http://localhost:8080/api/v1',
  getAccessToken: () => 'mock-access-token',
}))

describe('useCallStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts in idle state', () => {
    const store = useCallStore()
    expect(store.state).toBe('idle')
    expect(store.conversationId).toBeNull()
    expect(store.isMicEnabled).toBe(true)
    expect(store.isCamEnabled).toBe(true)
  })

  it('handles incoming call invite from another user', () => {
    const store = useCallStore()
    store.handleSignal({
      conversationId: 'conv-123',
      senderId: 'user-other',
      senderName: 'Bob',
      action: 'invite',
      isVideo: true,
      timestamp: new Date().toISOString(),
    })

    expect(store.state).toBe('incoming')
    expect(store.conversationId).toBe('conv-123')
    expect(store.callerName).toBe('Bob')
    expect(store.isVideo).toBe(true)
    expect(store.isCaller).toBe(false)
  })

  it('ignores call invite from self', () => {
    const store = useCallStore()
    store.handleSignal({
      conversationId: 'conv-123',
      senderId: 'user-my-id',
      senderName: 'My Name',
      action: 'invite',
      isVideo: false,
      timestamp: new Date().toISOString(),
    })

    expect(store.state).toBe('idle')
  })

  it('ignores stale call invites older than 45 seconds', () => {
    const store = useCallStore()
    const staleTime = new Date(Date.now() - 50 * 1000).toISOString()
    store.handleSignal({
      conversationId: 'conv-123',
      senderId: 'user-other',
      senderName: 'Bob',
      action: 'invite',
      isVideo: true,
      timestamp: staleTime,
    })

    expect(store.state).toBe('idle')
  })

  it('ignores call invites without timestamp or invalid timestamp', () => {
    const store = useCallStore()
    store.handleSignal({
      conversationId: 'conv-123',
      senderId: 'user-other',
      senderName: 'Bob',
      action: 'invite',
      isVideo: true,
      timestamp: '',
    })

    expect(store.state).toBe('idle')
  })

  it('handles decline and end signals', () => {
    const store = useCallStore()
    store.handleSignal({
      conversationId: 'conv-123',
      senderId: 'user-other',
      senderName: 'Bob',
      action: 'invite',
      isVideo: false,
      timestamp: new Date().toISOString(),
    })
    expect(store.state).toBe('incoming')

    store.handleSignal({
      conversationId: 'conv-123',
      senderId: 'user-other',
      senderName: 'Bob',
      action: 'decline',
      isVideo: false,
      timestamp: new Date().toISOString(),
    })
    expect(store.state).toBe('idle')
  })

  it('toggles microphone and camera state', () => {
    const store = useCallStore()
    expect(store.isMicEnabled).toBe(true)
    store.toggleMicrophone()
    expect(store.isMicEnabled).toBe(false)

    expect(store.isCamEnabled).toBe(true)
    store.toggleCamera()
    expect(store.isCamEnabled).toBe(false)
  })

  it('resolves LiveKit URL by substituting localhost with current LAN hostname', () => {
    // When hostname is localhost, keeps localhost
    expect(resolveLiveKitUrl('ws://localhost:7880')).toBe('ws://localhost:7880')

    // When client accesses from a LAN IP e.g. 192.168.1.6
    vi.stubGlobal('window', {
      location: { hostname: '192.168.1.6', protocol: 'http:' },
    })
    expect(resolveLiveKitUrl('ws://localhost:7880')).toBe('ws://192.168.1.6:7880')
    expect(resolveLiveKitUrl('ws://127.0.0.1:7880')).toBe('ws://192.168.1.6:7880')
    vi.unstubAllGlobals()
  })

  it('keeps caller in outgoing state when initiating call until peer accepts', async () => {
    const store = useCallStore()
    // Global fetch mock for signal
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ status: 'ok' }),
      }),
    )

    await store.startCall('conv-999', {
      isVideo: true,
      peerName: 'Alice',
      peerAvatar: 'https://example.com/alice.jpg',
    })

    expect(store.state).toBe('outgoing')
    expect(store.conversationId).toBe('conv-999')
    expect(store.callerName).toBe('Alice')
    expect(store.callerAvatar).toBe('https://example.com/alice.jpg')
    expect(store.isVideo).toBe(true)
    expect(store.isCaller).toBe(true)

    // Now peer accepts
    store.handleSignal({
      conversationId: 'conv-999',
      senderId: 'user-alice',
      senderName: 'Alice',
      action: 'accept',
      isVideo: true,
      timestamp: new Date().toISOString(),
    })

    expect(store.state).toBe('connected')
    expect(store.formattedDuration).toBe('00:00')

    // End call
    await store.endCall(false)
    expect(store.state).toBe('idle')
    expect(store.formattedDuration).toBe('00:00')
    vi.unstubAllGlobals()
  })

  it('correctly stores callerAvatar from incoming invite signal', () => {
    const store = useCallStore()
    store.handleSignal({
      conversationId: 'conv-abc',
      senderId: 'user-bob',
      senderName: 'Bob',
      senderAvatar: 'https://example.com/bob.jpg',
      action: 'invite',
      isVideo: false,
      timestamp: new Date().toISOString(),
    })

    expect(store.state).toBe('incoming')
    expect(store.callerAvatar).toBe('https://example.com/bob.jpg')
  })
})
