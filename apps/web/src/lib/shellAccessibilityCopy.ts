import type { Locale } from './i18n'

export type ShellAccessibilityCopy = {
  mainNavigation: string
  destinations: string
  goToOverview: string
  openProfileFor: (identifier: string) => string
  accountFallback: string
}

const COPY: Record<Locale, ShellAccessibilityCopy> = {
  vi: {
    mainNavigation: 'Điều hướng chính',
    destinations: 'Điểm đến',
    goToOverview: 'Đi tới tổng quan',
    openProfileFor: (identifier) => `Mở hồ sơ của ${identifier}`,
    accountFallback: 'tài khoản',
  },
  en: {
    mainNavigation: 'Main navigation',
    destinations: 'Destinations',
    goToOverview: 'Go to overview',
    openProfileFor: (identifier) => `Open profile for ${identifier}`,
    accountFallback: 'account',
  },
}

export function shellAccessibilityCopy(locale: Locale): ShellAccessibilityCopy {
  return COPY[locale]
}
