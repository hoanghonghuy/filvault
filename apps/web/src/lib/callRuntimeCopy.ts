import type { Locale } from '@/lib/i18n'

export type CallRuntimeCopy = {
  microphoneFailed: string
  cameraFailed: string
  connectionFailed: string
  unanswered: string
  declined: string
  callerCancelled: string
  ended: string
}

const COPY: Record<Locale, CallRuntimeCopy> = {
  vi: {
    microphoneFailed: 'Không thể bật micro. Hãy kiểm tra quyền truy cập thiết bị rồi thử lại.',
    cameraFailed: 'Không thể bật camera. Hãy kiểm tra quyền truy cập; cuộc gọi vẫn tiếp tục bằng thoại.',
    connectionFailed: 'Lỗi kết nối cuộc gọi',
    unanswered: 'Người nhận không trả lời',
    declined: 'Người nhận đã từ chối cuộc gọi',
    callerCancelled: 'Người gọi đã hủy cuộc gọi',
    ended: 'Cuộc gọi đã kết thúc',
  },
  en: {
    microphoneFailed: 'Could not enable the microphone. Check device permission and try again.',
    cameraFailed: 'Could not enable the camera. Check device permission; the call will continue with audio.',
    connectionFailed: 'Could not connect the call',
    unanswered: 'No answer',
    declined: 'The recipient declined the call',
    callerCancelled: 'The caller cancelled the call',
    ended: 'The call has ended',
  },
}

export function getCallRuntimeCopy(locale: Locale): CallRuntimeCopy {
  return COPY[locale]
}
