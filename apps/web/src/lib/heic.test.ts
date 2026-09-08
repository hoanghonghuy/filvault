import { describe, expect, it } from 'vitest'
import { isHeic, isHeicFile } from './heic'

describe('heic utility', () => {
  it('correctly identifies heic/heif file names (case-insensitive)', () => {
    expect(isHeic('photo.heic')).toBe(true)
    expect(isHeic('photo.HEIC')).toBe(true)
    expect(isHeic('image.heif')).toBe(true)
    expect(isHeic('image.HEIF')).toBe(true)
    expect(isHeic('/path/to/IMG_2026.Heic')).toBe(true)
  })

  it('correctly identifies heic/heif mime types', () => {
    expect(isHeic(null, 'image/heic')).toBe(true)
    expect(isHeic(null, 'image/heif')).toBe(true)
    expect(isHeic(null, 'image/heic-sequence')).toBe(true)
    expect(isHeic(null, 'image/heif-sequence')).toBe(true)
    expect(isHeic('photo.jpg', 'image/heic')).toBe(true)
  })

  it('returns false for standard image and non-image types', () => {
    expect(isHeic('photo.jpg')).toBe(false)
    expect(isHeic('photo.jpeg', 'image/jpeg')).toBe(false)
    expect(isHeic('photo.png', 'image/png')).toBe(false)
    expect(isHeic('photo.webp', 'image/webp')).toBe(false)
    expect(isHeic('archive.zip', 'application/zip')).toBe(false)
    expect(isHeic('heic_notes.txt', 'text/plain')).toBe(false)
    expect(isHeic(null, null)).toBe(false)
    expect(isHeic('', '')).toBe(false)
  })

  it('identifies File/Blob objects', () => {
    const heicFile = new File(['mock'], 'camera.heic', { type: 'image/heic' })
    expect(isHeicFile(heicFile)).toBe(true)

    const heicBlob = new Blob(['mock'], { type: 'image/heic' })
    expect(isHeicFile(heicBlob)).toBe(true)

    const pngFile = new File(['mock'], 'sample.png', { type: 'image/png' })
    expect(isHeicFile(pngFile)).toBe(false)
  })
})
