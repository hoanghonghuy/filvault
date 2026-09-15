import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView favorites feedback localization wiring', () => {
  it('uses locale-reactive copy for batch, load and single-file favorite feedback', () => {
    expect(source).toContain("ui.showToast(t.value.favoriteNoneSelected, 'info')")
    expect(source).toContain("t.value.favoriteBatchAdded.replace('{n}', String(fileIds.length))")
    expect(source).toContain('formatApiError(e, t.value.favoriteBatchFailed)')
    expect(source).toContain('formatApiError(e, t.value.favoritesLoadFailed)')
    expect(source).toContain("t.value.favoriteRemoved.replace('{name}', () => name)")
    expect(source).toContain("t.value.favoriteAdded.replace('{name}', () => name)")
    expect(source).toContain('formatApiError(e, t.value.favoriteUpdateFailed)')
  })

  it('uses callback replacement so dollar sequences in filenames stay literal', () => {
    const name = 'budget-$&-$1.txt'
    const template = 'Added "{name}" to favorites'
    expect(template.replace('{name}', () => name)).toBe('Added "budget-$&-$1.txt" to favorites')
  })

  it('does not retain the replaced English-only favorites feedback', () => {
    expect(source).not.toContain("ui.showToast('No files selected to favorite', 'info')")
    expect(source).not.toContain('Added ${fileIds.length} files to favorites')
    expect(source).not.toContain("formatApiError(e, 'Failed to favorite items')")
    expect(source).not.toContain("formatApiError(e, 'Failed to load favorites')")
    expect(source).not.toContain('from favorites`)')
    expect(source).not.toContain('to favorites`)')
    expect(source).not.toContain("formatApiError(e, 'Failed to update favorite')")
  })
})
