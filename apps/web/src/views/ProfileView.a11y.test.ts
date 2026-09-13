import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./ProfileView.vue', import.meta.url)), 'utf-8')

describe('Profile avatar action accessibility contract', () => {
  it('keeps edit and remove actions at the shared touch minimum', () => {
    expect(source).toMatch(/\.avatar-action-btn\s*\{[\s\S]*?width:\s*var\(--touch-min\)[\s\S]*?height:\s*var\(--touch-min\)/)
    expect(source).toMatch(/\.remove-avatar-btn\s*\{[\s\S]*?min-height:\s*var\(--touch-min\)/)
  })

  it('provides keyboard focus and disabled-state affordances', () => {
    expect(source).toMatch(/\.avatar-action-btn:focus-visible,[\s\S]*?\.remove-avatar-btn:focus-visible/)
    expect(source).toMatch(/outline:\s*3px solid var\(--accent\)/)
    expect(source).toMatch(/\.avatar-action-btn:disabled,[\s\S]*?\.remove-avatar-btn:disabled/)
  })

  it('does not animate the avatar control for reduced-motion users', () => {
    expect(source).toMatch(/@media\s*\(prefers-reduced-motion:\s*reduce\)/)
    expect(source).toMatch(/\.avatar-action-btn\s*\{[\s\S]*?transition:\s*none/)
  })
})
