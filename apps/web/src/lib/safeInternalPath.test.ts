import { describe, expect, it } from 'vitest'
import { safeInternalPath } from './safeInternalPath'

describe('safeInternalPath', () => {
  it('keeps in-app paths', () => {
    expect(safeInternalPath('/files')).toBe('/files')
    expect(safeInternalPath('/photos/albums/01ABC')).toBe('/photos/albums/01ABC')
  })

  it('rejects open redirects', () => {
    expect(safeInternalPath('https://evil.example')).toBe('/files')
    expect(safeInternalPath('//evil.example')).toBe('/files')
    expect(safeInternalPath('/\\evil.example')).toBe('/files')
    expect(safeInternalPath('files')).toBe('/files')
  })

  it('falls back when missing', () => {
    expect(safeInternalPath(undefined, '/files')).toBe('/files')
    expect(safeInternalPath('', '/verify-email')).toBe('/verify-email')
  })
})
