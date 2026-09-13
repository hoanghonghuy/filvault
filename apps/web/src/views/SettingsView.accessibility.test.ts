/// <reference types="node" />

import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./SettingsView.vue', import.meta.url)), 'utf-8')

describe('SettingsView accessibility contract', () => {
  it('keeps quick theme swatches at least 44px and lets the grid wrap on narrow screens', () => {
    const gridRule = source.match(/\.quick-swatches-grid\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    const swatchRule = source.match(/\.quick-swatch\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(gridRule).toContain('grid-template-columns: repeat(auto-fit, minmax(44px, 1fr));')
    expect(gridRule).not.toContain('repeat(6, 1fr)')
    expect(swatchRule).toContain('width: 44px;')
    expect(swatchRule).toContain('height: 44px;')
  })

  it('keeps Profile navigation localized and at least 44px tappable', () => {
    expect(source).toContain('to="/profile"')
    expect(source).toContain(':title="t.profile"')
    expect(source).toContain(':aria-label="t.profile"')

    const rule = source.match(/\.profile-arrow-link\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(rule).toContain('width: 44px;')
    expect(rule).toContain('height: 44px;')
    expect(rule).not.toContain('width: 36px;')
    expect(rule).not.toContain('height: 36px;')
  })
})
