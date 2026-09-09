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

  it('routes Chat quick toggle through the canonical appearance API', () => {
    expect(chatView).toMatch(/useTheme\(\)/)
    expect(chatView).toMatch(/toggleResolvedAppearance/)
    expect(chatView).not.toMatch(/localStorage\.setItem\('filvault\.theme'/)
    expect(chatView).not.toMatch(/localStorage\.removeItem\('filvault\.theme'/)
    expect(chatView).not.toMatch(/document\.documentElement\.dataset\.theme/)
  })
})
