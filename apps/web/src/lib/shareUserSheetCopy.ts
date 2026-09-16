import type { Locale } from '@/lib/i18n'

export type ShareUserSheetCopy = {
  permission: string
  email: string
  invalidEmail: string
  selfShare: string
  conflict: string
  notFound: string
  network: string
  fallback: string
  sharing: string
  share: string
  close: string
  done: string
  shareAnother: string
  invitedPrefix: string
  invitedSuffix: string
  sharedPrefix: string
  sharedSuffix: string
}

const copy: Record<Locale, ShareUserSheetCopy> = {
  vi: {
    permission: 'Người nhận sẽ có quyền đọc mục này. Email chưa đăng ký sẽ nhận lời mời tạo tài khoản.',
    email: 'Email',
    invalidEmail: 'Nhập địa chỉ email hợp lệ.',
    selfShare: 'Bạn không thể chia sẻ cho chính mình.',
    conflict: 'Mục này đã được chia sẻ cho người dùng đó.',
    notFound: 'Không tìm thấy mục để chia sẻ.',
    network: 'Không thể kết nối máy chủ. Hãy thử lại sau ít phút.',
    fallback: 'Không thể chia sẻ mục này.',
    sharing: 'Đang chia sẻ…',
    share: 'Chia sẻ',
    close: 'Đóng',
    done: 'Xong',
    shareAnother: 'Chia sẻ cho người khác',
    invitedPrefix: 'Đã gửi lời mời tới',
    invitedSuffix: 'Họ sẽ có quyền đọc sau khi đăng ký bằng email này.',
    sharedPrefix: 'Đã chia sẻ với',
    sharedSuffix: 'Họ có thể xem và tải mục này với quyền đọc.',
  },
  en: {
    permission: "They'll get read access to this item. Unknown emails receive an invite to sign up.",
    email: 'Email',
    invalidEmail: 'Enter a valid email address.',
    selfShare: "You can't share with yourself.",
    conflict: 'This item is already shared with that user.',
    notFound: 'Item not found.',
    network: "Can't reach the server. Try again in a moment.",
    fallback: 'Could not share this item.',
    sharing: 'Sharing…',
    share: 'Share',
    close: 'Close',
    done: 'Done',
    shareAnother: 'Share with another',
    invitedPrefix: 'Invitation sent to',
    invitedSuffix: "They'll get read access once they sign up with that email.",
    sharedPrefix: 'Shared with',
    sharedSuffix: 'They can view and download this item with read access.',
  },
}

export function shareUserSheetCopy(locale: Locale): ShareUserSheetCopy {
  return copy[locale]
}
