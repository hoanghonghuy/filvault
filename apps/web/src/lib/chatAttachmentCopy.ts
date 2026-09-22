export type ChatAttachmentLocale = 'vi' | 'en'

export type ChatAttachmentCopy = {
  viewImageAria: (name: string) => string
  previewUnavailable: string
  openAttachment: string
}

const COPY: Record<ChatAttachmentLocale, ChatAttachmentCopy> = {
  vi: {
    viewImageAria: (name) => `Xem ${name}`,
    previewUnavailable: 'Không thể xem trước ảnh',
    openAttachment: 'Nhấn để mở tệp',
  },
  en: {
    viewImageAria: (name) => `View ${name}`,
    previewUnavailable: 'Image preview unavailable',
    openAttachment: 'Open attachment',
  },
}

export function chatAttachmentCopy(locale: ChatAttachmentLocale): ChatAttachmentCopy {
  return COPY[locale]
}
