import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('chat surface contract', () => {
  const router = readSrc('../router/index.ts')
  const nav = readSrc('../lib/shellNav.ts')
  const shell = readSrc('../components/AppShell.vue')
  const chat = readSrc('./ChatView.vue')
  const types = readSrc('../api/types.ts')

  it('serves Chat at its own URL outside the Filvault shell', () => {
    expect(router).toMatch(/path:\s*'\/chat'/)
    expect(router).toMatch(/name:\s*'chat'/)
    expect(router).toMatch(/meta:\s*\{[^}]*bare:\s*true/)
    expect(nav).not.toMatch(/to:\s*'\/chat'/)
    expect(shell).toMatch(/meta\.bare|!isBare/)
  })

  it('keeps a Messenger-like full-screen anatomy', () => {
    expect(chat).toMatch(/class="chat-app/)
    expect(chat).toMatch(/class="chat-rail"/)
    expect(chat).toMatch(/class="message-thread"/)
    expect(chat).toMatch(/class="chat-composer"/)
  })

  it('switches rail/thread on mobile with an accessible back action', () => {
    expect(chat).toMatch(/'in-thread':/)
    expect(chat).toMatch(/aria-label="Back"/)
    expect(chat).toMatch(/function backToRail/)
  })

  it('supports search, media panel and attachment uploads', () => {
    expect(chat).toMatch(/\/chat\/conversations/)
    expect(chat).toMatch(/\/chat\/conversations\/\$\{[^}]+\}\/messages\/search/)
    expect(chat).toMatch(/\/chat\/conversations\/\$\{[^}]+\}\/media/)
    expect(chat).toMatch(/type="file"/)
  })

  it('animates new messages and keeps the latest window visible', () => {
    expect(chat).toMatch(/<TransitionGroup[^>]*name="msg"/)
    expect(chat).toMatch(/\.msg-enter-active/)
    expect(chat).toMatch(/function scrollToLatest|scrollToLatest\(\)/)
    expect(chat).toMatch(/class="jump-latest"/)
    expect(chat).toMatch(/showJump/)
  })

  it('sends optimistically with a pending bubble and composer auto-grow', () => {
    expect(chat).toMatch(/pendingMessage|isPending/)
    expect(chat).toMatch(/function autoGrow/)
    expect(chat).toMatch(/\?includePreview=true/)
    expect(chat).toMatch(/nextBefore|hasMore/)
  })

  it('defines typed chat API payloads', () => {
    expect(types).toMatch(/export interface ChatConversation/)
    expect(types).toMatch(/export interface ChatMessage/)
    expect(types).toMatch(/export interface ChatAttachment/)
  })
})
