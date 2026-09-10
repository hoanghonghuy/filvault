import type { Locale } from '@/lib/i18n'

export interface TrashCopy {
  deleting: (completed: number, total: number) => string
  partial: (deleted: number, failed: number) => string
  emptyFailed: string
  loadFailed: string
  restoreFailed: string
  deleteFailed: string
  deleteConfirmTitle: string
  deleteConfirmMessage: (name: string) => string
  emptyConfirmMessage: string
}

export function trashCopy(locale: Locale): TrashCopy {
  if (locale === 'vi') {
    return {
      deleting: (completed, total) => `Đang xóa ${completed}/${total}…`,
      partial: (deleted, failed) =>
        `Đã xóa vĩnh viễn ${deleted} mục; ${failed} mục không thể xóa.`,
      emptyFailed: 'Không thể dọn sạch thùng rác',
      loadFailed: 'Không thể tải thùng rác',
      restoreFailed: 'Không thể khôi phục mục',
      deleteFailed: 'Không thể xóa vĩnh viễn mục',
      deleteConfirmTitle: 'Xóa vĩnh viễn?',
      deleteConfirmMessage: (name) =>
        `“${name}” sẽ bị xóa vĩnh viễn. Hành động này không thể hoàn tác.`,
      emptyConfirmMessage:
        'Tất cả tệp và thư mục trong thùng rác sẽ bị xóa vĩnh viễn. Hành động này không thể hoàn tác.',
    }
  }

  return {
    deleting: (completed, total) => `Deleting ${completed}/${total}…`,
    partial: (deleted, failed) =>
      `Permanently deleted ${deleted} items; ${failed} could not be deleted.`,
    emptyFailed: 'Could not empty trash',
    loadFailed: 'Could not load trash',
    restoreFailed: 'Could not restore item',
    deleteFailed: 'Could not permanently delete item',
    deleteConfirmTitle: 'Delete forever?',
    deleteConfirmMessage: (name) =>
      `“${name}” will be permanently deleted. This cannot be undone.`,
    emptyConfirmMessage:
      'All files and folders in Trash will be permanently deleted. This cannot be undone.',
  }
}

export function formatTrashItemDate(iso: string | undefined, locale: Locale): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString(locale === 'vi' ? 'vi-VN' : 'en-US', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}
