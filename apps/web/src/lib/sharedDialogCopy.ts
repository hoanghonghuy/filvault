import type { Locale } from '@/lib/i18n'

export type SharedDialogCopy = {
  confirm: string
  save: string
  cancel: string
}

export function sharedDialogCopy(locale: Locale): SharedDialogCopy {
  if (locale === 'en') {
    return {
      confirm: 'Confirm',
      save: 'Save',
      cancel: 'Cancel',
    }
  }

  return {
    confirm: 'Xác nhận',
    save: 'Lưu',
    cancel: 'Hủy',
  }
}
