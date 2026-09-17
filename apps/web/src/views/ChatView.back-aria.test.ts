import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const chat = readFileSync(fileURLToPath(new URL('./ChatView.vue', import.meta.url)), 'utf-8')

describe('Chat back accessibility contract', () => {
  it('uses the reactive localized back label on the real control', () => {
    const backButton = chat.match(/<button[^>]*class="icon-btn back-btn"[^>]*>/)?.[0] ?? ''

    expect(backButton).toContain(':aria-label="t.back"')
    expect(backButton).toContain('@click="backToRail"')
  })
})
