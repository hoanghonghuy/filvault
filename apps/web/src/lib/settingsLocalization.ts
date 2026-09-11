import type { ActivityEventType } from '@/api/types'
import type { Locale } from '@/lib/i18n'

export type SettingsLocaleText = {
  avatar: string
  fallbackUser: string
  languageGroup: string
  quickThemesGroup: string
  sharedLinks: string
  noActiveShareLinks: string
  copyLink: string
  revokeLink: string
  linkCopied: string
  copyFailed: string
  revokeConfirmTitle: string
  revokeConfirmMessage: (fileName?: string) => string
  linkRevoked: string
  revokeFailed: string
  expires: (date: string) => string
  neverExpires: string
  activity: string
  activityLoadFailed: string
  activityLoadMoreFailed: string
  noRecentActivity: string
}

const COPY: Record<Locale, SettingsLocaleText> = {
  vi: {
    avatar: 'Ảnh đại diện',
    fallbackUser: 'Người dùng Filvault',
    languageGroup: 'Ngôn ngữ',
    quickThemesGroup: 'Chủ đề nhanh',
    sharedLinks: 'Liên kết đã chia sẻ',
    noActiveShareLinks: 'Không có liên kết chia sẻ đang hoạt động.',
    copyLink: 'Sao chép liên kết',
    revokeLink: 'Thu hồi liên kết',
    linkCopied: 'Đã sao chép liên kết',
    copyFailed: 'Không thể sao chép liên kết',
    revokeConfirmTitle: 'Thu hồi liên kết?',
    revokeConfirmMessage: (fileName) =>
      fileName
        ? `“${fileName}” sẽ không còn được chia sẻ công khai.`
        : 'Mục này sẽ không còn được chia sẻ công khai.',
    linkRevoked: 'Đã thu hồi liên kết',
    revokeFailed: 'Không thể thu hồi liên kết',
    expires: (date) => `Hết hạn ${date}`,
    neverExpires: 'Không hết hạn',
    activity: 'Hoạt động',
    activityLoadFailed: 'Không thể tải hoạt động.',
    activityLoadMoreFailed: 'Không thể tải thêm hoạt động',
    noRecentActivity: 'Chưa có hoạt động gần đây.',
  },
  en: {
    avatar: 'Avatar',
    fallbackUser: 'Filvault user',
    languageGroup: 'Language',
    quickThemesGroup: 'Quick themes',
    sharedLinks: 'Shared links',
    noActiveShareLinks: 'No active share links.',
    copyLink: 'Copy link',
    revokeLink: 'Revoke link',
    linkCopied: 'Link copied',
    copyFailed: 'Copy failed',
    revokeConfirmTitle: 'Revoke link?',
    revokeConfirmMessage: (fileName) =>
      fileName
        ? `“${fileName}” will no longer be shared publicly.`
        : 'This item will no longer be shared publicly.',
    linkRevoked: 'Link revoked',
    revokeFailed: 'Could not revoke link',
    expires: (date) => `Expires ${date}`,
    neverExpires: 'Never expires',
    activity: 'Activity',
    activityLoadFailed: 'Could not load activity.',
    activityLoadMoreFailed: 'Could not load more activity',
    noRecentActivity: 'No recent activity.',
  },
}

const ACTIVITY_LABELS: Record<Locale, Partial<Record<ActivityEventType, string>>> = {
  vi: {
    'file.uploaded': 'Đã tải lên',
    'file.trashed': 'Đã chuyển vào thùng rác',
    'file.restored': 'Đã khôi phục',
    'file.purged': 'Đã xóa vĩnh viễn',
    'folder.trashed': 'Đã chuyển thư mục vào thùng rác',
    'folder.restored': 'Đã khôi phục thư mục',
    'share.created': 'Đã tạo chia sẻ',
    'share.revoked': 'Đã thu hồi chia sẻ',
    'password.changed': 'Đã đổi mật khẩu',
    'settings.changed': 'Đã cập nhật cài đặt',
  },
  en: {
    'file.uploaded': 'Uploaded',
    'file.trashed': 'Moved to trash',
    'file.restored': 'Restored',
    'file.purged': 'Permanently deleted',
    'folder.trashed': 'Folder moved to trash',
    'folder.restored': 'Folder restored',
    'share.created': 'Share created',
    'share.revoked': 'Share revoked',
    'password.changed': 'Password changed',
    'settings.changed': 'Settings updated',
  },
}

function intlLocale(locale: Locale): string {
  return locale === 'vi' ? 'vi-VN' : 'en-US'
}

export function getSettingsText(locale: Locale): SettingsLocaleText {
  return COPY[locale]
}

export function settingsActivityLabel(locale: Locale, type: ActivityEventType): string {
  return ACTIVITY_LABELS[locale][type] ?? type
}

export function formatSettingsDate(locale: Locale, iso: string): string {
  return new Intl.DateTimeFormat(intlLocale(locale), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(new Date(iso))
}

export function formatSettingsRelativeTime(locale: Locale, iso: string, now = Date.now()): string {
  const seconds = Math.max(0, Math.round((now - new Date(iso).getTime()) / 1000))
  const formatter = new Intl.RelativeTimeFormat(intlLocale(locale), { numeric: 'auto' })

  if (seconds < 60) return formatter.format(0, 'second')
  const minutes = Math.round(seconds / 60)
  if (minutes < 60) return formatter.format(-minutes, 'minute')
  const hours = Math.round(minutes / 60)
  if (hours < 24) return formatter.format(-hours, 'hour')
  const days = Math.round(hours / 24)
  if (days < 30) return formatter.format(-days, 'day')
  return formatSettingsDate(locale, iso)
}
