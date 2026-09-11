import { describe, expect, it } from 'vitest'
import { MIME_CATEGORY_COLORS, mimeCategoryColor, mimeColorCategory } from './mimeColors'

describe('mimeColors', () => {
  it('maps representative MIME types to stable product categories', () => {
    expect(mimeColorCategory('image/png')).toBe('image')
    expect(mimeColorCategory('video/mp4')).toBe('video')
    expect(mimeColorCategory('audio/mpeg')).toBe('audio')
    expect(mimeColorCategory('application/pdf')).toBe('pdf')
    expect(mimeColorCategory('application/zip')).toBe('archive')
    expect(mimeColorCategory('application/vnd.ms-excel')).toBe('spreadsheet')
    expect(mimeColorCategory('application/msword')).toBe('document')
  })

  it('falls back to filename extension when MIME data is missing or generic', () => {
    expect(mimeColorCategory(undefined, 'report.pdf')).toBe('pdf')
    expect(mimeColorCategory('', 'budget.xlsx')).toBe('spreadsheet')
    expect(mimeColorCategory('application/octet-stream', 'notes.docx')).toBe('document')
    expect(mimeColorCategory(undefined, 'backup.7z')).toBe('archive')
  })

  it('keeps unknown and generic files deterministic', () => {
    expect(mimeColorCategory()).toBe('unknown')
    expect(mimeColorCategory(undefined, 'README')).toBe('file')
    expect(mimeCategoryColor('image/jpeg')).toBe(MIME_CATEGORY_COLORS.image)
    expect(mimeCategoryColor()).toBe(MIME_CATEGORY_COLORS.unknown)
  })
})
