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
const hasRemoteVideo = ref(false)
const isFullscreen = ref(false)

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

const peerAvatar = computed(() => callStore.callerAvatar || null)

function updateRemoteVideoState(activeRoom: Room) {
  let found = false
  activeRoom.remoteParticipants.forEach((participant: RemoteParticipant) => {
    participant.trackPublications.forEach((pub: RemoteTrackPublication) => {
      if (pub.kind === Track.Kind.Video && pub.track && !pub.isMuted) {
        found = true
        if (remoteVideoRef.value) {
          pub.track.attach(remoteVideoRef.value)
        }
      }
    })
  })
  hasRemoteVideo.value = found
}

function attachLocalCamera(activeRoom: Room) {
  if (!localVideoRef.value) return
  const camPub = activeRoom.localParticipant.getTrackPublication(Track.Source.Camera)
  if (camPub?.track) {
    camPub.track.attach(localVideoRef.value)
  }
}

function toggleFullscreen() {
  const elem = document.querySelector('.call-active-room') || document.documentElement
  if (!document.fullscreenElement) {
    void elem.requestFullscreen?.()
    isFullscreen.value = true
  } else {
    void document.exitFullscreen?.()
    isFullscreen.value = false
  }
}

// Watch for connected state & room instance
watch(
  [() => callStore.state, () => callStore.room],
  async ([state, room]) => {
    if (state !== 'connected' || !room) {
      hasRemoteVideo.value = false
      return
    }

    const activeRoom = room as Room
    await nextTick()

    // Unlock autoplay if browser suspended Web Audio
    try {
      await activeRoom.startAudio()
    } catch {
      // Ignored if user hasn't interacted yet
    }

    // Attach initial local video track
    attachLocalCamera(activeRoom)

    // Attach initial remote video and audio tracks
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
    updateRemoteVideoState(activeRoom)

    // Listen to local track published / unpublished
    activeRoom.on(RoomEvent.LocalTrackPublished, (pub) => {
      if (pub.kind === Track.Kind.Video && localVideoRef.value) {
        pub.track?.attach(localVideoRef.value)
      }
    })

    activeRoom.on(RoomEvent.LocalTrackUnpublished, (pub) => {
      if (pub.kind === Track.Kind.Video) {
        pub.track?.detach()
      }
    })

    // Listen to remote tracks subscribed
    activeRoom.on(RoomEvent.TrackSubscribed, (track: RemoteTrack) => {
      if (track.kind === Track.Kind.Video) {
        if (remoteVideoRef.value) track.attach(remoteVideoRef.value)
        hasRemoteVideo.value = true
      } else if (track.kind === Track.Kind.Audio && remoteAudioRef.value) {
        track.attach(remoteAudioRef.value)
      }
    })

    activeRoom.on(RoomEvent.TrackUnsubscribed, (track: RemoteTrack) => {
      track.detach()
      updateRemoteVideoState(activeRoom)
    })

    activeRoom.on(RoomEvent.TrackMuted, (pub) => {
      if (pub.kind === Track.Kind.Video) {
        updateRemoteVideoState(activeRoom)
      }
    })

    activeRoom.on(RoomEvent.TrackUnmuted, (pub) => {
      if (pub.kind === Track.Kind.Video) {
        updateRemoteVideoState(activeRoom)
      }
    })
  },
  { immediate: true },
)

// Re-attach local video if camera is re-enabled
watch(
  () => callStore.isCamEnabled,
  async (enabled) => {
    if (!enabled || !callStore.room || callStore.state !== 'connected') return
    await nextTick()
    attachLocalCamera(callStore.room as Room)
  },
)

onUnmounted(() => {
  if (isFullscreen.value && document.fullscreenElement) {
    void document.exitFullscreen?.()
  }
  callStore.endCall(false)
})
</script>

<template>
  <Teleport to="body">
    <div v-if="isOpen" class="call-overlay" role="dialog" :aria-label="callTitle" aria-modal="true">
      <!-- Offscreen audio element for remote audio tracks (never display: none to prevent browser mute) -->
      <audio ref="remoteAudioRef" autoplay playsinline class="offscreen-audio"></audio>

      <!-- 1. Incoming Call Prompt -->
      <div v-if="callStore.state === 'incoming'" class="call-incoming-card">
        <div class="caller-avatar-pulse">
          <div class="avatar-circle">
            <img v-if="peerAvatar" :src="peerAvatar" :alt="peerName" class="avatar-img" />
            <span v-else>{{ peerName.charAt(0).toUpperCase() }}</span>
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
            <Icon name="phone-off" :size="26" />
          </button>
          <button
            type="button"
            class="call-btn btn-accept"
            aria-label="Trả lời cuộc gọi"
            title="Trả lời"
            @click="() => callStore.acceptCall()"
          >
            <Icon name="phone" :size="26" />
          </button>
        </div>
      </div>

      <!-- 2. Outgoing Call Screen -->
      <div v-else-if="callStore.state === 'outgoing'" class="call-outgoing-card">
        <div class="caller-avatar-pulse outgoing">
          <div class="avatar-circle">
            <img v-if="peerAvatar" :src="peerAvatar" :alt="peerName" class="avatar-img" />
            <span v-else>{{ peerName.charAt(0).toUpperCase() }}</span>
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
            <Icon name="phone-off" :size="26" />
          </button>
        </div>
      </div>

      <!-- 3. Connected Call Screen -->
      <div v-else-if="callStore.state === 'connected'" class="call-active-room">
        <!-- Call Header Bar -->
        <header class="call-header-bar">
          <div class="header-peer-info">
            <div class="status-indicator">
              <span class="status-dot"></span>
              <span class="duration-badge">{{ callStore.formattedDuration }}</span>
            </div>
            <span class="header-peer-name">{{ peerName }}</span>
          </div>
          <button
            type="button"
            class="header-action-btn"
            :aria-label="isFullscreen ? 'Thu nhỏ' : 'Toàn màn hình'"
            :title="isFullscreen ? 'Thu nhỏ' : 'Toàn màn hình'"
            @click="toggleFullscreen"
          >
            <Icon :name="isFullscreen ? 'close' : 'more'" :size="20" />
          </button>
        </header>

        <!-- Media Surface Area -->
        <div class="call-media-container" :class="{ 'audio-only': !callStore.isVideo || !hasRemoteVideo }">
          <!-- Remote Video Element -->
          <video
            v-show="callStore.isVideo && hasRemoteVideo"
            ref="remoteVideoRef"
            autoplay
            playsinline
            class="remote-video"
          ></video>

          <!-- Audio-only / Camera off fallback presentation -->
          <div v-if="!callStore.isVideo || !hasRemoteVideo" class="audio-caller-display">
            <div class="avatar-circle large" :class="{ speaking: callStore.activeSpeaker }">
              <img v-if="peerAvatar" :src="peerAvatar" :alt="peerName" class="avatar-img" />
              <span v-else>{{ peerName.charAt(0).toUpperCase() }}</span>
            </div>
            <h2 class="active-peer-name">{{ peerName }}</h2>
            <p class="call-timer-label">
              {{ callStore.isVideo && !hasRemoteVideo ? 'Camera đối phương đang tắt' : 'Đang trong cuộc gọi' }}
            </p>
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
            <div v-if="!callStore.isCamEnabled" class="cam-off-overlay">
              <Icon name="camera-off" :size="20" />
              <span class="cam-off-badge">Camera tắt</span>
            </div>
          </div>
        </div>

        <!-- Floating Controls Dock -->
        <div class="call-controls-dock">
          <!-- Mic Toggle -->
          <button
            type="button"
            class="control-btn"
            :class="{ muted: !callStore.isMicEnabled }"
            :aria-label="callStore.isMicEnabled ? 'Tắt mic' : 'Bật mic'"
            :title="callStore.isMicEnabled ? 'Tắt mic' : 'Bật mic'"
            @click="callStore.toggleMicrophone"
          >
            <Icon :name="callStore.isMicEnabled ? 'mic' : 'mic-off'" :size="22" />
          </button>

          <!-- Camera Toggle (if video enabled) -->
          <button
            v-if="callStore.isVideo"
            type="button"
            class="control-btn"
            :class="{ muted: !callStore.isCamEnabled }"
            :aria-label="callStore.isCamEnabled ? 'Tắt camera' : 'Bật camera'"
            :title="callStore.isCamEnabled ? 'Tắt camera' : 'Bật camera'"
            @click="callStore.toggleCamera"
          >
            <Icon :name="callStore.isCamEnabled ? 'camera' : 'camera-off'" :size="22" />
          </button>

          <!-- Hangup / End Call -->
          <button
            type="button"
            class="control-btn btn-hangup"
            aria-label="Kết thúc cuộc gọi"
            title="Kết thúc cuộc gọi"
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
  background: rgba(10, 14, 23, 0.9);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  padding: var(--space-md);
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}

/* Offscreen audio avoids Chromium/WebKit suppression of display:none elements */
.offscreen-audio {
  position: fixed;
  top: -9999px;
  left: -9999px;
  width: 1px;
  height: 1px;
  opacity: 0.001;
  pointer-events: none;
}

/* Incoming & Outgoing cards */
.call-incoming-card,
.call-outgoing-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  background: #182234;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 28px;
  padding: 36px 28px;
  max-width: 360px;
  width: 100%;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
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
  margin-bottom: 24px;
}

.caller-avatar-pulse::before,
.caller-avatar-pulse::after {
  content: '';
  position: absolute;
  inset: -14px;
  border-radius: 50%;
  border: 2px solid #0084ff;
  opacity: 0.8;
  animation: ripple 1.8s infinite cubic-bezier(0.2, 0.8, 0.2, 1);
}

.caller-avatar-pulse::after {
  animation-delay: 0.6s;
}

.caller-avatar-pulse.outgoing::before,
.caller-avatar-pulse.outgoing::after {
  border-color: #3b82f6;
}

@keyframes ripple {
  0% {
    transform: scale(0.85);
    opacity: 0.9;
  }
  100% {
    transform: scale(1.45);
    opacity: 0;
  }
}

.avatar-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #0084ff 0%, #00c6ff 100%);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 700;
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(0, 132, 255, 0.3);
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-circle.large {
  width: 120px;
  height: 120px;
  font-size: 48px;
  box-shadow: 0 12px 32px rgba(0, 132, 255, 0.4);
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.25s ease;
  margin-bottom: 16px;
}

.avatar-circle.speaking {
  transform: scale(1.08);
  box-shadow: 0 0 0 6px #10b981, 0 12px 32px rgba(16, 185, 129, 0.5);
}

.caller-name,
.active-peer-name {
  font-size: 1.35rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 6px;
}

.call-subtitle,
.call-timer-label {
  font-size: 0.9rem;
  color: #94a3b8;
  margin: 0 0 32px;
}

.call-actions {
  display: flex;
  align-items: center;
  gap: 32px;
}

.call-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
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
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.45);
}

.btn-decline {
  background: #ef4444;
  box-shadow: 0 6px 20px rgba(239, 68, 68, 0.45);
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
  border-radius: 24px;
  overflow: hidden;
  background: #0b0f19;
}

.call-header-bar {
  position: absolute;
  top: 16px;
  left: 16px;
  right: 16px;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(15, 23, 42, 0.65);
  backdrop-filter: blur(12px);
  padding: 8px 16px;
  border-radius: 9999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.header-peer-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(0, 0, 0, 0.3);
  padding: 3px 8px;
  border-radius: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
  animation: pulseDot 2s infinite ease-in-out;
}

@keyframes pulseDot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.duration-badge {
  font-size: 12px;
  font-weight: 600;
  color: #e2e8f0;
  font-variant-numeric: tabular-nums;
}

.header-peer-name {
  font-size: 14px;
  font-weight: 600;
  color: #ffffff;
}

.header-action-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s ease;
}

.header-action-btn:hover {
  color: #ffffff;
}

.call-media-container {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: #000000;
}

.call-media-container.audio-only {
  background: radial-gradient(circle at center, #1e293b 0%, #0b0f19 100%);
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
  padding: 24px;
  text-align: center;
}

.local-video-wrapper {
  position: absolute;
  right: 16px;
  bottom: 88px;
  width: 110px;
  height: 150px;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.7);
  border: 2px solid rgba(255, 255, 255, 0.25);
  background: #1e293b;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 5;
}

.local-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* Mirror camera preview */
}

.cam-off-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #1e293b;
  color: #94a3b8;
  gap: 6px;
}

.cam-off-badge {
  font-size: 11px;
  font-weight: 500;
  color: #94a3b8;
  text-align: center;
}

/* Floating Controls Dock */
.call-controls-dock {
  position: absolute;
  bottom: 24px;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 16px;
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 9999px;
  padding: 10px 20px;
  box-shadow: 0 10px 36px rgba(0, 0, 0, 0.5);
}

.control-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.14);
  color: #ffffff;
  cursor: pointer;
  transition: transform 0.15s ease, background 0.15s ease;
  touch-action: manipulation;
}

.control-btn:active {
  transform: scale(0.92);
}

.control-btn.muted {
  background: #ef4444;
  color: #ffffff;
}

.control-btn.btn-hangup {
  background: #ef4444;
  color: #ffffff;
  box-shadow: 0 4px 16px rgba(239, 68, 68, 0.5);
}

@media (min-width: 768px) {
  .call-incoming-card,
  .call-outgoing-card {
    max-width: 400px;
    padding: 44px 36px;
  }

  .call-active-room {
    max-width: 960px;
    max-height: 85vh;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.7);
  }

  .local-video-wrapper {
    width: 150px;
    height: 200px;
    right: 24px;
    bottom: 96px;
  }
}
</style>
