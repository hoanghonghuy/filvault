import type { Locale } from '@/lib/i18n'
import type { ConfirmOptions } from '@/stores/ui'

const STORAGE_KEY = 'filvault.locale'

const logoutCopy = {
  vi: {
    title: 'Đăng xuất?',
    message: 'Bạn sẽ cần đăng nhập lại để truy cập các tệp của mình.',
    confirmLabel: 'Đăng xuất',
    cancelLabel: 'Hủy',
  },
  en: {
    title: 'Log out?',
    message: 'You will need to sign in again to access your files.',
    confirmLabel: 'Log out',
    cancelLabel: 'Cancel',
  },
} as const

export function currentLogoutLocale(): Locale {
  if (typeof localStorage === 'undefined') return 'vi'
  return localStorage.getItem(STORAGE_KEY) === 'en' ? 'en' : 'vi'
}

export function logoutConfirmationOptions(locale: Locale): ConfirmOptions {
  return {
    ...logoutCopy[locale],
    danger: true,
  }
}

export function isLegacyLogoutConfirmation(options: ConfirmOptions): boolean {
  return (
    options.title === logoutCopy.en.title &&
    options.message === logoutCopy.en.message &&
    options.confirmLabel === logoutCopy.en.confirmLabel
  )
}
