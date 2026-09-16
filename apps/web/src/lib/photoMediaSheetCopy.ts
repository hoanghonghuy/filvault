import type { Locale } from '@/lib/i18n'

export type PhotoMediaSheetCopy = {
  view: string
}

const copy: Record<Locale, PhotoMediaSheetCopy> = {
  vi: {
    view: 'Xem',
  },
  en: {
    view: 'View',
  },
}

export function photoMediaSheetCopy(locale: Locale): PhotoMediaSheetCopy {
  return copy[locale]
}
