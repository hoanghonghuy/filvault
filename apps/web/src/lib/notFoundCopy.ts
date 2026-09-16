import type { Locale } from '@/lib/i18n'

export type NotFoundCopy = {
  eyebrow: string
  title: string
  body: string
  home: string
  files: string
  signIn: string
  back: string
}

const copy: Record<Locale, NotFoundCopy> = {
  vi: {
    eyebrow: 'Không tìm thấy trang',
    title: 'Đường dẫn này không tồn tại',
    body: 'Liên kết có thể đã cũ hoặc địa chỉ đã được nhập sai. Bạn có thể quay lại nơi an toàn mà không mất thông tin tài khoản.',
    home: 'Về trang chủ',
    files: 'Mở Tệp của tôi',
    signIn: 'Đăng nhập',
    back: 'Quay lại',
  },
  en: {
    eyebrow: 'Page not found',
    title: 'This path does not exist',
    body: 'The link may be stale or the address may have been mistyped. You can recover safely without exposing account details.',
    home: 'Go home',
    files: 'Open My Files',
    signIn: 'Sign in',
    back: 'Go back',
  },
}

export function notFoundCopy(locale: Locale): NotFoundCopy {
  return copy[locale]
}
