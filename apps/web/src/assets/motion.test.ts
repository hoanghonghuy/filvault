import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('motion token contract (DESIGN.md §1)', () => {
  const mainCss = readSrc('./main.css')

  it('defines Material 3 easing tokens', () => {
    expect(mainCss).toMatch(/--ease-standard:\s*cubic-bezier\(0\.2,\s*0,\s*0,\s*1\)/)
    expect(mainCss).toMatch(
      /--ease-emphasized-decelerate:\s*cubic-bezier\(0\.05,\s*0\.7,\s*0\.1,\s*1\)/,
    )
    expect(mainCss).toMatch(
      /--ease-emphasized-accelerate:\s*cubic-bezier\(0\.3,\s*0,\s*0\.8,\s*0\.15\)/,
    )
  })

  it('defines duration tokens', () => {
    expect(mainCss).toMatch(/--duration-short:\s*100ms/)
    expect(mainCss).toMatch(/--duration-medium:\s*200ms/)
    expect(mainCss).toMatch(/--duration-long:\s*300ms/)
  })

  it('disables motion under prefers-reduced-motion', () => {
    expect(mainCss).toMatch(
      /@media\s*\(\s*prefers-reduced-motion:\s*reduce\s*\)[\s\S]*?transition-duration:\s*0\.01ms/,
    )
  })
})

describe('press feedback contract', () => {
  const mainCss = readSrc('./main.css')

  it('scales buttons and tappable rows on :active', () => {
    expect(mainCss).toMatch(/\.btn:not\(:disabled\):active[\s\S]*?transform:\s*scale\(0\.97\)/)
    expect(mainCss).toMatch(/\.row\.tappable:active[\s\S]*?transform:\s*scale\(0\.98\)/)
  })
})

describe('bottom sheet enter/exit animation contract', () => {
  const sheet = readSrc('../components/BottomSheet.vue')
  const mainCss = readSrc('./main.css')

  it('wraps the panel tree in a Transition named "sheet"', () => {
    expect(sheet).toContain('<Transition name="sheet"')
    expect(sheet).toContain('.sheet-enter-active')
    expect(sheet).toContain('.sheet-leave-active')
  })

  it('enters slower than it leaves', () => {
    const tokenMs = (name: string): number => {
      const match = mainCss.match(new RegExp(`--${name}:\\s*(\\d+)ms`))
      return match ? Number(match[1]) : NaN
    }
    const durationOf = (block: string): number => {
      const varRef = block.match(/var\(--([a-z-]+)\)/)
      if (varRef) return tokenMs(varRef[1])
      const msValue = block.match(/(\d+)ms/)
      return msValue ? Number(msValue[1]) : NaN
    }
    const enter = sheet.match(/\.sheet-enter-active\s*\{[^}]*\}/)?.[0] ?? ''
    const leave = sheet.match(/\.sheet-leave-active\s*\{[^}]*\}/)?.[0] ?? ''
    expect(durationOf(enter)).toBeGreaterThan(durationOf(leave))
  })
})

describe('toast enter/exit animation contract', () => {
  const toastHost = readSrc('../components/ToastHost.vue')

  it('wraps the toast in a Transition named "toast"', () => {
    expect(toastHost).toContain('<Transition name="toast"')
    expect(toastHost).toContain('.toast-enter-active')
    expect(toastHost).toContain('.toast-leave-active')
  })
})

describe('route transition contract', () => {
  const shell = readSrc('../components/AppShell.vue')

  it('renders RouterView inside a Transition with fade-slide classes', () => {
    expect(shell).toContain('<RouterView v-slot="{ Component }"')
    expect(shell).toMatch(/<Transition\s+name="page"/)
    expect(shell).toContain('.page-enter-active')
    expect(shell).toContain('.page-leave-active')
  })
})

describe('list move transition contract (TransitionGroup)', () => {
  const mainCss = readSrc('./main.css')
  const transitionGroupPattern = (extra = '') =>
    new RegExp(`<TransitionGroup\\s[^>]*name="row"[^>]*tag="section"${extra}`, 'g')

  it('uses TransitionGroup for Files browser rows', () => {
    const files = readSrc('../views/FilesView.vue')
    // search results + browser rows + favorites list
    expect(files.match(transitionGroupPattern())?.length).toBe(3)
  })

  it('uses TransitionGroup for Trash sections', () => {
    const trash = readSrc('../views/TrashView.vue')
    expect(trash.match(transitionGroupPattern())?.length).toBe(2)
  })

  it('defines shared row enter/leave/move classes in main.css', () => {
    expect(mainCss).toContain('.row-enter-from')
    expect(mainCss).toContain('.row-leave-active')
    expect(mainCss).toContain('.row-move')
  })
})
