import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./SharedWithMeView.vue', import.meta.url)), 'utf-8')

describe('SharedWithMeView toolbar contract', () => {
  it('does not expose a fake sorting/filter affordance', () => {
    expect(source).not.toContain('t.sortByTime')
    expect(source).not.toContain('name="filter"')
    expect(source).not.toContain('class="sub-label"')
  })

  it('uses explicit list/grid glyphs and accessible toggle state', () => {
    expect(source).toContain('class="view-mode-glyph"')
    expect(source).toContain(':aria-label="t.viewGrid"')
    expect(source).toContain(':aria-pressed="viewMode === \'grid\'"')
    expect(source).toContain(':title="viewMode === \'list\' ? t.viewGrid : t.viewList"')
    expect(source).not.toContain("viewMode === 'list' ? 'palette' : 'file'")
  })

  it('uses the shared minimum touch target for the view toggle', () => {
    expect(source).toMatch(/\.sub-icon-btn\s*\{[^}]*width:\s*var\(--touch-min\)/s)
    expect(source).toMatch(/\.sub-icon-btn\s*\{[^}]*height:\s*var\(--touch-min\)/s)
  })
})
