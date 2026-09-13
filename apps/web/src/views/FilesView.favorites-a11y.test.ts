import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView favorites keyboard accessibility', () => {
  it('makes favorite rows focusable, named action controls', () => {
    const favoritesSection = source.slice(source.indexOf(`v-else-if="segment === 'favorites'"`))
    expect(favoritesSection).toContain('role="button"')
    expect(favoritesSection).toContain('tabindex="0"')
    expect(favoritesSection).toContain(':aria-label="file.name"')
    expect(favoritesSection).toContain('@keydown="onFavoriteFileKeydown($event, file)"')
  })

  it('activates favorites with Enter or Space while ignoring nested controls', () => {
    expect(source).toContain('if (event.target !== event.currentTarget) return')
    expect(source).toContain("if (event.key !== 'Enter' && event.key !== ' ' && event.key !== 'Spacebar') return")
    expect(source).toContain('void openFileActions(file)')
    expect(source).toContain('@click.stop="openFileActions(file)"')
  })
})
