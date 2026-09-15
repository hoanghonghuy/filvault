import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')
const searchSection = source.slice(source.indexOf(`v-if="segment === 'all' && !loading && searchResults"`), source.indexOf('<!-- TeraBox Root Vault Shortcut -->'))

describe('FilesView search result keyboard accessibility', () => {
  it('makes folder and file result rows focusable named controls', () => {
    expect(searchSection.match(/role="button"/g)).toHaveLength(2)
    expect(searchSection.match(/tabindex="0"/g)).toHaveLength(2)
    expect(searchSection).toContain(':aria-label="folder.name"')
    expect(searchSection).toContain(':aria-label="file.name"')
    expect(searchSection).toContain('@keydown="onSearchFolderKeydown($event, folder.id)"')
    expect(searchSection).toContain('@keydown="onSearchFileKeydown($event, file)"')
  })

  it('supports Enter and Space without duplicating nested button activation', () => {
    expect(source).toContain("return event.key === 'Enter' || event.key === ' ' || event.key === 'Spacebar'")
    expect(source).toContain('event.target !== event.currentTarget || !isPrimaryKeyboardActivation(event)')
    expect(searchSection).toContain('@click.stop="openFolder(folder.id)"')
    expect(searchSection).toContain('@click.stop="openFileActions(file)"')
  })
})
