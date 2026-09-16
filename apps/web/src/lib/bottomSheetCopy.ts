import type { Locale } from '@/lib/i18n'

export type BottomSheetCopy = {
  untitledDialogAria: string
}

const copy: Record<Locale, BottomSheetCopy> = {
  vi: {
    untitledDialogAria: 'Hộp thoại',
  },
  en: {
    untitledDialogAria: 'Dialog',
  },
}

export function bottomSheetCopy(locale: Locale): BottomSheetCopy {
  return copy[locale]
}
