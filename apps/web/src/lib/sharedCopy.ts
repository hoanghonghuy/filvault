import type { Locale } from '@/lib/i18n'

export interface SharedCopy {
  loadIncomingFailed: string
  loadLinksFailed: string
  openFolderFailed: string
  downloadFailed: string
  downloading: (name: string) => string
  linkCopied: string
  copyFailed: string
  revokeTitle: string
  revokeMessage: (name: string) => string
  revokeConfirm: string
  revokeSuccess: string
  revokeFailed: string
  copyLink: string
  openLink: string
  revokeLink: string
  activeForever: string
  expired: string
  expiresOn: (date: string) => string
  mySharesEmptyTitle: string
  mySharesEmptyDescription: string
  sharesViewsAria: string
  shareActionsAria: string
  sharedFolderContentsAria: string
  sharedFolderPathAria: string
}

export function sharedCopy(locale: Locale): SharedCopy {
  if (locale === 'vi') {
    return {
      loadIncomingFailed: 'Không thể tải các mục được chia sẻ',
      loadLinksFailed: 'Không thể tải các liên kết chia sẻ của bạn',
      openFolderFailed: 'Không thể mở thư mục',
      downloadFailed: 'Không thể tải tệp xuống',
      downloading: (name) => `Đang tải “${name}”`,
      linkCopied: 'Đã sao chép liên kết',
      copyFailed: 'Không thể sao chép liên kết',
      revokeTitle: 'Thu hồi liên kết?',
      revokeMessage: (name) => `“${name}” sẽ không còn được chia sẻ công khai.`,
      revokeConfirm: 'Thu hồi liên kết',
      revokeSuccess: 'Đã thu hồi liên kết',
      revokeFailed: 'Không thể thu hồi liên kết',
      copyLink: 'Sao chép liên kết',
      openLink: 'Mở liên kết',
      revokeLink: 'Thu hồi liên kết',
      activeForever: 'Có hiệu lực vĩnh viễn',
      expired: 'Đã hết hạn',
      expiresOn: (date) => `Hết hạn ${date}`,
      mySharesEmptyTitle: 'Chưa có liên kết chia sẻ nào',
      mySharesEmptyDescription: 'Khi bạn tạo liên kết công khai cho tệp, chúng sẽ xuất hiện tại đây.',
      sharesViewsAria: 'Các chế độ xem chia sẻ',
      shareActionsAria: 'Thao tác chia sẻ',
      sharedFolderContentsAria: 'Nội dung thư mục được chia sẻ',
      sharedFolderPathAria: 'Đường dẫn thư mục được chia sẻ',
    }
  }

  return {
    loadIncomingFailed: 'Could not load shared items',
    loadLinksFailed: 'Could not load your share links',
    openFolderFailed: 'Could not open folder',
    downloadFailed: 'Could not download file',
    downloading: (name) => `Downloading “${name}”`,
    linkCopied: 'Link copied',
    copyFailed: 'Could not copy link',
    revokeTitle: 'Revoke link?',
    revokeMessage: (name) => `“${name}” will no longer be shared publicly.`,
    revokeConfirm: 'Revoke link',
    revokeSuccess: 'Link revoked',
    revokeFailed: 'Failed to revoke link',
    copyLink: 'Copy link',
    openLink: 'Open link',
    revokeLink: 'Revoke link',
    activeForever: 'Active forever',
    expired: 'Expired',
    expiresOn: (date) => `Expires ${date}`,
    mySharesEmptyTitle: 'No share links yet',
    mySharesEmptyDescription: 'Public links you create for files will appear here.',
    sharesViewsAria: 'Share views',
    shareActionsAria: 'Share actions',
    sharedFolderContentsAria: 'Shared folder contents',
    sharedFolderPathAria: 'Shared folder path',
  }
}

export function formatSharedDate(iso: string, locale: Locale): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString(locale === 'vi' ? 'vi-VN' : 'en-US', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export function formatShareExpiry(
  expiresAt: string | null,
  locale: Locale,
  now = Date.now(),
): string {
  const copy = sharedCopy(locale)
  if (!expiresAt) return copy.activeForever
  const expires = new Date(expiresAt)
  if (Number.isNaN(expires.getTime())) return ''
  if (expires.getTime() < now) return copy.expired
  return copy.expiresOn(formatSharedDate(expiresAt, locale))
}
