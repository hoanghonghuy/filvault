import type { Locale } from '@/lib/i18n'

export interface VaultMoveCopy {
  moveInFailed: string
  moveInSuccess: string
  moveInBatchSuccess: (count: number) => string
  confirmSingleTitle: string
  confirmSingleMessage: (name: string) => string
  confirmSingleLabel: string
  confirmBatchTitle: (count: number) => string
  confirmBatchMessage: (count: number) => string
  confirmBatchLabel: string
}

export function vaultMoveCopy(locale: Locale): VaultMoveCopy {
  if (locale === 'vi') {
    return {
      moveInFailed: 'Không thể chuyển vào kho cá nhân',
      moveInSuccess: 'Đã chuyển vào kho cá nhân',
      moveInBatchSuccess: (count) => `Đã chuyển ${count} tệp vào kho cá nhân`,
      confirmSingleTitle: 'Chuyển vào kho cá nhân?',
      confirmSingleMessage: (name) => `"${name}" sẽ được bảo vệ bằng mật khẩu và không hiển thị trong danh sách tệp thông thường.`,
      confirmSingleLabel: 'Chuyển vào kho',
      confirmBatchTitle: (count) => `Chuyển ${count} tệp vào kho cá nhân?`,
      confirmBatchMessage: (count) =>
        count === 1
          ? 'Tệp này sẽ được bảo vệ bằng mật khẩu và không hiển thị trong danh sách tệp thông thường.'
          : 'Các tệp này sẽ được bảo vệ bằng mật khẩu và không hiển thị trong danh sách tệp thông thường.',
      confirmBatchLabel: 'Chuyển vào kho',
    }
  }

  return {
    moveInFailed: 'Could not move to Personal Vault',
    moveInSuccess: 'Moved to Personal Vault',
    moveInBatchSuccess: (count) => `Moved ${count} ${count === 1 ? 'file' : 'files'} to Personal Vault`,
    confirmSingleTitle: 'Move to Personal Vault?',
    confirmSingleMessage: (name) => `"${name}" will be password-protected and hidden from your regular file list.`,
    confirmSingleLabel: 'Move to Vault',
    confirmBatchTitle: (count) => `Move ${count} ${count === 1 ? 'file' : 'files'} to Personal Vault?`,
    confirmBatchMessage: (count) =>
      count === 1
        ? 'This file will be password-protected and hidden from your regular file list.'
        : 'These files will be password-protected and hidden from your regular file list.',
    confirmBatchLabel: 'Move to Vault',
  }
}
