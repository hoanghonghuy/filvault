/**
 * @vitest-environment jsdom
 */
import { describe, it, expect } from 'vitest'
import {
  rgbToHsl,
  hslToRgb,
  rgbToHex,
  PRESET_THEMES,
  resolveWallpaperTheme,
} from './wallpaperPalette'

describe('wallpaperPalette', () => {
  it('converts rgb to hsl and back correctly', () => {
    // Pure red
    const [h1, s1, l1] = rgbToHsl(255, 0, 0)
    expect(h1).toBe(0)
    expect(s1).toBe(1)
    expect(l1).toBe(0.5)

    const [r1, g1, b1] = hslToRgb(h1, s1, l1)
    expect(r1).toBe(255)
    expect(g1).toBe(0)
    expect(b1).toBe(0)

    // Pure green
    const [h2, s2, l2] = rgbToHsl(0, 255, 0)
    expect(h2).toBe(120)
    expect(s2).toBe(1)
    expect(l2).toBe(0.5)

    // Cyan
    const [h3, s3, l3] = rgbToHsl(0, 255, 255)
    expect(h3).toBe(180)
  })

  it('converts rgb to hex correctly', () => {
    expect(rgbToHex(255, 0, 0)).toBe('#ff0000')
    expect(rgbToHex(0, 132, 255)).toBe('#0084ff')
    expect(rgbToHex(16, 185, 129)).toBe('#10b981')
  })

  it('resolves preset themes synchronously and returns null for empty or none', async () => {
    expect(await resolveWallpaperTheme(null)).toBeNull()
    expect(await resolveWallpaperTheme('')).toBeNull()
    expect(await resolveWallpaperTheme('none')).toBeNull()

    const aurora = await resolveWallpaperTheme('aurora')
    expect(aurora).toBeDefined()
    expect(aurora?.primary).toBe(PRESET_THEMES.aurora.primary)
    expect(aurora?.isDark).toBe(true)

    const doodle = await resolveWallpaperTheme('doodle')
    expect(doodle).toBeDefined()
    expect(doodle?.primary).toBe(PRESET_THEMES.doodle.primary)
    expect(doodle?.isDark).toBe(false)
  })

  it('returns fallback theme for invalid or broken image', async () => {
    const theme = await resolveWallpaperTheme('data:image/png;base64,invalid')
    expect(theme).toBeDefined()
    expect(theme?.primary).toBeTruthy()
    expect(theme?.gradient).toContain('linear-gradient')
  })
})
