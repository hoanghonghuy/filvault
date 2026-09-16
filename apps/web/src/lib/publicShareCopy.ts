import type { Locale } from '@/lib/i18n'

export type PublicShareCopy = {
  loading: string
  unavailable: string
  temporary: string
  retry: string
  expires: string
  preview: string
  previewing: string
  previewFailed: string
  download: string
  downloading: string
  downloadFailed: string
  sharedVia: string
}

const COPY: Record<Locale, PublicShareCopy> = {
  vi: {
    loading: 'Đang tải liên kết…',
    unavailable: 'Liên kết này không còn khả dụng',
    temporary: 'Tạm thời không thể tải liên kết này.',
    retry: 'Thử lại',
    expires: 'Hết hạn',
    preview: 'Xem trước',
    previewing: 'Đang chuẩn bị xem trước…',
    previewFailed:
      'Không thể chuẩn bị bản xem trước. Liên kết vẫn còn hiệu lực; hãy thử lại.',
    download: 'Tải xuống',
    downloading: 'Đang chuẩn bị tải xuống…',
    downloadFailed:
      'Không thể chuẩn bị tệp tải xuống. Liên kết vẫn còn hiệu lực; hãy thử lại.',
    sharedVia: 'Được chia sẻ qua Filvault',
  },
  en: {
    loading: 'Loading link…',
    unavailable: 'This link is not available',
    temporary: 'This link could not be loaded right now.',
    retry: 'Retry',
    expires: 'Expires',
    preview: 'Preview',
    previewing: 'Preparing preview…',
    previewFailed: 'Could not prepare the preview. The link is still available; try again.',
    download: 'Download',
    downloading: 'Preparing download…',
    downloadFailed: 'Could not prepare the download. The link is still available; try again.',
    sharedVia: 'Shared via Filvault',
  },
}

export function publicShareCopy(locale: Locale): PublicShareCopy {
  return COPY[locale]
}
