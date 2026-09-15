import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf-8')

describe('Files Personal Vault design contract', () => {
  it('uses non-destructive app accent semantics for the Vault shortcut', () => {
    const match = source.match(/\.vault-icon-box\s*\{([\s\S]*?)\n\}/)
    expect(match).not.toBeNull()
    const block = match?.[1] ?? ''
    expect(block).toContain('background: var(--accent-soft)')
    expect(block).toContain('color: var(--accent)')
    expect(block).not.toContain('#ef4444')
    expect(block).not.toContain('rgba(239, 68, 68')
  })

  it('keeps the Personal Vault shortcut routed to Vault', () => {
    expect(source).toMatch(/class="vault-entry-card(?:\s+[^"]*)?"/)
    expect(source).toContain('to="/vault"')
  })
})
