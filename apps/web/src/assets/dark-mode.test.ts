/// <reference types="node" />

import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('dark mode contract', () => {
  const css = readSrc('./main.css')
  const settings = readSrc('../views/SettingsView.vue')
  const themeView = readSrc('../views/ThemeView.vue')
  const chatView = readSrc('../views/ChatView.vue')
  const main = readSrc('../main.ts')
  const themeModule = readSrc('../lib/theme.ts')

  it('defines a dark theme token override', () => {
    expect(css).toMatch(/\[data-theme='dark'\]/)
    for (const token of ['--ink', '--body', '--canvas', '--surface-soft', '--surface-card', '--hairline']) {
      expect(css).toMatch(new RegExp(`\\[data-theme='dark'\\][\\s\\S]*${token}:`))
    }
  })

  it('keeps semantic status colors in dark mode', () => {
    expect(css).toMatch(/\[data-theme='dark'\][\s\S]*--danger:/)
    expect(css).toMatch(/\[data-theme='dark'\][\s\S]*--success:/)
    expect(css).toMatch(/\[data-theme='dark'\][\s\S]*--warning:/)
  })

  it('hydrates appearance before mounting Vue', () => {
    expect(main).toMatch(/hydrateAppearance\(\)/)
    expect(themeModule).toMatch(/export function hydrateAppearance/)
  })

  it('routes Settings through the canonical appearance API', () => {
    expect(settings).toMatch(/useTheme\(\)/)
    expect(settings).toMatch(/setAppearanceMode/)
    expect(settings).not.toMatch(/localStorage\.setItem\('filvault\.theme'/)
    expect(settings).not.toMatch(/document\.documentElement\.dataset\.theme/)
  })

  it('routes Theme Center through the canonical appearance API', () => {
    expect(themeView).toMatch(/useTheme\(\)/)
    expect(themeView).toMatch(/setAppearanceMode/)
    expect(themeView).not.toMatch(/setFollowSystemDark/)
    expect(themeView).not.toMatch(/document\.documentElement\.dataset\.theme/)
  })

  it('keeps Chat appearance read-only and routes global changes to canonical Settings', () => {
    expect(chatView).toMatch(/useTheme\(\)/)
    expect(chatView).toMatch(/resolvedIsDark/)
    expect(chatView).toMatch(/navigateTo\('\/settings#appearance'\)/)
    expect(chatView).not.toMatch(/toggleResolvedAppearance/)
    expect(chatView).not.toMatch(/localStorage\.setItem\('filvault\.theme'/)
    expect(chatView).not.toMatch(/localStorage\.removeItem\('filvault\.theme'/)
    expect(chatView).not.toMatch(/document\.documentElement\.dataset\.theme/)
  })

  it('uses subtle, pale sidebar backgrounds for light themes so colors are not overly saturated', () => {
    expect(css).toMatch(/\[data-color-theme='peach'\][\s\S]*?--sidebar-bg:\s*#fbe6ee;/)
    expect(css).toMatch(/\[data-color-theme='peach'\][\s\S]*?--surface-soft:\s*#fdf7f9;/)
    expect(css).toMatch(/\[data-color-theme='spring'\][\s\S]*?--sidebar-bg:\s*#fce7f1;/)
    expect(css).toMatch(/\[data-color-theme='spring'\][\s\S]*?--surface-soft:\s*#fdf8fb;/)
  })

  it('ensures light theme sidebar-bg is visually distinct from content surface-soft', () => {
    const peachMatch = css.match(/\[data-color-theme='peach'\]\{([^}]+)\}/s) || css.match(/\[data-color-theme='peach'\]\s*\{([\s\S]*?)\n\}/)
    expect(peachMatch).not.toBeNull()
    const block = peachMatch ? peachMatch[1] : ''
    const sidebar = block.match(/--sidebar-bg:\s*([^;]+);/)?.[1]?.trim()
    const surface = block.match(/--surface-soft:\s*([^;]+);/)?.[1]?.trim()
    expect(sidebar).toBeDefined()
    expect(surface).toBeDefined()
    expect(sidebar).not.toBe(surface)
  })
})
