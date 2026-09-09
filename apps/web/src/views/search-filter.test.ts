import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('search filter sheet contract', () => {
  const sheet = readSrc('../components/SearchFilterSheet.vue')

  it('is built on the shared BottomSheet', () => {
    expect(sheet).toMatch(/import\s+BottomSheet\s+from/)
    expect(sheet).toContain('<BottomSheet')
  })

  it('offers every spec type option and emits the selection on apply', () => {
    for (const value of ['all', 'image', 'video', 'document', 'archive', 'folder']) {
      expect(sheet).toMatch(new RegExp(`['"]${value}['"]`))
    }
    expect(sheet).toMatch(/emit\('apply'/)
  })

  it('offers sort options with asc/desc order controls', () => {
    for (const value of ['relevance', 'name', 'date', 'size']) {
      expect(sheet).toMatch(new RegExp(`['"]${value}['"]`))
    }
    expect(sheet).toMatch(/['"]asc['"]/)
    expect(sheet).toMatch(/['"]desc['"]/)
  })

  it('has date range inputs feeding from/to', () => {
    expect(sheet).toMatch(/type="date"/)
    const script = sheet.match(/<script[\s\S]*?<\/script>/)?.[0] ?? ''
    expect(script).toMatch(/\bfrom\b/)
    expect(script).toMatch(/\bto\b/)
  })
})

describe('FilesView filtered search contract', () => {
  const files = readSrc('../views/FilesView.vue')

  it('opens a SearchFilterSheet from a toolbar button', () => {
    expect(files).toMatch(/import\s+SearchFilterSheet\s+from/)
    expect(files).toContain('<SearchFilterSheet')
    expect(files).toMatch(/aria-label="Search filters"/)
  })

  it('sends type, folderId, from, to, sort and order to the API', () => {
    const fn = files.match(/function buildSearchQuery[\s\S]*?\n\}/)?.[0] ?? ''
    for (const param of ['type', 'folderId', 'from', 'to', 'sort', 'order']) {
      expect(fn).toContain(`'${param}'`)
    }
    expect(files).toMatch(/runSearch/)
  })

  it('re-runs the search when filters change', () => {
    expect(files).toMatch(/watch\(filters/)
  })

  it('shows removable chips for active filters plus a clear-all control', () => {
    expect(files).toMatch(/class="filter-chips"/)
    expect(files).toMatch(/aria-label="Remove filter"/)
    expect(files).toMatch(/Clear all/)
  })

  it('labels filtered empty results and offers clearing filters', () => {
    expect(files).toContain('No results match your filters')
    expect(files).toMatch(/clearFilters/)
  })

  it('keeps the animated row list while showing a filtered caption', () => {
    expect(files).toMatch(/TransitionGroup[^>]*name="row"/)
    expect(files).toMatch(/results · filtered|results·filtered|· filtered/)
  })

  it('exposes browse item selection with aria-pressed on role=button cards', () => {
    expect(files).toMatch(/class="file-item-card[\s\S]*?:aria-pressed="isSelecting \? selectedFolderIds\.has\(folder\.id\) : undefined"/)
    expect(files).toMatch(/class="file-item-card[\s\S]*?:aria-pressed="isSelecting \? selectedFileIds\.has\(file\.id\) : undefined"/)
    const browseCards = files.match(/class="file-item-card tappable"[\s\S]*?@keydown="onFileItemKeydown/g) ?? []
    expect(browseCards.length).toBeGreaterThanOrEqual(1)
    for (const card of browseCards) {
      expect(card).not.toContain('aria-selected')
    }
  })

  it('routes the personal vault shortcut through RouterLink', () => {
    expect(files).toMatch(/<RouterLink[\s\S]*?class="vault-entry-card/)
    expect(files).toContain('to="/vault"')
  })
})

describe('search filters typing contract', () => {
  const types = readSrc('../api/types.ts')

  it('exports the SearchFilters shape used by view and sheet', () => {
    expect(types).toMatch(/export interface SearchFilters \{[\s\S]*?type:[\s\S]*?\}/)
    expect(types).toMatch(/export interface SearchFilters \{[\s\S]*?sort:[\s\S]*?\}/)
    expect(types).toMatch(/export interface SearchFilters \{[\s\S]*?order:[\s\S]*?\}/)
    expect(types).toMatch(/export interface SearchFilters \{[\s\S]*?from\?: string[\s\S]*?\}/)
    expect(types).toMatch(/export interface SearchFilters \{[\s\S]*?to\?: string[\s\S]*?\}/)
  })
})
