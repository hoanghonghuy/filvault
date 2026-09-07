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
    expect(nav).toMatch(/to:\s*'\/chat'/)
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

  it('supports enter-to-send, day separators, timestamps and media thumbnails', () => {
    expect(chat).toMatch(/@keydown.enter/)
    expect(chat).toMatch(/shiftKey/)
    expect(chat).toMatch(/class="day-separator"/)
    expect(chat).toMatch(/function shouldShowDate|shouldShowDate\(/)
    expect(chat).toMatch(/formatTime\(/)
    expect(chat).toMatch(/item\.thumbnailUrl/)
  })

  it('renders inline images with a lightbox and polished send/skeleton states', () => {
    expect(chat).toMatch(/MediaLightbox/)
    expect(chat).toMatch(/function openInlineImage|openInlineImage\(/)
    expect(chat).toMatch(/class="inline-image"/)
    expect(chat).toMatch(/:class="\{\s*active:.*draft/)
    expect(chat).toMatch(/LoadingSkeletonThread/)
    expect(chat).toMatch(/composerRef\.value\?\.(focus|value)/)
  })

  it('gives the conversation rail Messenger-like filtering and identity', () => {
    expect(chat).toMatch(/railFilter|filterQuery/)
    expect(chat).toMatch(/filteredConversations/)
    expect(chat).toMatch(/avatarClass|avatar-color/)
    expect(chat).toMatch(/function formatRelativeDay|formatRelativeDay\(/)
    expect(chat).toMatch(/rail-empty/)
  })

  it('keeps mobile in the inbox until a conversation is chosen', () => {
    expect(chat).toMatch(/function isMobileViewport/)
    expect(chat).toMatch(/!isMobileViewport\(\)/)
    expect(chat).toMatch(/class="chat-app" :class="\{ 'in-thread': inThread \}"/)
    expect(chat).toMatch(/name="arrow-left"/)
  })

  it('uses responsive Messenger-style details and progressive actions', () => {
    expect(chat).toMatch(/threadSearchOpen/)
    expect(chat).toMatch(/promptNewConversation/)
    expect(chat).toMatch(/<UploadProgress/)
    expect(chat).toMatch(/@media \(min-width: 1200px\)/)
    expect(chat).toMatch(/@media \(min-width: 768px\) and \(max-width: 1199px\)/)
  })

  it('defines typed chat API payloads', () => {
    expect(types).toMatch(/export interface ChatConversation/)
    expect(types).toMatch(/export interface ChatMessage/)
    expect(types).toMatch(/export interface ChatAttachment/)
  })

  it('supports reliable message mutations and failed-send recovery', () => {
    expect(chat).toMatch(/function editMessage|editMessage\(/)
    expect(chat).toMatch(/function removeMessage|removeMessage\(/)
    expect(chat).toMatch(/Retry/)
    expect(chat).toMatch(/Discard/)
    expect(chat).toMatch(/clientMessageId/)
  })

  it('captures attachment work against its original conversation', () => {
    expect(chat).toMatch(/attachmentQueue/)
    expect(chat).toMatch(/conversationId/)
    expect(chat).toMatch(/Cancel upload|cancelUpload|AbortController/)
    expect(chat).toMatch(/retryUpload|Retry upload/)
  })

  it('reconnects realtime events with a durable cursor and deduplicates messages', () => {
    const store = readSrc('../stores/chat.ts')
    expect(store).toMatch(/lastEventId/)
    expect(store).toMatch(/AbortController/)
    expect(store).toMatch(/reconcileMessage/)
    expect(store).toMatch(/message\.created|message\.edited|message\.removed/)
  })

  it('renders a friendly Messenger-style empty thread with a quick wave button', () => {
    expect(chat).toMatch(/class="thread-empty-state"/)
    expect(chat).toMatch(/class="thread-empty-avatar"/)
    expect(chat).toMatch(/thread-empty-wave-btn/)
    expect(chat).toMatch(/function sendQuickWave|sendQuickWave\(\)/)
  })

  it('hides inline action buttons and triggers message options via long-press', () => {
    expect(chat).not.toMatch(/class="message-bubble"[^>]*>[\s\S]*?<div v-if="canMutateMessage\(message\)" class="message-actions"/)
    expect(chat).toMatch(/openMessageMenu/)
    expect(chat).toMatch(/messageMenuOpen/)
    expect(chat).toMatch(/copyMessageText/)
    expect(chat).toMatch(/triggerEditMessage/)
    expect(chat).toMatch(/triggerRemoveMessage/)
  })

  it('triggers conversation row options via long-press', () => {
    expect(chat).toMatch(/openConvMenu/)
    expect(chat).toMatch(/convMenuOpen/)
    expect(chat).toMatch(/toggleMuteConversation/)
    expect(chat).toMatch(/deleteConversation/)
  })

  it('provides a Messenger-grade chat info sheet with theme picker and nicknames', () => {
    expect(chat).toMatch(/threadInfoOpen/)
    expect(chat).toMatch(/class="chat-info-content"/)
    expect(chat).toMatch(/chatThemes/)
    expect(chat).toMatch(/threadThemeStyle/)
    expect(chat).toMatch(/changeNickname/)
    expect(chat).toMatch(/function changeNickname[\s\S]*?threadInfoOpen\.value = false/)
  })

  it('highlights unread conversations with bold title, preview, unread date and badge', () => {
    expect(chat).toMatch(/isConversationUnread/)
    expect(chat).toMatch(/class="unread-badge"/)
    expect(chat).toMatch(/unread-date/)
    expect(chat).toMatch(/\.conversation-row\.unread/)
    expect(types).toMatch(/unreadCount\?: number/)
  })

  it('supports peek preview without triggering read status', () => {
    expect(chat).toMatch(/peekOpen/)
    expect(chat).toMatch(/openPeekPreview/)
    expect(chat).toMatch(/openChatFromPeek/)
    expect(chat).toMatch(/peek-preview-content/)
    expect(chat).toMatch(/peek-messages-list/)
  })

  it('provides real-time typing indicators in both rail and thread for peers only', () => {
    const store = readSrc('../stores/chat.ts')
    expect(store).toMatch(/typingUsers/)
    expect(store).toMatch(/sendTyping/)
    expect(store).toMatch(/typing\.indicator/)
    expect(store).toMatch(/userId === auth\.user\?\.id/)
    expect(chat).toMatch(/conversation-typing/)
    expect(chat).toMatch(/typing-bubble/)
    expect(chat).toMatch(/typing-dot/)
    expect(chat).toMatch(/onComposerInput/)
    expect(chat).toMatch(/isPeerTyping/)
  })

  it('displays seen receipt indicator for read messages', () => {
    const store = readSrc('../stores/chat.ts')
    expect(store).toMatch(/markAsRead/)
    expect(store).toMatch(/message\.read/)
    expect(chat).toMatch(/seen-indicator/)
    expect(chat).toMatch(/isMessageSeenByPeer/)
    expect(types).toMatch(/peerLastReadMessageId\?: string/)
  })

  it('tracks real-time peer presence with dynamic status and indicator', () => {
    const store = readSrc('../stores/chat.ts')
    const authStore = readSrc('../stores/auth.ts')
    expect(store).toMatch(/presence\.changed/)
    expect(types).toMatch(/peerStatus\?: 'online' \| 'offline'/)
    expect(types).toMatch(/peerLastSeenAt\?: string/)
    expect(types).toMatch(/activeStatusEnabled\?: boolean/)
    expect(authStore).toMatch(/updateActiveStatus/)
    expect(chat).toMatch(/isPeerOnline/)
    expect(chat).toMatch(/formatLastSeen/)
    expect(chat).toMatch(/auth\.user\?\.activeStatusEnabled === false/)
    expect(chat).toMatch(/activeDaysAgo/)
    expect(chat).toMatch(/activeOnDate/)
    expect(chat).toMatch(/toggleActiveStatus/)
    expect(chat).toMatch(/v-if="isPeerOnline\(conv\)"\s+class="online-indicator"/)
    expect(chat).toMatch(/thread-status/)
    expect(chat).toMatch(/class="chat-info-status"/)
  })
})

