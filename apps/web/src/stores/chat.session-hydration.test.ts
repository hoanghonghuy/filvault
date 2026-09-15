/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { ChatConversation, ChatMessage, User } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'

const apiMock = vi.hoisted(() => vi.fn<(path: string, options?: unknown) => Promise<unknown>>())

vi.mock('@/api/client', () => ({
  API_BASE: 'http://test.local',
  api: apiMock,
  getAccessToken: vi.fn<() => string | null>(() => 'access-token'),
  refreshAccessToken: vi.fn<() => Promise<boolean>>(async () => false),
}))

vi.mock('@/stores/call', () => ({
  useCallStore: () => ({
    handleSignal: vi.fn<(signal: unknown) => void>(),
  }),
}))

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function makeUser(id: string, email: string): User {
  return {
    id,
    email,
    displayName: email.split('@')[0] ?? email,
    emailVerified: true,
    storageUsed: 0,
    storageQuota: 1_073_741_824,
    imageThumbnailsEnabled: true,
    videoThumbnailsEnabled: true,
    trashAutoDeleteEnabled: false,
    trashRetentionDays: 30,
    createdAt: '2026-01-01T00:00:00.000Z',
  }
}

function makeConversation(id: string, title: string): ChatConversation {
  return {
    id,
    title,
    createdAt: '2026-01-01T00:00:00.000Z',
    updatedAt: '2026-01-01T00:00:00.000Z',
  }
}

function makeMessage(id: string, body: string): ChatMessage {
  return {
    id,
    body,
    createdAt: '2026-01-01T00:00:00.000Z',
    senderId: 'sender',
    senderName: 'Sender',
  }
}

describe('useChatStore session hydration isolation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    sessionStorage.clear()
    apiMock.mockReset()
  })

  it('does not commit stale conversation hydration after logout', async () => {
    const pending = deferred<{ conversations: ChatConversation[] }>()
    apiMock.mockImplementation((path) => {
      if (path === '/chat/conversations?includePreview=true') return pending.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const auth = useAuthStore()
    auth.user = makeUser('user-a', 'a@example.com')
    const chat = useChatStore()
    chat.conversations = [makeConversation('existing', 'Existing')]

    const hydration = chat.loadConversations()

    auth.user = null
    pending.resolve({ conversations: [makeConversation('stale', 'Stale account A')] })
    await hydration

    expect(chat.conversations).toEqual([])
  })

  it('does not commit stale conversation hydration into a different account', async () => {
    const pending = deferred<{ conversations: ChatConversation[] }>()
    apiMock.mockImplementation((path) => {
      if (path === '/chat/conversations?includePreview=true') return pending.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const auth = useAuthStore()
    auth.user = makeUser('user-a', 'a@example.com')
    const chat = useChatStore()

    const hydration = chat.loadConversations()
    auth.user = makeUser('user-b', 'b@example.com')

    pending.resolve({ conversations: [makeConversation('stale', 'Account A list')] })
    await hydration

    expect(chat.conversations).toEqual([])
  })

  it('keeps current-account conversation hydration authoritative', async () => {
    apiMock.mockResolvedValue({
      conversations: [makeConversation('current', 'Current account')],
    })
    const auth = useAuthStore()
    auth.user = makeUser('user-a', 'a@example.com')
    const chat = useChatStore()

    await chat.loadConversations()

    expect(chat.conversations).toEqual([makeConversation('current', 'Current account')])
  })

  it('does not commit stale selected-conversation hydration after logout', async () => {
    const pending = deferred<{ messages: ChatMessage[] }>()
    apiMock.mockImplementation((path) => {
      if (path === '/chat/conversations/conv-a/messages?limit=50') return pending.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const auth = useAuthStore()
    auth.user = makeUser('user-a', 'a@example.com')
    const chat = useChatStore()
    chat.messages = { 'conv-a': [makeMessage('old', 'Old message')] }

    const hydration = chat.openConversation('conv-a')
    expect(chat.selectedId).toBe('conv-a')

    auth.user = null
    pending.resolve({ messages: [makeMessage('stale', 'Stale account A message')] })
    await hydration

    expect(chat.selectedId).toBeNull()
    expect(chat.messages).toEqual({})
  })

  it('does not commit stale selected-conversation hydration into a different account', async () => {
    const pending = deferred<{ messages: ChatMessage[] }>()
    apiMock.mockImplementation((path) => {
      if (path === '/chat/conversations/conv-a/messages?limit=50') return pending.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })

    const auth = useAuthStore()
    auth.user = makeUser('user-a', 'a@example.com')
    const chat = useChatStore()

    const hydration = chat.openConversation('conv-a')
    auth.user = makeUser('user-b', 'b@example.com')

    pending.resolve({ messages: [makeMessage('stale', 'Account A thread')] })
    await hydration

    expect(chat.selectedId).toBeNull()
    expect(chat.messages).toEqual({})
  })

  it('clears account-owned chat state when the authenticated user changes', async () => {
    const auth = useAuthStore()
    auth.user = makeUser('user-a', 'a@example.com')
    const chat = useChatStore()
    chat.conversations = [makeConversation('conv-a', 'A chat')]
    chat.messages = { 'conv-a': [makeMessage('m1', 'Secret')] }
    chat.selectedId = 'conv-a'

    auth.user = makeUser('user-b', 'b@example.com')
    await Promise.resolve()

    expect(chat.conversations).toEqual([])
    expect(chat.messages).toEqual({})
    expect(chat.selectedId).toBeNull()
  })
})
