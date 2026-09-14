import { describe, expect, it } from 'vitest'
import { formatSearchResultCount } from './searchResultCount'

describe('formatSearchResultCount', () => {
  it('uses English singular and plural grammar', () => {
    expect(formatSearchResultCount(0, 'en')).toBe('0 results')
    expect(formatSearchResultCount(1, 'en')).toBe('1 result')
    expect(formatSearchResultCount(2, 'en')).toBe('2 results')
  })

  it('keeps Vietnamese count-invariant', () => {
    expect(formatSearchResultCount(0, 'vi')).toBe('0 kết quả')
    expect(formatSearchResultCount(1, 'vi')).toBe('1 kết quả')
    expect(formatSearchResultCount(2, 'vi')).toBe('2 kết quả')
  })
})
