import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./CallModal.vue', import.meta.url)), 'utf-8')

describe('CallModal design token contract', () => {
  it('uses shared danger semantics for destructive and muted call controls', () => {
    expect(source).toMatch(/\.btn-decline\s*\{[\s\S]*?background:\s*var\(--danger\);/)
    expect(source).toMatch(/\.control-btn\.muted\s*\{[\s\S]*?background:\s*var\(--danger\);/)
    expect(source).toMatch(/\.control-btn\.btn-hangup\s*\{[\s\S]*?background:\s*var\(--danger\);/)
    expect(source).toContain('color-mix(in srgb, var(--danger) 45%, transparent)')
    expect(source).toContain('color-mix(in srgb, var(--danger) 50%, transparent)')
    expect(source).not.toContain('background: #ef4444')
    expect(source).not.toContain('rgba(239, 68, 68')
  })

  it('preserves high-contrast call control icons', () => {
    expect(source).toMatch(/\.call-btn\s*\{[\s\S]*?color:\s*#ffffff;/)
    expect(source).toMatch(/\.control-btn\.muted\s*\{[\s\S]*?color:\s*#ffffff;/)
    expect(source).toMatch(/\.control-btn\.btn-hangup\s*\{[\s\S]*?color:\s*#ffffff;/)
  })
})
