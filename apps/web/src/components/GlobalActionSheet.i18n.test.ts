import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./GlobalActionSheet.vue', import.meta.url)), 'utf-8')

describe('GlobalActionSheet localization contract', () => {
  it('resolves unambiguous actions from stable ids instead of English display labels', () => {
    expect(source).toContain('const actionLabelsById')
    expect(source).toContain('open: t.value.open')
    expect(source).toContain('rename: t.value.rename')
    expect(source).toContain('download: t.value.download')
    expect(source).toContain('restore: t.value.restore')
    expect(source).toContain('actionLabelsById.value[item.id]')
  })

  it('keeps a narrow fallback only for context-dependent legacy actions', () => {
    expect(source).toContain('const legacyActionLabels')
    expect(source).toContain("'Add to favorites': t.value.addToFavorites")
    expect(source).toContain("'Remove from favorites': t.value.removeFromFavorites")
    expect(source).not.toContain("'Open': t.value.open")
    expect(source).not.toContain("'Rename': t.value.rename")
    expect(source).not.toContain("'Download': t.value.download")
    expect(source).not.toContain("'Restore': t.value.restore")
  })
})
