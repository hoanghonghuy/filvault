import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView accessibility localization wiring', () => {
  it('uses reactive locale copy for navigation and item action labels', () => {
    expect(source).toContain(':aria-label="t.fileViewsAria"')
    expect(source).toContain(':aria-label="t.fileHierarchyAria"')
    expect(source).toContain(':aria-label="`${t.openFolderAria}: ${folder.name}`"')
    expect(source).toContain(':aria-label="`${t.folderActionsAria}: ${folder.name}`"')
    expect(source).toContain(':aria-label="`${t.fileActionsAria}: ${file.name}`"')
  })

  it('uses the shared reactive locale key for the Files search placeholder without a hardcoded fallback', () => {
    expect(source).toContain(':placeholder="t.searchInFilvault"')
    expect(source).not.toContain("t.searchInFilvault || 'Tìm kiếm trong Filvault…'")
  })

  it('does not retain the audited English-only accessibility labels', () => {
    expect(source).not.toContain('aria-label="File views"')
    expect(source).not.toContain('aria-label="View hierarchy"')
    expect(source).not.toContain('aria-label="Open folder"')
    expect(source).not.toContain('aria-label="Folder actions"')
    expect(source).not.toContain('aria-label="File actions"')
  })
})
