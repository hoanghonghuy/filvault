import { describe, expect, it } from 'vitest'
import { formatVaultFileCount } from './vaultFileCount'

describe('formatVaultFileCount', () => {
  it('uses English singular and plural grammar', () => {
    expect(formatVaultFileCount(0, 'en')).toBe('0 files')
    expect(formatVaultFileCount(1, 'en')).toBe('1 file')
    expect(formatVaultFileCount(2, 'en')).toBe('2 files')
  })

  it('keeps the Vietnamese unit count-invariant', () => {
    expect(formatVaultFileCount(0, 'vi')).toBe('0 tệp')
    expect(formatVaultFileCount(1, 'vi')).toBe('1 tệp')
    expect(formatVaultFileCount(2, 'vi')).toBe('2 tệp')
  })
})
