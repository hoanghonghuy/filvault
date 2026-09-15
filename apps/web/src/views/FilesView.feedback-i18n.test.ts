import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView core feedback localization wiring', () => {
  it('uses locale-reactive copy for browse, create-folder and download feedback', () => {
    expect(source).toContain('formatApiError(e, t.value.filesLoadFailed)')
    expect(source).toContain('ui.showToast(t.value.folderCreated)')
    expect(source).toContain('formatApiError(e, t.value.folderCreateFailed)')
    expect(source).toContain('formatApiError(e, t.value.fileDownloadFailed)')
  })

  it('does not retain the replaced English-only fallback literals', () => {
    expect(source).not.toContain("formatApiError(e, 'Failed to load files')")
    expect(source).not.toContain("ui.showToast('Folder created')")
    expect(source).not.toContain("formatApiError(e, 'Failed to create folder')")
    expect(source).not.toContain("formatApiError(e, 'Download failed')")
  })
})
