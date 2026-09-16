import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

const chat = readSrc('./ChatView.vue')
const i18n = readSrc('../lib/i18n.ts')

describe('ChatView mobile back accessibility contract', () => {
  it('binds the real mobile back button to locale-reactive copy and navigation', () => {
    const backButton = chat.match(/<button\s+type="button"\s+class="icon-btn back-btn"[\s\S]*?>/)?.[0] ?? ''

    expect(backButton).not.toBe('')
    expect(backButton).toContain(':aria-label="t.back"')
    expect(backButton).toContain('@click="backToRail"')
  })

  it('defines the back accessible name in both supported locale tables', () => {
    expect(i18n.match(/\bback:/g)).toHaveLength(2)
  })
})
