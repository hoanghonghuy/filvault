import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./TrashView.vue', import.meta.url)), 'utf8')

describe('TrashView semantic color contract', () => {
  it('uses shared danger tokens for destructive controls', () => {
    expect(source).toContain('background: var(--danger-soft)')
    expect(source).toContain('color: var(--danger)')
    expect(source).not.toContain('color: #ef4444')
    expect(source).not.toContain('rgba(239, 68, 68')
  })

  it('uses the shared warning token for folder category treatment', () => {
    expect(source).toContain('color: var(--warning)')
    expect(source).toContain('color-mix(in srgb, var(--warning) 14%, transparent)')
    expect(source).not.toContain('color: #f59e0b')
    expect(source).not.toContain('rgba(245, 158, 11')
  })
})
