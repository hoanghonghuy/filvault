import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./CallModal.vue', import.meta.url)), 'utf-8')

describe('CallModal accessibility and responsive contract', () => {
  it('exposes canonical mic and camera pressed state', () => {
    expect(source).toContain(':aria-pressed="callStore.isMicEnabled"')
    expect(source).toContain(':aria-pressed="callStore.isCamEnabled"')
  })

  it('keeps connected call controls at least 44px', () => {
    expect(source).toMatch(/\.header-action-btn\s*\{[\s\S]*?width:\s*44px;[\s\S]*?height:\s*44px;/)
    expect(source).toMatch(/\.control-btn\s*\{[\s\S]*?min-width:\s*44px;[\s\S]*?min-height:\s*44px;/)
  })

  it('accounts for device safe areas and mobile landscape', () => {
    expect(source).toContain('env(safe-area-inset-top)')
    expect(source).toContain('env(safe-area-inset-right)')
    expect(source).toContain('env(safe-area-inset-bottom)')
    expect(source).toContain('env(safe-area-inset-left)')
    expect(source).toContain('@media (max-width: 767px) and (orientation: landscape)')
  })

  it('disables non-essential call motion when reduced motion is requested', () => {
    expect(source).toContain('@media (prefers-reduced-motion: reduce)')
    expect(source).toMatch(/prefers-reduced-motion: reduce[\s\S]*?animation:\s*none;/)
    expect(source).toMatch(/prefers-reduced-motion: reduce[\s\S]*?transition:\s*none;/)
  })
})
