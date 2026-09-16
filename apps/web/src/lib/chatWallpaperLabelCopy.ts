import type { Locale } from '@/lib/i18n'

export type ChatWallpaperLabelCopy = {
  default: string
  customImage: string
}

const copy: Record<Locale, ChatWallpaperLabelCopy> = {
  vi: {
    default: 'Mặc định',
    customImage: 'Ảnh tùy chỉnh',
  },
  en: {
    default: 'Default',
    customImage: 'Custom image',
  },
}

export function chatWallpaperLabelCopy(locale: Locale): ChatWallpaperLabelCopy {
  return copy[locale]
}
