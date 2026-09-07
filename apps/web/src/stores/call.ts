import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'
import { Room, RoomEvent, type RemoteParticipant } from 'livekit-client'
import { useAuthStore } from '@/stores/auth'
import { callAudio } from '@/lib/callAudio'
import { API_BASE, getAccessToken } from '@/api/client'

export type CallState = 'idle' | 'incoming' | 'outgoing' | 'connected' | 'ended'

export interface CallSignalPayload {
  conversationId: string
  senderId: string
  senderName: string
  action: 'invite' | 'accept' | 'decline' | 'end'
  isVideo: boolean
  timestamp: string
}

export const useCallStore = defineStore('call', () => {
  const auth = useAuthStore()
  const state = ref<CallState>('idle')
  const conversationId = ref<string | null>(null)
  const isVideo = ref(false)
  const isCaller = ref(false)
  const callerName = ref('')
  const callerId = ref('')
  const isMicEnabled = ref(true)
  const isCamEnabled = ref(true)
  const activeSpeaker = ref<string | null>(null)
  const participants = ref<Map<string, { identity: string; name: string; isSpeaking: boolean }>>(
    new Map(),
  )
  const room = shallowRef<Room | null>(null)
  const error = ref<string | null>(null)

  function resetState() {
    state.value = 'idle'
    conversationId.value = null
    isVideo.value = false
    isCaller.value = false
    callerName.value = ''
    callerId.value = ''
    isMicEnabled.value = true
    isCamEnabled.value = true
    activeSpeaker.value = null
    participants.value.clear()
    error.value = null
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
    try {
      const { token, url } = await fetchToken(convId)
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
      })

      r.on(RoomEvent.ParticipantDisconnected, (p: RemoteParticipant) => {
        participants.value.delete(p.identity)
      })

      r.on(RoomEvent.ActiveSpeakersChanged, (speakers) => {
        if (speakers.length > 0 && speakers[0]) {
          activeSpeaker.value = speakers[0].identity
        } else {
          activeSpeaker.value = null
        }
      })

      r.on(RoomEvent.Disconnected, () => {
        void endCall(false)
      })

      await r.connect(url, token)
      await r.localParticipant.setMicrophoneEnabled(isMicEnabled.value)
      if (isVideo.value) {
        await r.localParticipant.setCameraEnabled(isCamEnabled.value)
      }
      state.value = 'connected'
      callAudio.stop()
    } catch (e: unknown) {
      const err = e as Error
      error.value = err?.message || 'Lỗi kết nối cuộc gọi'
      void endCall(false)
    }
  }

  async function startCall(convId: string, options: { isVideo?: boolean } = {}) {
    resetState()
    conversationId.value = convId
    isVideo.value = options.isVideo ?? false
    isCaller.value = true
    state.value = 'outgoing'
    callAudio.playOutgoingRing()

    await sendSignal(convId, 'invite', isVideo.value)
    void connectToRoom(convId)
  }

  function handleSignal(payload: CallSignalPayload) {
    // Ignore signals sent by self
    if (payload.senderId === auth.user?.id) return

    switch (payload.action) {
      case 'invite':
        // If already in a call, ignore
        if (state.value !== 'idle') return
        conversationId.value = payload.conversationId
        callerId.value = payload.senderId
        callerName.value = payload.senderName
        isVideo.value = payload.isVideo
        isCaller.value = false
        state.value = 'incoming'
        callAudio.playIncomingRing()
        break

      case 'accept':
        if (state.value === 'outgoing' && conversationId.value === payload.conversationId) {
          callAudio.stop()
          state.value = 'connected'
        }
        break

      case 'decline':
      case 'end':
        if (conversationId.value === payload.conversationId) {
          void endCall(false)
        }
        break
    }
  }

  async function acceptCall() {
    if (!conversationId.value) return
    callAudio.stop()
    const convId = conversationId.value
    state.value = 'connected'
    await sendSignal(convId, 'accept', isVideo.value)
    void connectToRoom(convId)
  }

  async function declineCall() {
    if (!conversationId.value) return
    const convId = conversationId.value
    callAudio.stop()
    await sendSignal(convId, 'decline', isVideo.value)
    resetState()
  }

  async function endCall(notifyRemote = true) {
    callAudio.stop()
    callAudio.playEndCall()
    const convId = conversationId.value
    if (room.value) {
      void room.value.disconnect()
      room.value = null
    }
    if (notifyRemote && convId) {
      void sendSignal(convId, 'end', isVideo.value)
    }
    resetState()
  }

  async function toggleMicrophone() {
    isMicEnabled.value = !isMicEnabled.value
    if (room.value?.localParticipant) {
      await room.value.localParticipant.setMicrophoneEnabled(isMicEnabled.value)
    }
  }

  async function toggleCamera() {
    isCamEnabled.value = !isCamEnabled.value
    if (room.value?.localParticipant) {
      await room.value.localParticipant.setCameraEnabled(isCamEnabled.value)
    }
  }

  return {
    state,
    conversationId,
    isVideo,
    isCaller,
    callerName,
    callerId,
    isMicEnabled,
    isCamEnabled,
    activeSpeaker,
    participants,
    room,
    error,
    startCall,
    handleSignal,
    acceptCall,
    declineCall,
    endCall,
    toggleMicrophone,
    toggleCamera,
  }
})
