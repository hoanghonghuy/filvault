import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { MIME_CATEGORY_COLORS } from '@/lib/mimeColors'

const source = readFileSync(fileURLToPath(new URL('./OverviewView.vue', import.meta.url)), 'utf-8')

describe('OverviewView category color contract', () => {
  it('uses the shared MIME category palette for file/media shortcuts', () => {
    expect(source).toContain("import { MIME_CATEGORY_COLORS } from '@/lib/mimeColors'")
    expect(source).toContain('color: MIME_CATEGORY_COLORS.folder')
    expect(source).toContain('color: MIME_CATEGORY_COLORS.image')
    expect(source).toContain('color: MIME_CATEGORY_COLORS.video')
  })

  it('keeps canonical shared category colors stable', () => {
    expect(MIME_CATEGORY_COLORS.folder).toBe('#f59e0b')
    expect(MIME_CATEGORY_COLORS.image).toBe('#8b5cf6')
    expect(MIME_CATEGORY_COLORS.video).toBe('#ec4899')
  })
})
