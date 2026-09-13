import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { MIME_CATEGORY_COLORS, mimeCategoryColor } from '@/lib/mimeColors'

const source = readFileSync(fileURLToPath(new URL('./SharedWithMeView.vue', import.meta.url)), 'utf-8')

describe('SharedWithMeView MIME color contract', () => {
  it('uses the shared MIME/category color helper instead of a duplicate local map', () => {
    expect(source).toContain("import { mimeCategoryColor } from '@/lib/mimeColors'")
    expect(source).not.toContain('function getFileTypeColor(')
    expect(source.match(/mimeCategoryColor\(undefined, /g)).toHaveLength(4)
    expect(source.match(/mimeCategoryColor\(f\.mimeType, f\.name\)/g)).toHaveLength(2)
  })

  it('keeps representative filename categories aligned with the shared palette', () => {
    expect(mimeCategoryColor(undefined, 'photo.heic')).toBe(MIME_CATEGORY_COLORS.image)
    expect(mimeCategoryColor(undefined, 'sheet.ods')).toBe(MIME_CATEGORY_COLORS.spreadsheet)
    expect(mimeCategoryColor(undefined, 'notes.md')).toBe(MIME_CATEGORY_COLORS.document)
  })
})
