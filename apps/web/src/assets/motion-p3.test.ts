import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('router scroll behavior contract', () => {
  const router = readSrc('../router/index.ts')

  it('returns to top on navigation and restores on back', () => {
    expect(router).toMatch(/scrollBehavior\s*\(/)
    expect(router).toMatch(/savedPosition/)
    expect(router).toMatch(/\{\s*left:\s*0,\s*top:\s*0\s*\}/)
  })
})

describe('logout confirmation contract', () => {
  const settings = readSrc('../views/SettingsView.vue')

  it('asks for confirmation before logging out', () => {
    const fn = settings.match(/async function logout[\s\S]*?\n\}/)?.[0] ?? ''
    expect(fn).toContain('ui.confirm(')
    expect(fn.indexOf('ui.confirm(')).toBeLessThan(fn.indexOf('auth.logout()'))
  })
})

describe('hardcoded timing cleanup contract', () => {
  const violations: string[] = []

  it('uses motion tokens instead of raw ms/s durations in component styles', () => {
    for (const file of ['../views/SettingsView.vue', '../components/StorageBar.vue']) {
      const source = readSrc(file)
      const matches = source.match(/(?:transition|animation):[^;]*?(?:\d+m?s)/g) ?? []
      for (const m of matches) {
        if (!m.includes('var(--duration') && !m.includes('var(--ease')) {
          violations.push(`${file}: ${m.trim().slice(0, 80)}`)
        }
      }
    }
    expect(violations).toEqual([])
  })

  it('animates the storage fill with transform (scaleX), not layout width', () => {
    const storageBar = readSrc('../components/StorageBar.vue')
    expect(storageBar).toMatch(/\.fill\s*\{[^}]*transform:\s*scaleX/)
    expect(storageBar).not.toMatch(/transition:\s*width/)
  })
})

describe('bottom sheet focus trap contract', () => {
  const sheet = readSrc('../components/BottomSheet.vue')

  it('keeps Tab cycling inside the panel while open', () => {
    const script = sheet.match(/<script setup lang="ts">([\s\S]*?)<\/script>/)?.[1] ?? ''
    expect(script).toMatch(/focusablesInPanel|querySelectorAll/)
    expect(script).toMatch(/Tab/)
  })
})

describe('upload progress contract', () => {
  it('renders a ProgressBar component instead of a text-only status line', () => {
    const files = readSrc('../views/FilesView.vue')
    expect(files).toContain('<UploadProgress')
    expect(files).toContain('useFileUploadQueue')
    expect(files).not.toMatch(/Uploading…\s*\{\{/)
  })

  it('creates a reusable UploadProgress component wired to tokens', () => {
    const bar = readSrc('../components/UploadProgress.vue')
    expect(bar).toContain('--ease-standard')
    expect(bar).toMatch(/aria-live="polite"/)
    expect(bar).toMatch(/role="progressbar"/)
    expect(bar).toMatch(/aggregateProgress/)
  })
})
