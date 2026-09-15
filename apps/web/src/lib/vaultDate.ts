import type { Locale } from '@/lib/i18n'

const DATE_LOCALES: Record<Locale, string> = {
  vi: 'vi-VN',
  en: 'en-US',
}

export function formatVaultDate(iso: string, locale: Locale): string {
  if (!iso) return ''

  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''

  return date.toLocaleDateString(DATE_LOCALES[locale], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
