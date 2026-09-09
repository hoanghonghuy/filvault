import { describe, expect, it } from 'vitest'
import { userInitials } from './userInitials'

describe('userInitials', () => {
  it('uses first letters of up to two name parts', () => {
    expect(userInitials('Hong Huy')).toBe('HH')
    expect(userInitials('Ada')).toBe('A')
  })

  it('falls back to the first letter of email when name is empty', () => {
    expect(userInitials('', 'you@example.com')).toBe('Y')
  })

  it('falls back to a single letter when nothing useful exists', () => {
    expect(userInitials('  ', '')).toBe('U')
  })
})
