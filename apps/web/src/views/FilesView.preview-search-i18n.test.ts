import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView preview/search fallback localization wiring', () => {
  it('uses locale-reactive fallback copy while preserving formatApiError', () => {
    expect(source).toContain('formatApiError(e, t.value.filePreviewFailed)')
    expect(source).toContain('formatApiError(e, t.value.filesSearchFailed)')
  })

  it('does not retain the replaced English-only fallback literals', () => {
    expect(source).not.toContain("formatApiError(e, 'Could not preview file')")
    expect(source).not.toContain("formatApiError(e, 'Search failed')")
  })
})
