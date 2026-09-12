import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const chatView = readFileSync(
  fileURLToPath(new URL('../views/ChatView.vue', import.meta.url)),
  'utf-8',
)
const scopedCss = chatView.match(/<style scoped>([\s\S]*?)<\/style>/)?.[1] ?? ''
const allCss = [...chatView.matchAll(/<style(?: scoped)?>([\s\S]*?)<\/style>/g)]
  .map((match) => match[1])
  .join('\n')

describe('Chat semantic accent contract', () => {
  it('defines the Chat accent and outgoing bubble tokens at the Chat root', () => {
    const root = scopedCss.match(/\.chat-app\s*\{([\s\S]*?)\}/)?.[1] ?? ''
    expect(root).toContain('--chat-accent: var(--accent)')
    expect(root).toContain('--chat-accent-secondary: var(--accent)')
    expect(root).toContain('--chat-bubble-outgoing: linear-gradient(')
  })

  it('does not hardcode Messenger blue inside Chat scoped CSS', () => {
    expect(allCss.toLowerCase()).not.toContain('#0084ff')
    expect(allCss).not.toContain('rgba(0, 132, 255')
  })

  it('keeps the named Messenger-blue palette entry as theme data', () => {
    expect(chatView).toMatch(/id: 'blue'[\s\S]*color: '#0084ff'/)
  })
})
