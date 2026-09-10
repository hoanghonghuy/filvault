import type { Locale } from '@/lib/i18n'

export interface TrashRetentionSettings {
  trashAutoDeleteEnabled: boolean
  trashRetentionDays: number
}

export function trashRetentionNotice(
  settings: TrashRetentionSettings | null | undefined,
  locale: Locale,
): string {
  if (!settings) {
    return locale === 'vi'
      ? 'Chưa thể xác định chính sách tự động xóa của thùng rác.'
      : 'Trash automatic-deletion policy is currently unavailable.'
  }

  if (!settings.trashAutoDeleteEnabled) {
    return locale === 'vi'
      ? 'Tự động xóa vĩnh viễn đang tắt. Các mục sẽ ở lại đây cho đến khi bạn xóa chúng.'
      : 'Automatic permanent deletion is off. Items stay here until you delete them.'
  }

  const days = Math.max(1, Math.floor(Number(settings.trashRetentionDays) || 1))
  return locale === 'vi'
    ? `Các mục trong thùng rác sẽ tự động bị xóa vĩnh viễn sau ${days} ngày.`
    : `Items in Trash are permanently deleted automatically after ${days} days.`
}
