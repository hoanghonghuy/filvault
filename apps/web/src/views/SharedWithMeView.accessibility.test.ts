import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./SharedWithMeView.vue', import.meta.url)), 'utf-8')

describe('SharedWithMeView accessibility contract', () => {
  it('keeps share actions and breadcrumb links touch-accessible', () => {
    const actionRule = source.match(/\.share-action-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    const breadcrumbRule = source.match(/\.browse-path-link\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(actionRule).toContain('width: var(--touch-min);')
    expect(actionRule).toContain('height: var(--touch-min);')
    expect(breadcrumbRule).toContain('min-height: var(--touch-min);')
    expect(breadcrumbRule).toContain('text-overflow: ellipsis;')
    expect(breadcrumbRule).toContain('white-space: nowrap;')
  })
})
