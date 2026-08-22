import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('sliding nav indicator contract', () => {
  const shell = readSrc('../components/AppShell.vue')

  it('renders an indicator element inside bottom nav links', () => {
    expect(shell).toMatch(/class="bottom-icon"[\s\S]*?nav-indicator/)
    expect(shell).toContain('.nav-indicator')
  })

  it('animates the indicator with transform only', () => {
    const indicator = shell.match(/\.nav-indicator\s*\{[^}]*\}/)?.[0] ?? ''
    expect(indicator).toContain('transform')
    expect(indicator).not.toMatch(/transition:\s*all/)
  })

  it('scales the indicator for active links', () => {
    expect(shell).toMatch(
      /\.bottom-link\.router-link-active\s+\.nav-indicator[\s\S]*?transform:[^;}]*scaleX\(1\)/,
    )
  })
})

describe('FAB press feedback contract', () => {
  const fab = readSrc('../components/UploadFab.vue')

  it('shrinks and lifts on press', () => {
    expect(fab).toMatch(/\.fab:not\(:disabled\):active\s*\{[^}]*transform:\s*scale\(0\.94\)/)
    expect(fab).toMatch(/\.fab:not\(:disabled\):active\s*\{[^}]*box-shadow/)
  })
})

describe('optimistic delete contract (FilesView)', () => {
  const files = readSrc('../views/FilesView.vue')

  it('removes rows before calling the API', () => {
    expect(files).toMatch(/function removeOptimistic\(/)
    const fn = files.match(/function deleteFile[\s\S]*?\n\}/)?.[0] ?? ''
    const apiCallIndex = fn.indexOf("api(`/files/${id}`")
    const optimisticIndex = fn.indexOf('removeOptimistic(')
    expect(optimisticIndex).toBeGreaterThan(-1)
    expect(apiCallIndex).toBeGreaterThan(optimisticIndex)
  })

  it('restores rows when the API fails', () => {
    const fn = files.match(/async function deleteFile[\s\S]*?\n\}/)?.[0] ?? ''
    expect(fn).toContain('await loadBrowser()')
  })
})

describe('optimistic trash actions contract (TrashView)', () => {
  const trash = readSrc('../views/TrashView.vue')

  it('updates local lists before calling restore/delete APIs', () => {
    expect(trash).toMatch(/function removeFromLocal\(/)
    for (const name of ['restoreFile', 'restoreFolder', 'permanentDelete']) {
      const fn = trash.match(new RegExp(`async function ${name}[\\s\\S]*?\\n\\}`))?.[0] ?? ''
      const apiCall = fn.match(/api\(/)?.index ?? -1
      const optimistic = fn.indexOf('removeFromLocal(')
      expect(optimistic).toBeGreaterThan(-1)
      expect(optimistic).toBeLessThan(apiCall)
    }
  })
})
