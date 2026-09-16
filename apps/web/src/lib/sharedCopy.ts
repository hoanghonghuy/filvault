import type { ApiErrorCopy } from '@/api/errors'
import type { Locale } from '@/lib/i18n'

export interface SharedCopy {
  apiError: ApiErrorCopy
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
      apiError: {
        network: 'Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.',
        codes: {
          UNAUTHORIZED: 'Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.',
          FORBIDDEN: 'Bạn không có quyền thực hiện thao tác này.',
          NOT_FOUND: 'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
          CONFLICT: 'Tên đã tồn tại ở vị trí này.',
          RATE_LIMITED: 'Bạn thao tác quá nhanh. Vui lòng thử lại sau.',
          QUOTA_EXCEEDED: 'Dung lượng lưu trữ đã đầy. Hãy giải phóng dung lượng rồi thử lại.',
          FILE_TOO_LARGE: 'Tệp vượt quá giới hạn 100 MiB.',
          VALIDATION_ERROR: 'Dữ liệu không hợp lệ. Hãy kiểm tra loại tệp rồi thử lại.',
          INVALID_STATE: 'Không thể thực hiện thao tác này lúc này.',
          UPLOAD_EXPIRED: 'Phiên tải lên đã hết hạn. Hãy tải lại tệp.',
        },
      },
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
    apiError: {},
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
