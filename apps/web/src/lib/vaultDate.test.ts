import { describe, expect, it } from 'vitest'
import { formatVaultDate } from './vaultDate'

describe('formatVaultDate', () => {
  const timestamp = '2026-09-13T12:00:00.000Z'

  it('uses the explicit Vietnamese locale', () => {
    expect(formatVaultDate(timestamp, 'vi')).toBe(
      new Date(timestamp).toLocaleDateString('vi-VN', { year: 'numeric', month: 'short', day: 'numeric' }),
    )
  })

  it('uses the explicit English locale', () => {
    expect(formatVaultDate(timestamp, 'en')).toBe(
      new Date(timestamp).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' }),
    )
  })

  it('does not render bogus text for empty or invalid timestamps', () => {
    expect(formatVaultDate('', 'vi')).toBe('')
    expect(formatVaultDate('not-a-date', 'en')).toBe('')
  })
})
