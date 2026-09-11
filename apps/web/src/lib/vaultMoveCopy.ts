import type { Locale } from '@/lib/i18n'

export interface VaultMoveCopy {
  moveInFailed: string
  moveInSuccess: string
  moveInBatchSuccess: (count: number) => string
}

export function vaultMoveCopy(locale: Locale): VaultMoveCopy {
  if (locale === 'vi') {
    return {
      moveInFailed: 'Không thể chuyển vào kho cá nhân',
      moveInSuccess: 'Đã chuyển vào kho cá nhân',
      moveInBatchSuccess: (count) => `Đã chuyển ${count} tệp vào kho cá nhân`,
    }
  }

  return {
    moveInFailed: 'Could not move to Personal Vault',
    moveInSuccess: 'Moved to Personal Vault',
    moveInBatchSuccess: (count) => `Moved ${count} ${count === 1 ? 'file' : 'files'} to Personal Vault`,
  }
}
