/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ChatView from './ChatView.vue'
import { useAuthStore } from '@/stores/auth'

const conversation = {
  id: '01KCHATCONVERSATION00000000001',
  title: '',
  type: 'direct' as const,
  peer: {
    id: '01KCHATPEER00000000000000001',
    name: 'Recipient',
    email: 'recipient@example.com',
  },
  createdAt: '2026-08-28T12:00:00.000Z',
  updatedAt: '2026-08-28T12:00:00.000Z',
  lastMessageAt: '2026-08-28T12:00:00.000Z',
  preview: {
    messageId: '01KCHATMESSAGE000000000000001',
    body: 'Hello from Recipient',
    createdAt: '2026-08-28T12:00:00.000Z',
    attachments: [],
  },
}

const messages = [
  {
    id: '01KCHATMESSAGE000000000000001',
    conversationId: conversation.id,
    body: 'Hello from Recipient',
    senderId: conversation.peer.id,
    createdAt: '2026-08-28T12:00:00.000Z',
    attachments: [],
  },
]

vi.mock('@/api/client', () => ({
  API_BASE: 'http://localhost:8080/api/v1',
  api: vi.fn<(path: string, options?: RequestInit) => Promise<unknown>>(async (path: string, options?: RequestInit) => {
    if (path === '/chat/conversations?includePreview=true') {
      return { conversations: [conversation] }
    }
    if (path.startsWith(`/chat/conversations/${conversation.id}/messages`)) {
      if (options?.method === 'POST') {
        const payload = JSON.parse(options.body as string)
        return {
          id: '01KNEWMESSAGE000000000000001',
          conversationId: conversation.id,
          body: payload.body,
          senderId: '01KCHATAUTH00000000000000001',
          clientMessageId: payload.clientMessageId,
          createdAt: new Date().toISOString(),
          attachments: [],
        }
      }
      return { messages, hasMore: false }
    }
    if (path === `/chat/conversations/${conversation.id}/media`) {
      return { media: [] }
    }
    throw new Error(`Unexpected API path: ${path}`)
  }),
  formatBytes: (bytes: number) => `${bytes} B`,
  getAccessToken: () => 'test-token',
  uploadToPresigned: vi.fn<() => void>(),
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn<() => void>() }),
}))

describe('ChatView', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    const auth = useAuthStore(pinia)
    auth.user = {
      id: '01KCHATAUTH00000000000000001',
      email: 'sender@example.com',
      displayName: 'Sender',
      emailVerified: true,
      storageUsed: 0,
      storageQuota: 1024,
      imageThumbnailsEnabled: true,
      videoThumbnailsEnabled: true,
      trashAutoDeleteEnabled: false,
      trashRetentionDays: 30,
      createdAt: '2026-08-28T12:00:00.000Z',
    }
    vi.stubGlobal(
      'fetch',
      vi.fn<() => Promise<Response>>().mockRejectedValue(new Error('test stream is not connected')),
    )
    vi.stubGlobal(
      'matchMedia',
      vi.fn<() => MediaQueryList>(() => ({
        matches: false,
        media: '',
        onchange: null,
        addListener: vi.fn<() => void>(),
        removeListener: vi.fn<() => void>(),
        dispatchEvent: vi.fn<() => boolean>(),
        addEventListener: vi.fn<() => void>(),
        removeEventListener: vi.fn<() => void>(),
      })),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders the inbox and incoming message bubble from the chat API', async () => {
    const wrapper = mount(ChatView, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })

    await flushPromises()

    expect(wrapper.find('.chat-app').exists()).toBe(true)
    expect(wrapper.find('.chat-rail').exists()).toBe(true)
    expect(wrapper.find('.message-thread').exists()).toBe(true)
    expect(wrapper.find('.conversation-title').text()).toBe('Recipient')
    expect(wrapper.find('.message-bubble:not(.outgoing)').text()).toContain('Hello from Recipient')

    wrapper.unmount()
  })

  it('sends message successfully even in insecure context where crypto.randomUUID is undefined', async () => {
    const originalCrypto = globalThis.crypto
    vi.stubGlobal('crypto', {
      getRandomValues: (arr: Uint8Array) => originalCrypto.getRandomValues(arr),
    })

    const wrapper = mount(ChatView, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })

    await flushPromises()

    const textarea = wrapper.find<HTMLTextAreaElement>('.chat-composer textarea')
    expect(textarea.exists()).toBe(true)
    await textarea.setValue('helu')

    const sendBtn = wrapper.find('.send-btn')
    expect(sendBtn.exists()).toBe(true)
    await sendBtn.trigger('click')

    await flushPromises()

    const { api } = await import('@/api/client')
    expect(api).toHaveBeenCalledWith(
      `/chat/conversations/${conversation.id}/messages`,
      expect.objectContaining({
        method: 'POST',
        body: expect.stringContaining('"body":"helu"'),
      }),
    )

    wrapper.unmount()
  })
})
