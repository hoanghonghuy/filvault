export type ChatAttachmentLocale = 'vi' | 'en'

export type ChatAttachmentCopy = {
  viewImageAria: (name: string) => string
}

const COPY: Record<ChatAttachmentLocale, ChatAttachmentCopy> = {
  vi: {
    viewImageAria: (name) => `Xem ${name}`,
  },
  en: {
    viewImageAria: (name) => `View ${name}`,
  },
}

export function chatAttachmentCopy(locale: ChatAttachmentLocale): ChatAttachmentCopy {
  return COPY[locale]
}
