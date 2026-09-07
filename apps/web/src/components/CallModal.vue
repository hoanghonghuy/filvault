<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { useCallStore } from '@/stores/call'
import Icon from '@/components/AppIcon.vue'
import {
  Track,
  RoomEvent,
  type RemoteParticipant,
  type RemoteTrackPublication,
  type RemoteTrack,
  type Room,
} from 'livekit-client'

const callStore = useCallStore()

const localVideoRef = ref<HTMLVideoElement | null>(null)
const remoteVideoRef = ref<HTMLVideoElement | null>(null)
const remoteAudioRef = ref<HTMLAudioElement | null>(null)

const isOpen = computed(() => callStore.state !== 'idle')

const callTitle = computed(() => {
  if (callStore.state === 'incoming') {
    return callStore.isVideo ? 'Cuộc gọi video đến' : 'Cuộc gọi thoại đến'
  }
  if (callStore.state === 'outgoing') {
    return callStore.isVideo ? 'Đang gọi video...' : 'Đang gọi thoại...'
  }
  return callStore.isVideo ? 'Cuộc gọi video' : 'Cuộc gọi thoại'
})

const peerName = computed(() => {
  if (callStore.callerName) return callStore.callerName
  if (callStore.participants.size > 0) {
    const first = callStore.participants.values().next().value
    return first?.name || 'Thành viên'
  }
  return 'Cuộc trò chuyện'
})

// Attach media tracks when connected
watch(
  [() => callStore.state, () => callStore.room],
  async ([state, room]) => {
    if (state !== 'connected' || !room) return
    const activeRoom = room as Room
    await nextTick()

    // Attach local video track
    const localVideoPublication = activeRoom.localParticipant.getTrackPublication(
      Track.Source.Camera,
    )
    if (localVideoPublication?.track && localVideoRef.value) {
      localVideoPublication.track.attach(localVideoRef.value)
    }

    // Attach remote video & audio tracks
    activeRoom.remoteParticipants.forEach((participant: RemoteParticipant) => {
      participant.trackPublications.forEach((pub: RemoteTrackPublication) => {
        if (pub.track) {
          if (pub.kind === Track.Kind.Video && remoteVideoRef.value) {
            pub.track.attach(remoteVideoRef.value)
          } else if (pub.kind === Track.Kind.Audio && remoteAudioRef.value) {
            pub.track.attach(remoteAudioRef.value)
          }
        }
      })
    })

    // Listen to new tracks subscribed
    activeRoom.on(RoomEvent.TrackSubscribed, (track: RemoteTrack) => {
      if (track.kind === Track.Kind.Video && remoteVideoRef.value) {
        track.attach(remoteVideoRef.value)
      } else if (track.kind === Track.Kind.Audio && remoteAudioRef.value) {
        track.attach(remoteAudioRef.value)
      }
    })

    activeRoom.on(RoomEvent.TrackUnsubscribed, (track: RemoteTrack) => {
      track.detach()
    })
  },
  { immediate: true },
)

onUnmounted(() => {
  callStore.endCall(false)
})
</script>

<template>
  <Teleport to="body">
    <div v-if="isOpen" class="call-overlay" role="dialog" :aria-label="callTitle" aria-modal="true">
      <!-- Hidden audio element for remote audio tracks -->
      <audio ref="remoteAudioRef" autoplay playsinline class="hidden-audio"></audio>

      <!-- 1. Incoming Call Prompt -->
      <div v-if="callStore.state === 'incoming'" class="call-incoming-card">
        <div class="caller-avatar-pulse">
          <div class="avatar-circle">
            {{ peerName.charAt(0).toUpperCase() }}
          </div>
        </div>
        <h2 class="caller-name">{{ peerName }}</h2>
        <p class="call-subtitle">{{ callTitle }}</p>

        <div class="call-actions">
          <button
            type="button"
            class="call-btn btn-decline"
            aria-label="Từ chối cuộc gọi"
            title="Từ chối"
            @click="() => callStore.declineCall()"
          >
            <Icon name="phone-off" :size="24" />
          </button>
          <button
            type="button"
            class="call-btn btn-accept"
            aria-label="Trả lời cuộc gọi"
            title="Trả lời"
            @click="() => callStore.acceptCall()"
          >
            <Icon name="phone" :size="24" />
          </button>
        </div>
      </div>

      <!-- 2. Outgoing Call Screen -->
      <div v-else-if="callStore.state === 'outgoing'" class="call-outgoing-card">
        <div class="caller-avatar-pulse outgoing">
          <div class="avatar-circle">
            {{ peerName.charAt(0).toUpperCase() }}
          </div>
        </div>
        <h2 class="caller-name">{{ peerName }}</h2>
        <p class="call-subtitle">Đang đổ chuông...</p>

        <div class="call-actions">
          <button
            type="button"
            class="call-btn btn-decline"
            aria-label="Hủy cuộc gọi"
            title="Hủy"
            @click="() => callStore.endCall()"
          >
            <Icon name="phone-off" :size="24" />
          </button>
        </div>
      </div>

      <!-- 3. Connected Call Screen -->
      <div v-else-if="callStore.state === 'connected'" class="call-active-room">
        <div class="call-media-container" :class="{ 'audio-only': !callStore.isVideo }">
          <!-- Remote Video Surface -->
          <video
            v-show="callStore.isVideo"
            ref="remoteVideoRef"
            autoplay
            playsinline
            class="remote-video"
          ></video>

          <!-- Audio-only presentation -->
          <div v-if="!callStore.isVideo" class="audio-caller-display">
            <div class="avatar-circle large" :class="{ speaking: callStore.activeSpeaker }">
              {{ peerName.charAt(0).toUpperCase() }}
            </div>
            <h2 class="active-peer-name">{{ peerName }}</h2>
            <p class="call-timer-label">Đang kết nối</p>
          </div>

          <!-- Local Video PiP (Picture in Picture) -->
          <div v-show="callStore.isVideo" class="local-video-wrapper">
            <video
              ref="localVideoRef"
              autoplay
              playsinline
              muted
              class="local-video"
              :class="{ disabled: !callStore.isCamEnabled }"
            ></video>
            <span v-if="!callStore.isCamEnabled" class="cam-off-badge">Camera tắt</span>
          </div>
        </div>

        <!-- Floating Controls Dock -->
        <div class="call-controls-dock">
          <!-- Mic Toggle -->
          <button
            type="button"
            class="control-btn"
            :class="{ active: !callStore.isMicEnabled }"
            :aria-label="callStore.isMicEnabled ? 'Tắt mic' : 'Bật mic'"
            @click="callStore.toggleMicrophone"
          >
            <Icon :name="callStore.isMicEnabled ? 'mic' : 'mic-off'" :size="22" />
          </button>

          <!-- Camera Toggle (if video enabled) -->
          <button
            v-if="callStore.isVideo"
            type="button"
            class="control-btn"
            :class="{ active: !callStore.isCamEnabled }"
            :aria-label="callStore.isCamEnabled ? 'Tắt camera' : 'Bật camera'"
            @click="callStore.toggleCamera"
          >
            <Icon :name="callStore.isCamEnabled ? 'camera' : 'camera-off'" :size="22" />
          </button>

          <!-- Hangup / End Call -->
          <button
            type="button"
            class="control-btn btn-hangup"
            aria-label="Kết thúc cuộc gọi"
            title="Kết thúc"
            @click="() => callStore.endCall()"
          >
            <Icon name="phone-off" :size="24" />
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.call-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(10, 14, 23, 0.85);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  padding: var(--space-md);
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}

.hidden-audio {
  display: none;
}

/* Incoming & Outgoing cards */
.call-incoming-card,
.call-outgoing-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  background: var(--surface-card, #1c2438);
  border: 1px solid var(--border-soft, rgba(255, 255, 255, 0.1));
  border-radius: var(--radius-xl, 24px);
  padding: var(--space-xl) var(--space-2xl);
  max-width: 360px;
  width: 100%;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.caller-avatar-pulse {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--space-lg);
}

.caller-avatar-pulse::before {
  content: '';
  position: absolute;
  inset: -12px;
  border-radius: 50%;
  border: 2px solid #0084ff;
  opacity: 0.8;
  animation: ripple 1.6s infinite ease-out;
}

@keyframes ripple {
  0% {
    transform: scale(0.85);
    opacity: 1;
  }
  100% {
    transform: scale(1.4);
    opacity: 0;
  }
}

.avatar-circle {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: linear-gradient(135deg, #0084ff 0%, #00c6ff 100%);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
}

.avatar-circle.large {
  width: 110px;
  height: 110px;
  font-size: 44px;
  box-shadow: 0 8px 24px rgba(0, 132, 255, 0.4);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.avatar-circle.speaking {
  transform: scale(1.08);
  box-shadow: 0 0 0 6px #10b981;
}

.caller-name,
.active-peer-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 4px;
}

.call-subtitle,
.call-timer-label {
  font-size: 0.875rem;
  color: var(--text-muted, #94a3b8);
  margin: 0 0 var(--space-xl);
}

.call-actions {
  display: flex;
  align-items: center;
  gap: var(--space-xl);
}

.call-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 60px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  color: #ffffff;
  transition: transform 0.15s ease, filter 0.15s ease;
  touch-action: manipulation;
}

.call-btn:active {
  transform: scale(0.92);
}

.btn-accept {
  background: #10b981;
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.4);
}

.btn-decline {
  background: #ef4444;
  box-shadow: 0 4px 16px rgba(239, 68, 68, 0.4);
}

/* 3. Connected Call Room */
.call-active-room {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
}

.call-media-container {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-lg, 16px);
  overflow: hidden;
  background: #000000;
}

.call-media-container.audio-only {
  background: transparent;
}

.remote-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.audio-caller-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.local-video-wrapper {
  position: absolute;
  right: var(--space-md);
  bottom: 96px;
  width: 120px;
  height: 160px;
  border-radius: var(--radius-md, 12px);
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
  border: 2px solid rgba(255, 255, 255, 0.2);
  background: #1e293b;
  display: flex;
  align-items: center;
  justify-content: center;
}

.local-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* Mirror camera preview */
}

.cam-off-badge {
  font-size: 11px;
  color: #94a3b8;
  text-align: center;
}

/* Floating Controls Dock */
.call-controls-dock {
  position: absolute;
  bottom: var(--space-lg);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-pill, 9999px);
  padding: 8px 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.control-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
  cursor: pointer;
  transition: transform 0.15s ease, background 0.15s ease;
  touch-action: manipulation;
}

.control-btn:active {
  transform: scale(0.92);
}

.control-btn.active {
  background: #f59e0b;
  color: #ffffff;
}

.control-btn.btn-hangup {
  background: #ef4444;
  color: #ffffff;
  box-shadow: 0 4px 14px rgba(239, 68, 68, 0.45);
}

@media (min-width: 768px) {
  .call-incoming-card,
  .call-outgoing-card {
    max-width: 420px;
    padding: var(--space-2xl);
  }

  .call-active-room {
    max-width: 900px;
    max-height: 80vh;
    border-radius: var(--radius-xl, 24px);
    overflow: hidden;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.6);
  }

  .local-video-wrapper {
    width: 160px;
    height: 220px;
    right: var(--space-lg);
    bottom: 96px;
  }
}
</style>
