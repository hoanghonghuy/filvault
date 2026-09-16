import type { Locale } from '@/lib/i18n'

export type CallModalCopy = {
  incomingVideo: string
  incomingAudio: string
  outgoingVideo: string
  outgoingAudio: string
  videoCall: string
  audioCall: string
  member: string
  conversation: string
  declineCall: string
  decline: string
  answerCall: string
  answer: string
  ringing: string
  cancelCall: string
  cancel: string
  minimize: string
  fullscreen: string
  remoteCameraOff: string
  inCall: string
  cameraOff: string
  muteMic: string
  unmuteMic: string
  turnCameraOff: string
  turnCameraOn: string
  endCall: string
  callEnded: string
}

const COPY: Record<Locale, CallModalCopy> = {
  vi: {
    incomingVideo: 'Cuộc gọi video đến',
    incomingAudio: 'Cuộc gọi thoại đến',
    outgoingVideo: 'Đang gọi video…',
    outgoingAudio: 'Đang gọi thoại…',
    videoCall: 'Cuộc gọi video',
    audioCall: 'Cuộc gọi thoại',
    member: 'Thành viên',
    conversation: 'Cuộc trò chuyện',
    declineCall: 'Từ chối cuộc gọi',
    decline: 'Từ chối',
    answerCall: 'Trả lời cuộc gọi',
    answer: 'Trả lời',
    ringing: 'Đang đổ chuông…',
    cancelCall: 'Hủy cuộc gọi',
    cancel: 'Hủy',
    minimize: 'Thu nhỏ',
    fullscreen: 'Toàn màn hình',
    remoteCameraOff: 'Camera đối phương đang tắt',
    inCall: 'Đang trong cuộc gọi',
    cameraOff: 'Camera tắt',
    muteMic: 'Tắt mic',
    unmuteMic: 'Bật mic',
    turnCameraOff: 'Tắt camera',
    turnCameraOn: 'Bật camera',
    endCall: 'Kết thúc cuộc gọi',
    callEnded: 'Cuộc gọi đã kết thúc',
  },
  en: {
    incomingVideo: 'Incoming video call',
    incomingAudio: 'Incoming voice call',
    outgoingVideo: 'Calling video…',
    outgoingAudio: 'Calling…',
    videoCall: 'Video call',
    audioCall: 'Voice call',
    member: 'Member',
    conversation: 'Conversation',
    declineCall: 'Decline call',
    decline: 'Decline',
    answerCall: 'Answer call',
    answer: 'Answer',
    ringing: 'Ringing…',
    cancelCall: 'Cancel call',
    cancel: 'Cancel',
    minimize: 'Exit full screen',
    fullscreen: 'Full screen',
    remoteCameraOff: 'Their camera is off',
    inCall: 'In call',
    cameraOff: 'Camera off',
    muteMic: 'Mute microphone',
    unmuteMic: 'Unmute microphone',
    turnCameraOff: 'Turn camera off',
    turnCameraOn: 'Turn camera on',
    endCall: 'End call',
    callEnded: 'Call ended',
  },
}

export function callModalCopy(locale: Locale): CallModalCopy {
  return COPY[locale]
}
