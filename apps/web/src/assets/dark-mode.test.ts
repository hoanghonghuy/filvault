/// <reference types="node" />

import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('dark mode contract', () => {
  const css = readSrc('./main.css')
  const settings = readSrc('../views/SettingsView.vue')
  const main = readSrc('../main.ts')

  it('defines a dark theme token override', () => {
    expect(css).toMatch(/\[data-theme='dark'\]/)
    for (const token of ['--ink', '--body', '--canvas', '--surface-soft', '--surface-card', '--hairline']) {
      expect(css).toMatch(new RegExp(`\\[data-theme='dark'\\][\\s\\S]*${token}:`))
    }
  })

  it('applies a saved theme before mounting Vue', () => {
    expect(main).toMatch(/localStorage\.getItem\('filvault.theme'\)/)
    expect(main).toMatch(/document\.documentElement\.dataset\.theme/)
  })

  it('exposes a theme toggle in Settings wired through useTheme', () => {
    expect(settings).toMatch(/t\.appearance/)
    expect(settings).toMatch(/t\.darkMode/)
    expect(settings).toMatch(/useTheme\(\)/)
    expect(settings).toMatch(/setDarkMode/)
    expect(settings).toMatch(/isDarkMode/)
  })
})
