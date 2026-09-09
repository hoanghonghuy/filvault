/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it } from 'vitest'
import { THEMES, useTheme } from './theme'

describe('multi-theme system contract', () => {
  beforeEach(() => {
    localStorage.clear()
    delete document.documentElement.dataset.theme
    delete document.documentElement.dataset.colorTheme
  })

  it('defines the complete palette of themes matching mobile presets', () => {
    const ids = THEMES.map((t) => t.id)
    expect(ids).toContain('default')
    expect(ids).toContain('cyan')
    expect(ids).toContain('teal')
    expect(ids).toContain('sage')
    expect(ids).toContain('sunshine')
    expect(ids).toContain('peach')
    expect(ids).toContain('lavender')
    expect(ids).toContain('pearl')
    expect(ids).toContain('pebble')
    expect(ids).toContain('dark')
    expect(ids).toContain('material')
    expect(ids).toContain('spring')
    expect(ids).toContain('summer')
    expect(ids).toContain('autumn')
    expect(ids).toContain('winter')
  })

  it('applies color theme to dataset and localStorage', () => {
    const { applyColorTheme, currentColorTheme } = useTheme()
    applyColorTheme('peach')

    expect(currentColorTheme.value).toBe('peach')
    expect(localStorage.getItem('filvault.colorTheme')).toBe('peach')
    expect(document.documentElement.dataset.colorTheme).toBe('peach')
  })

  it('activates dark mode when dark theme is applied', () => {
    const { applyColorTheme } = useTheme()
    applyColorTheme('dark')

    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(localStorage.getItem('filvault.theme')).toBe('dark')
  })

  it('persists followSystemDark state', () => {
    const { setFollowSystemDark, followSystemDark } = useTheme()
    setFollowSystemDark(true)
    expect(followSystemDark.value).toBe(true)
    expect(localStorage.getItem('filvault.followSystemDark')).toBe('true')

    setFollowSystemDark(false)
    expect(followSystemDark.value).toBe(false)
    expect(localStorage.getItem('filvault.followSystemDark')).toBe('false')
  })

  it('does not mark any theme as premium until entitlements exist', () => {
    for (const theme of THEMES) {
      expect(theme).not.toHaveProperty('isPro')
    }
  })

  it('applies seasonal themes without entitlement checks', () => {
    const { applyColorTheme, currentColorTheme } = useTheme()
    applyColorTheme('spring')

    expect(currentColorTheme.value).toBe('spring')
    expect(localStorage.getItem('filvault.colorTheme')).toBe('spring')
    expect(document.documentElement.dataset.colorTheme).toBe('spring')
  })
})
