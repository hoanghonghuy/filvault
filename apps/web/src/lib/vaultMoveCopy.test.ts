import { describe, expect, it } from 'vitest'
import { vaultMoveCopy } from './vaultMoveCopy'

describe('vaultMoveCopy', () => {
  it('keeps single-file move feedback locale-consistent', () => {
    const vi = vaultMoveCopy('vi')
    const en = vaultMoveCopy('en')

    expect(vi.moveInFailed).toBe('Không thể chuyển vào kho cá nhân')
    expect(vi.moveInSuccess).toBe('Đã chuyển vào kho cá nhân')
    expect(en.moveInFailed).toBe('Could not move to Personal Vault')
    expect(en.moveInSuccess).toBe('Moved to Personal Vault')
  })

  it('formats batch success copy with an explicit count in both locales', () => {
    expect(vaultMoveCopy('vi').moveInBatchSuccess(3)).toBe('Đã chuyển 3 tệp vào kho cá nhân')
    expect(vaultMoveCopy('en').moveInBatchSuccess(1)).toBe('Moved 1 file to Personal Vault')
    expect(vaultMoveCopy('en').moveInBatchSuccess(3)).toBe('Moved 3 files to Personal Vault')
  })

  it('keeps single-file confirmation copy locale-consistent', () => {
    const vi = vaultMoveCopy('vi')
    const en = vaultMoveCopy('en')

    expect(vi.confirmSingleTitle).toBe('Chuyển vào kho cá nhân?')
    expect(vi.confirmSingleMessage('secret.pdf')).toContain('"secret.pdf"')
    expect(vi.confirmSingleLabel).toBe('Chuyển vào kho')

    expect(en.confirmSingleTitle).toBe('Move to Personal Vault?')
    expect(en.confirmSingleMessage('secret.pdf')).toBe(
      '"secret.pdf" will be password-protected and hidden from your regular file list.',
    )
    expect(en.confirmSingleLabel).toBe('Move to Vault')
  })

  it('formats batch confirmation copy with count-aware English grammar', () => {
    const vi = vaultMoveCopy('vi')
    const en = vaultMoveCopy('en')

    expect(vi.confirmBatchTitle(3)).toBe('Chuyển 3 tệp vào kho cá nhân?')
    expect(vi.confirmBatchMessage).toContain('bảo vệ bằng mật khẩu')
    expect(vi.confirmBatchLabel).toBe('Chuyển vào kho')

    expect(en.confirmBatchTitle(1)).toBe('Move 1 file to Personal Vault?')
    expect(en.confirmBatchTitle(3)).toBe('Move 3 files to Personal Vault?')
    expect(en.confirmBatchMessage).toBe(
      'These files will be password-protected and hidden from your regular file list.',
    )
    expect(en.confirmBatchLabel).toBe('Move to Vault')
  })
})
