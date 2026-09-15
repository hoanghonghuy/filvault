import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')
const i18nSource = readFileSync(fileURLToPath(new URL('../lib/i18n.ts', import.meta.url)), 'utf8')

describe('FilesView preview and search fallback localization', () => {
  it('uses locale-reactive copy for preview and search failures', () => {
    expect(source).toContain('formatApiError(e, t.value.filesPreviewFailed)')
    expect(source).toContain('formatApiError(e, t.value.filesSearchFailed)')
  })

  it('keeps bounded VI/EN fallback copy in the shared dictionary', () => {
    expect(i18nSource).toContain("filesPreviewFailed: 'Không thể xem trước tệp'")
    expect(i18nSource).toContain("filesSearchFailed: 'Không thể tìm kiếm tệp'")
    expect(i18nSource).toContain("filesPreviewFailed: 'Could not preview file'")
    expect(i18nSource).toContain("filesSearchFailed: 'Search failed'")
  })

  it('does not retain English-only fallback literals in FilesView', () => {
    expect(source).not.toContain("formatApiError(e, 'Could not preview file')")
    expect(source).not.toContain("formatApiError(e, 'Search failed')")
  })
})
