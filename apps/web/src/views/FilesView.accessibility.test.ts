import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf-8')

describe('FilesView accessibility contract', () => {
  it('keeps row action buttons touch-accessible', () => {
    const actionRule = source.match(/\.file-item-more-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(actionRule).toContain('width: var(--touch-min);')
    expect(actionRule).toContain('height: var(--touch-min);')
    expect(actionRule).toContain('flex-shrink: 0;')
  })

  it('keeps sub-toolbar icon controls touch-accessible', () => {
    const toolbarRule = source.match(/\.sub-icon-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(toolbarRule).toContain('width: var(--touch-min);')
    expect(toolbarRule).toContain('height: var(--touch-min);')
  })
  it('keeps search-clear control touch-accessible', () => {
    const clearRule = source.match(/\.files-search-clear\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(clearRule).toContain('width: var(--touch-min);')
    expect(clearRule).toContain('height: var(--touch-min);')
  })

})
