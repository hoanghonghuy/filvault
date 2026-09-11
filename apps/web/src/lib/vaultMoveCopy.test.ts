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
})
