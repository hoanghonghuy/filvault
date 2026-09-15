import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./SharedWithMeView.vue', import.meta.url)), 'utf-8')

describe('SharedWithMeView accessibility contract', () => {
  it('keeps segmented share tabs touch-accessible', () => {
    const tabRule = source.match(/\.tab-pill\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(tabRule).toContain('min-height: var(--touch-min);')
    expect(source).toContain('role="tablist"')
    expect(source).toContain('role="tab"')
  })

  it('keeps share actions and breadcrumb links touch-accessible', () => {
    const actionRule = source.match(/\.share-action-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    const breadcrumbRule = source.match(/\.browse-path-link\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(actionRule).toContain('width: var(--touch-min);')
    expect(actionRule).toContain('height: var(--touch-min);')
    expect(breadcrumbRule).toContain('min-height: var(--touch-min);')
    expect(breadcrumbRule).toContain('text-overflow: ellipsis;')
    expect(breadcrumbRule).toContain('white-space: nowrap;')
    expect(source).toContain(':aria-label="`${t.download}: ${f.name}`"')
    expect(source).not.toContain(':aria-label="t.download"')
  })
})
