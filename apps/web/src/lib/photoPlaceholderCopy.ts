import type { Locale } from '@/lib/i18n'

export type PhotoPlaceholderCopy = {
  photo: string
  video: string
}

const copy: Record<Locale, PhotoPlaceholderCopy> = {
  vi: {
    photo: 'Ảnh',
    video: 'Video',
  },
  en: {
    photo: 'Photo',
    video: 'Video',
  },
}

export function photoPlaceholderCopy(locale: Locale): PhotoPlaceholderCopy {
  return copy[locale]
}
