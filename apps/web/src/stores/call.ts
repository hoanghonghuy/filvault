import { defineStore } from 'pinia'
import { computed, ref, shallowRef } from 'vue'
import { Room, RoomEvent, type RemoteParticipant } from 'livekit-client'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { callAudio } from '@/lib/callAudio'
import { getCallRuntimeCopy } from '@/lib/callRuntimeCopy'
import { useI18n } from '@/lib/i18n'
import { API_BASE, getAccessToken } from '@/api/client'

export type CallState = 'idle' | 'incoming' | 'outgoing' | 'connected' | 'ended'
export type CallConnectionStatus = 'idle' | 'connecting' | 'connected' | 'reconnecting'

export interface CallSignalPayload {
  conversationId: string
  senderId: string
  senderName: string
  senderAvatar?: string
  action: 'invite' | 'accept' | 'decline' | 'end'
  isVideo: boolean
  timestamp: string
}

export function resolveLiveKitUrl(rawUrl: string): string {
  if (typeof window === 'undefined' || !window.location?.hostname) {
    return rawUrl
  }
  let url = rawUrl
  const currentHost = window.location.hostname
  if (currentHost !== 'localhost' && currentHost !== '127.0.0.1') {
    url = url.replace(/localhost|127\.0\.0\.1/, currentHost)
  }
  if (window.location.protocol === 'https:' && url.startsWith('ws://')) {
    url = url.replace(/^ws:\/\//, 'wss://')
  }
  return url
}

export const useCallStore = defineStore('call', () => {
  const auth = useAuthStore()
  const ui = useUiStore()
  const { locale } = useI18n()
  const runtimeCopy = computed(() => getCallRuntimeCopy(locale.value))
  const state = ref<CallState>('idle')
  const connectionStatus = ref<CallConnectionStatus>('idle')
  const conversationId = ref<string | null>(null)
  const isVideo = ref(false)
  const isCaller = ref(false)
  const callerName = ref('')
  const callerAvatar = ref<string | null>(null)
  const callerId = ref('')
  const isMicEnabled = ref(true)
  const isCamEnabled = ref(true)
  const activeSpeaker = ref<string | null>(null)
  const participants = ref<Map<string, { identity: string; name: string; isSpeaking: boolean }>>(
    new Map(),
  )
  const room = shallowRef<Room | null>(null)
  const error = ref<string | null>(null)
  const microphoneError = ref<string | null>(null)
  const cameraError = ref<string | null>(null)
  const deviceError = computed(() => microphoneError.value ?? cameraError.value)
  const callDuration = ref(0)
  let callDurationTimer: number | null = null
  let incomingRingTimer: number | null = null
  let outgoingRingTimer: number | null = null
  const intentionalDisconnectRooms = new WeakSet<Room>()

  const formattedDuration = computed(() => {
    const mins = Math.floor(callDuration.value / 60)
    const secs = callDuration.value % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  })

  function startTimer() {
    stopTimer()
    callDuration.value = 0
    callDurationTimer = window.setInterval(() => {
      callDuration.value++
    }, 1000)
  }

  function stopTimer() {
    if (callDurationTimer !== null) {
      window.clearInterval(callDurationTimer)
      callDurationTimer = null
    }
  }

  function disconnectRoom(activeRoom: Room) {
    intentionalDisconnectRooms.add(activeRoom)
    void activeRoom.disconnect()
  }

  function resetState() {
    stopTimer()
    if (room.value) {
      disconnectRoom(room.value)
      room.value = null
    }
    callDuration.value = 0
    if (incomingRingTimer !== null) {
      window.clearTimeout(incomingRingTimer)
      incomingRingTimer = null
    }
    if (outgoingRingTimer !== null) {
      window.clearTimeout(outgoingRingTimer)
      outgoingRingTimer = null
    }
    state.value = 'idle'
    connectionStatus.value = 'idle'
    conversationId.value = null
    isVideo.value = false
    isCaller.value = false
    callerName.value = ''
    callerAvatar.value = null
    callerId.value = ''
    isMicEnabled.value = true
    isCamEnabled.value = true
    activeSpeaker.value = null
    participants.value.clear()
    error.value = null
    microphoneError.value = null
    cameraError.value = null
    callAudio.stop()
  }

  async function fetchToken(convId: string): Promise<{ token: string; url: string; room: string }> {
    const token = getAccessToken()
    const res = await fetch(`${API_BASE}/chat/conversations/${convId}/call/token`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })
    if (!res.ok) throw new Error(`Failed to get call token (${res.status})`)
    return res.json()
  }

  async function sendSignal(
    convId: string,
    action: 'invite' | 'accept' | 'decline' | 'end',
    video: boolean,
  ) {
    const token = getAccessToken()
    try {
      await fetch(`${API_BASE}/chat/conversations/${convId}/call/signal`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ action, isVideo: video }),
      })
    } catch {
      // ignore network errors for signals
    }
  }

  async function connectToRoom(convId: string) {
    connectionStatus.value = 'connecting'
    error.value = null
    microphoneError.value = null
    cameraError.value = null
    try {
      const { token, url } = await fetchToken(convId)
      if (state.value === 'idle' || state.value === 'ended' || conversationId.value !== convId) {
        connectionStatus.value = 'idle'
        return
      }
      const connectUrl = resolveLiveKitUrl(url)

      if (typeof navigator !== 'undefined' && !navigator.mediaDevices?.getUserMedia) {
        throw new Error(runtimeCopy.value.mediaRequiresSecureContext)
      }

      const r = new Room({
        adaptiveStream: true,
        dynacast: true,
      })
      room.value = r

      r.on(RoomEvent.ParticipantConnected, (p: RemoteParticipant) => {
        participants.value.set(p.identity, {
          identity: p.identity,
          name: p.name || p.identity,
          isSpeaking: false,
        })
        participants.value = new Map(participants.value)
      })

      r.on(RoomEvent.ParticipantDisconnected, (p: RemoteParticipant) => {
        participants.value.delete(p.identity)
        participants.value = new Map(participants.value)
        if (state.value === 'connected' && participants.value.size === 0) {
          ui.showToast(runtimeCopy.value.peerLeft, 'info')
          void endCall(false)
        }
      })

      r.on(RoomEvent.ActiveSpeakersChanged, (speakers) => {
        if (speakers.length > 0 && speakers[0]) {
          activeSpeaker.value = speakers[0].identity
        } else {
          activeSpeaker.value = null
        }
      })

      r.on(RoomEvent.Reconnecting, () => {
        if (state.value !== 'connected') return
        connectionStatus.value = 'reconnecting'
        error.value = runtimeCopy.value.reconnecting
        ui.showToast(runtimeCopy.value.reconnecting, 'info')
      })

      r.on(RoomEvent.Reconnected, () => {
        if (state.value !== 'connected') return
        connectionStatus.value = 'connected'
        error.value = null
        ui.showToast(runtimeCopy.value.reconnected, 'success')
      })

      r.on(RoomEvent.Disconnected, () => {
        connectionStatus.value = 'idle'
        if (intentionalDisconnectRooms.has(r)) {
          intentionalDisconnectRooms.delete(r)
          return
        }
        if (state.value === 'connected') {
          const message = runtimeCopy.value.disconnected
          error.value = message
          ui.showToast(message, 'error')
        }
        void endCall(false)
      })

      await r.connect(connectUrl, token)

      // Bail out if call was cancelled during connection
      const currentCallState = state.value as CallState
      if (currentCallState === 'idle' || currentCallState === 'ended' || conversationId.value !== convId) {
        disconnectRoom(r)
        room.value = null
        connectionStatus.value = 'idle'
        return
      }

      // Populate pre-existing participants (Bug L4)
      r.remoteParticipants.forEach((p) => {
        participants.value.set(p.identity, {
          identity: p.identity,
          name: p.name || p.identity,
          isSpeaking: false,
        })
      })
      participants.value = new Map(participants.value)

      try {
        await r.localParticipant.setMicrophoneEnabled(isMicEnabled.value)
        microphoneError.value = null
      } catch {
        isMicEnabled.value = false
        microphoneError.value = runtimeCopy.value.microphoneFailed
        ui.showToast(microphoneError.value, 'error')
      }

      // Camera: graceful degradation (Bug L5)
      if (isVideo.value) {
        try {
          await r.localParticipant.setCameraEnabled(isCamEnabled.value)
          cameraError.value = null
        } catch {
          isCamEnabled.value = false
          cameraError.value = runtimeCopy.value.cameraFailed
          ui.showToast(cameraError.value, 'info')
        }
      }

      // Only NOW transition to connected state (Bug L1)
      state.value = 'connected'
      connectionStatus.value = 'connected'
      startTimer()
    } catch (e: unknown) {
      const err = e as Error
      const msg =
        err?.message === runtimeCopy.value.mediaRequiresSecureContext
          ? runtimeCopy.value.mediaRequiresSecureContext
          : runtimeCopy.value.connectionFailed
      connectionStatus.value = 'idle'
      error.value = msg
      ui.showToast(msg, 'error')
      void endCall(false)
    }
  }

  async function startCall(
    convId: string,
    options: { isVideo?: boolean; peerName?: string; peerAvatar?: string } = {},
  ) {
    resetState()
    conversationId.value = convId
    isVideo.value = options.isVideo ?? false
    isCaller.value = true
    callerName.value = options.peerName ?? ''
    callerAvatar.value = options.peerAvatar ?? null
    state.value = 'outgoing'
    callAudio.playOutgoingRing()

    outgoingRingTimer = window.setTimeout(() => {
      if (state.value === 'outgoing') {
        ui.showToast(runtimeCopy.value.unanswered, 'info')
        void endCall(true)
      }
    }, 45000)

    await sendSignal(convId, 'invite', isVideo.value)
  }

  function handleSignal(payload: CallSignalPayload) {
    // Ignore signals sent by self
    if (payload.senderId === auth.user?.id) return

    switch (payload.action) {
      case 'invite': {
        // If already in a call, reject as busy
        if (state.value !== 'idle') {
          void sendSignal(payload.conversationId, 'decline', payload.isVideo)
          return
        }

        // Ignore stale/expired call invites
        if (!payload.timestamp) return
        const inviteTime = new Date(payload.timestamp).getTime()
        const now = Date.now()
        // An invite older than 45 seconds or skewed into future is expired
        if (isNaN(inviteTime) || now - inviteTime > 45 * 1000 || inviteTime - now > 60 * 1000) {
          return
        }

        conversationId.value = payload.conversationId
        callerId.value = payload.senderId
        callerName.value = payload.senderName
        callerAvatar.value = payload.senderAvatar || null
        isVideo.value = payload.isVideo
        isCaller.value = false
        state.value = 'incoming'
        callAudio.playIncomingRing()

        if (incomingRingTimer !== null) {
          window.clearTimeout(incomingRingTimer)
        }
        const remainingMs = Math.max(5000, 45000 - (now - inviteTime))
        incomingRingTimer = window.setTimeout(() => {
          if (state.value === 'incoming') {
            callAudio.stop()
            resetState()
          }
        }, remainingMs)
        break
      }

      case 'accept':
        if (outgoingRingTimer !== null) {
          window.clearTimeout(outgoingRingTimer)
          outgoingRingTimer = null
        }
        if (state.value === 'outgoing' && conversationId.value === payload.conversationId) {
          callAudio.stop()
          void connectToRoom(payload.conversationId)
        }
        break

      case 'decline':
        if (conversationId.value === payload.conversationId) {
          if (state.value === 'outgoing') {
            ui.showToast(runtimeCopy.value.declined, 'info')
          }
          void endCall(false)
        }
        break

      case 'end':
        if (conversationId.value === payload.conversationId) {
          if (state.value === 'incoming') {
            ui.showToast(runtimeCopy.value.callerCancelled, 'info')
          } else if (state.value === 'connected') {
            ui.showToast(runtimeCopy.value.ended, 'info')
          }
          void endCall(false)
        }
        break
    }
  }

  async function acceptCall() {
    if (!conversationId.value) return
    if (incomingRingTimer !== null) {
      window.clearTimeout(incomingRingTimer)
      incomingRingTimer = null
    }
    callAudio.stop()
    const convId = conversationId.value
    await sendSignal(convId, 'accept', isVideo.value)
    void connectToRoom(convId)
  }

  async function declineCall() {
    if (!conversationId.value) return
    if (incomingRingTimer !== null) {
      window.clearTimeout(incomingRingTimer)
      incomingRingTimer = null
    }
    const convId = conversationId.value
    callAudio.stop()
    await sendSignal(convId, 'decline', isVideo.value)
    resetState()
  }

  async function endCall(notifyRemote = true) {
    stopTimer()
    callAudio.stop()
    callAudio.playEndCall()
    const convId = conversationId.value
    connectionStatus.value = 'idle'
    if (room.value) {
      disconnectRoom(room.value)
      room.value = null
    }
    if (notifyRemote && convId) {
      void sendSignal(convId, 'end', isVideo.value)
    }
    state.value = 'ended'
    window.setTimeout(() => resetState(), 1800)
  }

  async function toggleMicrophone() {
    const next = !isMicEnabled.value
    if (room.value?.localParticipant) {
      try {
        await room.value.localParticipant.setMicrophoneEnabled(next)
        isMicEnabled.value = next
        microphoneError.value = null
      } catch {
        microphoneError.value = runtimeCopy.value.microphoneToggleFailed
        ui.showToast(microphoneError.value, 'error')
      }
    } else {
      isMicEnabled.value = next
    }
  }

  async function toggleCamera() {
    const next = !isCamEnabled.value
    if (room.value?.localParticipant) {
      try {
        await room.value.localParticipant.setCameraEnabled(next)
        isCamEnabled.value = next
        cameraError.value = null
      } catch {
        cameraError.value = runtimeCopy.value.cameraToggleFailed
        ui.showToast(cameraError.value, 'error')
      }
    } else {
      isCamEnabled.value = next
    }
  }

  return {
    state,
    connectionStatus,
    conversationId,
    isVideo,
    isCaller,
    callerName,
    callerAvatar,
    callerId,
    isMicEnabled,
    isCamEnabled,
    activeSpeaker,
    participants,
    room,
    error,
    deviceError,
    callDuration,
    formattedDuration,
    startCall,
    handleSignal,
    acceptCall,
    declineCall,
    endCall,
    toggleMicrophone,
    toggleCamera,
  }
})
