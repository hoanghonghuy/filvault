import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('chat surface contract', () => {
  const router = readSrc('../router/index.ts')
  const nav = readSrc('../lib/shellNav.ts')
  const chat = readSrc('./ChatView.vue')
  const types = readSrc('../api/types.ts')

  it('adds Chat as a dedicated authenticated route and shell nav item', () => {
    expect(router).toMatch(/path:\s*'\/chat'/)
    expect(router).toMatch(/name:\s*'chat'/)
    expect(nav).toMatch(/to:\s*'\/chat'/)
    expect(nav).toMatch(/icon:\s*'chat'/)
  })

  it('keeps a Messenger-like three-area layout', () => {
    expect(chat).toMatch(/class="chat-layout"/)
    expect(chat).toMatch(/class="conversation-list"/)
    expect(chat).toMatch(/class="message-thread"/)
    expect(chat).toMatch(/class="chat-composer"/)
  })

  it('supports text messages and media attachment uploads', () => {
    expect(chat).toMatch(/\/chat\/conversations/)
    expect(chat).toMatch(/\/chat\/attachments\/upload-sessions/)
    expect(chat).toMatch(/\/chat\/attachments\/\$\{[^}]+\.fileId\}\/complete/)
    expect(chat).toMatch(/type="file"/)
  })

  it('defines typed chat API payloads', () => {
    expect(types).toMatch(/export interface ChatConversation/)
    expect(types).toMatch(/export interface ChatMessage/)
    expect(types).toMatch(/export interface ChatAttachment/)
  })

  it('adds message search and a conversation media panel', () => {
    expect(chat).toMatch(/class="chat-search"/)
    expect(chat).toMatch(/\/chat\/conversations\/\$\{[^}]+\}\/messages\/search/)
    expect(chat).toMatch(/class="chat-media-panel"/)
    expect(chat).toMatch(/\/chat\/conversations\/\$\{[^}]+\}\/media/)
  })
})
