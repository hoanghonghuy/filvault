import { describe, expect, it } from 'vitest'
import {
  isLegacyLogoutConfirmation,
  logoutConfirmationOptions,
} from './logoutConfirmation'

describe('logout confirmation contract', () => {
  it('provides the same destructive intent in English and Vietnamese', () => {
    expect(logoutConfirmationOptions('en')).toEqual({
      title: 'Log out?',
      message: 'You will need to sign in again to access your files.',
      confirmLabel: 'Log out',
      cancelLabel: 'Cancel',
      danger: true,
    })
    expect(logoutConfirmationOptions('vi')).toEqual({
      title: 'Đăng xuất?',
      message: 'Bạn sẽ cần đăng nhập lại để truy cập các tệp của mình.',
      confirmLabel: 'Đăng xuất',
      cancelLabel: 'Hủy',
      danger: true,
    })
  })

  it('recognizes only the historical Settings logout confirmation', () => {
    expect(
      isLegacyLogoutConfirmation({
        title: 'Log out?',
        message: 'You will need to sign in again to access your files.',
        confirmLabel: 'Log out',
      }),
    ).toBe(true)

    expect(
      isLegacyLogoutConfirmation({
        title: 'Delete file?',
        message: 'You will need to sign in again to access your files.',
        confirmLabel: 'Delete',
      }),
    ).toBe(false)
  })
})
