import { describe, expect, it } from 'vitest'
import { safeInternalPath } from './safeInternalPath'

describe('safeInternalPath', () => {
  it('keeps in-app paths', () => {
    expect(safeInternalPath('/files')).toBe('/files')
    expect(safeInternalPath('/photos/albums/01ABC')).toBe('/photos/albums/01ABC')
    expect(safeInternalPath('/')).toBe('/')
  })

  it('rejects open redirects', () => {
    expect(safeInternalPath('https://evil.example')).toBe('/')
    expect(safeInternalPath('//evil.example')).toBe('/')
    expect(safeInternalPath('/\\evil.example')).toBe('/')
    expect(safeInternalPath('files')).toBe('/')
  })

  it('falls back when missing', () => {
    expect(safeInternalPath(undefined)).toBe('/')
    expect(safeInternalPath('', '/verify-email')).toBe('/verify-email')
  })
})
