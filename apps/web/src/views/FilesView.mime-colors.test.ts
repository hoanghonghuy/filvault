import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { MIME_CATEGORY_COLORS, mimeCategoryColor } from '@/lib/mimeColors'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf-8')

describe('FilesView MIME color contract', () => {
  it('uses the shared MIME/category helper instead of a duplicate local map', () => {
    expect(source).toContain("import { mimeCategoryColor } from '@/lib/mimeColors'")
    expect(source).not.toContain('function getFileTypeColor(')
    expect(source.match(/mimeCategoryColor\(file\.mimeType, file\.name\)/g)).toHaveLength(4)
  })

  it('keeps MIME-first categorization with filename fallback', () => {
    expect(mimeCategoryColor('image/jpeg', 'unknown.bin')).toBe(MIME_CATEGORY_COLORS.image)
    expect(mimeCategoryColor('', 'archive.7z')).toBe(MIME_CATEGORY_COLORS.archive)
    expect(mimeCategoryColor('', 'notes.md')).toBe(MIME_CATEGORY_COLORS.document)
  })
})
