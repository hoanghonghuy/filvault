import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./OverviewView.vue', import.meta.url)), 'utf-8')

describe('OverviewView accessibility contract', () => {
  it('keeps segmented overview tabs touch-accessible', () => {
    const tabRule = source.match(/\.tab-pill\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(tabRule).toContain('min-height: var(--touch-min);')
    expect(source).toContain('role="tablist"')
    expect(source).toContain('role="tab"')
    expect(source).toContain(`:aria-selected="activeTab === 'recent'"`)
    expect(source).toContain(`:aria-selected="activeTab === 'favorites'"`)
  })
})
