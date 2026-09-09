import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('FilesView vault entry shortcut contract', () => {
  const files = readSrc('./FilesView.vue')

  it('navigates to /vault via RouterLink, not the favorites segment', () => {
    const card = files.match(/<!-- TeraBox Root Vault Shortcut -->[\s\S]*?<\/RouterLink>/)?.[0] ?? ''
    expect(card).toContain('<RouterLink')
    expect(card).toContain('to="/vault"')
    expect(card).not.toContain("segment = 'favorites'")
    expect(card).not.toContain('@click')
  })

  it('labels the card as personal vault with vault-specific secondary copy', () => {
    const card = files.match(/<!-- TeraBox Root Vault Shortcut -->[\s\S]*?<\/RouterLink>/)?.[0] ?? ''
    expect(card).toContain('t.personalVault')
    expect(card).toContain('t.vaultSubtitle')
    expect(card).not.toContain('t.favorites')
  })

  it('exposes an accessible name and keyboard-focus styling for the vault destination', () => {
    const card = files.match(/<!-- TeraBox Root Vault Shortcut -->[\s\S]*?<\/RouterLink>/)?.[0] ?? ''
    expect(card).toMatch(/:aria-label="t\.personalVault"/)
    expect(files).toMatch(/\.vault-entry-card:focus-visible/)
  })
})
