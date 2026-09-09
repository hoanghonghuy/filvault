/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  APPEARANCE_MODE_STORAGE_KEY,
  COLOR_THEME_STORAGE_KEY,
  THEMES,
  applySystemAppearanceIfNeeded,
  hydrateAppearance,
  migrateAppearanceMode,
  resolveAppearanceIsDark,
  syncAppearanceFromStorage,
  useTheme,
} from './theme'

function mockMatchMedia(matches = false) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    configurable: true,
    value: vi.fn<(query: string) => MediaQueryList>().mockImplementation(
      (query: string) =>
        ({
          matches,
          media: query,
          addEventListener: vi.fn<(type: string, listener: (event: MediaQueryListEvent) => void) => void>(),
          removeEventListener: vi.fn<(type: string, listener: (event: MediaQueryListEvent) => void) => void>(),
        }) as MediaQueryList,
    ),
  })
}

describe('multi-theme system contract', () => {
  beforeEach(() => {
    localStorage.clear()
    delete document.documentElement.dataset.theme
    delete document.documentElement.dataset.colorTheme
    mockMatchMedia(false)
    syncAppearanceFromStorage()
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
    expect(ids).not.toContain('dark')
    expect(ids).toContain('material')
    expect(ids).toContain('spring')
    expect(ids).toContain('summer')
    expect(ids).toContain('autumn')
    expect(ids).toContain('winter')
  })

  it('applies color theme to dataset and localStorage without changing appearance mode', () => {
    const { applyColorTheme, currentColorTheme, appearanceMode } = useTheme()
    setAppearanceModeLight()
    applyColorTheme('peach')

    expect(currentColorTheme.value).toBe('peach')
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('peach')
    expect(document.documentElement.dataset.colorTheme).toBe('peach')
    expect(appearanceMode.value).toBe('light')
    expect(document.documentElement.dataset.theme).toBeUndefined()
  })

  it('persists appearance mode separately from color theme', () => {
    const { setAppearanceMode, appearanceMode, applyColorTheme } = useTheme()
    applyColorTheme('lavender')
    setAppearanceMode('dark')

    expect(appearanceMode.value).toBe('dark')
    expect(localStorage.getItem(APPEARANCE_MODE_STORAGE_KEY)).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('lavender')
    expect(document.documentElement.dataset.colorTheme).toBe('lavender')
  })

  it('supports system appearance mode and updates when OS preference changes', () => {
    mockMatchMedia(false)

    const { setAppearanceMode, appearanceMode, resolvedIsDark } = useTheme()
    setAppearanceMode('system')

    expect(appearanceMode.value).toBe('system')
    expect(localStorage.getItem(APPEARANCE_MODE_STORAGE_KEY)).toBe('system')
    expect(resolvedIsDark.value).toBe(false)
    expect(document.documentElement.dataset.theme).toBeUndefined()

    mockMatchMedia(true)
    applySystemAppearanceIfNeeded()

    expect(resolvedIsDark.value).toBe(true)
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('migrates legacy dark color theme into dark appearance with a neutral accent theme', () => {
    localStorage.clear()
    localStorage.setItem(COLOR_THEME_STORAGE_KEY, 'dark')
    localStorage.setItem('filvault.theme', 'dark')
    syncAppearanceFromStorage()

    expect(migrateAppearanceMode()).toBe('dark')
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('default')
    expect(document.documentElement.dataset.colorTheme).toBe('default')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('keeps explicit light appearance when stale legacy colorTheme is dark', () => {
    localStorage.clear()
    localStorage.setItem(APPEARANCE_MODE_STORAGE_KEY, 'light')
    localStorage.setItem(COLOR_THEME_STORAGE_KEY, 'dark')
    syncAppearanceFromStorage()

    const { appearanceMode, resolvedIsDark } = useTheme()
    expect(migrateAppearanceMode()).toBe('light')
    expect(appearanceMode.value).toBe('light')
    expect(resolvedIsDark.value).toBe(false)
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('default')
    expect(document.documentElement.dataset.colorTheme).toBe('default')
    expect(document.documentElement.dataset.theme).toBeUndefined()
  })

  it('keeps explicit system appearance when stale legacy colorTheme is dark', () => {
    localStorage.clear()
    localStorage.setItem(APPEARANCE_MODE_STORAGE_KEY, 'system')
    localStorage.setItem(COLOR_THEME_STORAGE_KEY, 'dark')
    mockMatchMedia(false)
    syncAppearanceFromStorage()

    const { appearanceMode, resolvedIsDark } = useTheme()
    expect(migrateAppearanceMode()).toBe('system')
    expect(appearanceMode.value).toBe('system')
    expect(resolvedIsDark.value).toBe(false)
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('default')
    expect(document.documentElement.dataset.theme).toBeUndefined()
  })

  it('persists explicit light appearance across hydrate and sync', () => {
    const { setAppearanceMode, appearanceMode } = useTheme()
    setAppearanceMode('light')

    hydrateAppearance()
    expect(appearanceMode.value).toBe('light')
    expect(document.documentElement.dataset.theme).toBeUndefined()

    localStorage.setItem(COLOR_THEME_STORAGE_KEY, 'dark')
    syncAppearanceFromStorage()

    expect(appearanceMode.value).toBe('light')
    expect(localStorage.getItem(APPEARANCE_MODE_STORAGE_KEY)).toBe('light')
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('default')
    expect(document.documentElement.dataset.theme).toBeUndefined()
  })

  it('migrates legacy followSystemDark flag to system appearance mode', () => {
    localStorage.clear()
    localStorage.setItem('filvault.followSystemDark', 'true')
    localStorage.setItem('filvault.theme', 'light')
    expect(migrateAppearanceMode()).toBe('system')
  })

  it('hydrates appearance before Vue mount', () => {
    localStorage.setItem(APPEARANCE_MODE_STORAGE_KEY, 'dark')
    localStorage.setItem(COLOR_THEME_STORAGE_KEY, 'teal')
    delete document.documentElement.dataset.theme
    delete document.documentElement.dataset.colorTheme

    hydrateAppearance()

    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(document.documentElement.dataset.colorTheme).toBe('teal')
  })

  it('keeps Settings and Theme Center on the same appearance state API', () => {
    const settings = useTheme()
    const themeCenter = useTheme()

    settings.setAppearanceMode('dark')
    settings.applyColorTheme('sunshine')

    expect(themeCenter.appearanceMode.value).toBe('dark')
    expect(themeCenter.currentColorTheme.value).toBe('sunshine')
    expect(themeCenter.resolvedIsDark.value).toBe(true)
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
    expect(localStorage.getItem(COLOR_THEME_STORAGE_KEY)).toBe('spring')
    expect(document.documentElement.dataset.colorTheme).toBe('spring')
  })

  it('resolves appearance darkness for each mode', () => {
    mockMatchMedia(true)

    expect(resolveAppearanceIsDark('dark')).toBe(true)
    expect(resolveAppearanceIsDark('light')).toBe(false)
    expect(resolveAppearanceIsDark('system')).toBe(true)
  })
})

function setAppearanceModeLight() {
  useTheme().setAppearanceMode('light')
}
