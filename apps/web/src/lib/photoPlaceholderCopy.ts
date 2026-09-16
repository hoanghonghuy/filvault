import type { Locale } from '@/lib/i18n'

export type PhotoPlaceholderCopy = {
  photo: string
}

const copy: Record<Locale, PhotoPlaceholderCopy> = {
  vi: {
    photo: 'Ảnh',
  },
  en: {
    photo: 'Photo',
  },
}

export function photoPlaceholderCopy(locale: Locale): PhotoPlaceholderCopy {
  return copy[locale]
}
