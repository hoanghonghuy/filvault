import type { Locale } from '@/lib/i18n'

export type CallRuntimeCopy = {
  microphoneFailed: string
  cameraFailed: string
  microphoneToggleFailed: string
  cameraToggleFailed: string
  mediaRequiresSecureContext: string
  connectionFailed: string
  peerLeft: string
  reconnecting: string
  reconnected: string
  disconnected: string
  unanswered: string
  declined: string
  callerCancelled: string
  ended: string
}

const COPY: Record<Locale, CallRuntimeCopy> = {
  vi: {
    microphoneFailed: 'Không thể bật micro. Hãy kiểm tra quyền truy cập thiết bị rồi thử lại.',
    cameraFailed: 'Không thể bật camera. Hãy kiểm tra quyền truy cập; cuộc gọi vẫn tiếp tục bằng thoại.',
    microphoneToggleFailed: 'Không thể thay đổi micro. Hãy kiểm tra quyền truy cập thiết bị rồi thử lại.',
    cameraToggleFailed: 'Không thể thay đổi camera. Hãy kiểm tra quyền truy cập thiết bị rồi thử lại.',
    mediaRequiresSecureContext: 'Micro và camera cần kết nối HTTPS hoặc localhost.',
    connectionFailed: 'Lỗi kết nối cuộc gọi',
    peerLeft: 'Đối phương đã rời cuộc gọi',
    reconnecting: 'Kết nối cuộc gọi không ổn định. Đang thử kết nối lại.',
    reconnected: 'Đã kết nối lại cuộc gọi',
    disconnected: 'Mất kết nối cuộc gọi. Hãy kiểm tra mạng và gọi lại.',
    unanswered: 'Người nhận không trả lời',
    declined: 'Người nhận đã từ chối cuộc gọi',
    callerCancelled: 'Người gọi đã hủy cuộc gọi',
    ended: 'Cuộc gọi đã kết thúc',
  },
  en: {
    microphoneFailed: 'Could not enable the microphone. Check device permission and try again.',
    cameraFailed: 'Could not enable the camera. Check device permission; the call will continue with audio.',
    microphoneToggleFailed: 'Could not change the microphone. Check device permission and try again.',
    cameraToggleFailed: 'Could not change the camera. Check device permission and try again.',
    mediaRequiresSecureContext: 'Microphone and camera access requires HTTPS or localhost.',
    connectionFailed: 'Could not connect the call',
    peerLeft: 'The other participant left the call',
    reconnecting: 'The call connection is unstable. Reconnecting…',
    reconnected: 'Call reconnected',
    disconnected: 'The call disconnected. Check your network and call again.',
    unanswered: 'No answer',
    declined: 'The recipient declined the call',
    callerCancelled: 'The caller cancelled the call',
    ended: 'The call has ended',
  },
}

export function getCallRuntimeCopy(locale: Locale): CallRuntimeCopy {
  return COPY[locale]
}
