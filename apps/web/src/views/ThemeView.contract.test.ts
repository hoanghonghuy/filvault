/**
 * @vitest-environment node
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./ThemeView.vue', import.meta.url)), 'utf-8')

describe('ThemeView navigation contract', () => {
  it('does not expose dead tab navigation while keeping primary theme controls', () => {
    expect(source).not.toContain('role="tablist"')
    expect(source).not.toContain('role="tab"')
    expect(source).not.toContain('activeTab')
    expect(source).toContain('class="theme-header-title"')
    expect(source).toContain('role="radiogroup"')
    expect(source).toContain('class="swatches-grid"')
    expect(source).toContain('class="seasonal-grid"')
  })
})
