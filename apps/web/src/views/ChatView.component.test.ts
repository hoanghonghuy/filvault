/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ChatView from './ChatView.vue'
import { useAuthStore } from '@/stores/auth'

const { uploadToPresignedMock, resolvePendingUploads } = vi.hoisted(() => {
  const pendingResolvers: Array<() => void> = []
  const uploadToPresignedMock = vi.fn<
    (
      url: string,
      file: File,
      contentType: string,
      onProgress?: (ratio: number) => void,
      signal?: AbortSignal,
    ) => Promise<void>
  >(
    (
      _url: string,
      _file: File,
      _contentType: string,
      onProgress?: (ratio: number) => void,
      signal?: AbortSignal,
    ) =>
      new Promise<void>((resolve, reject) => {
        onProgress?.(0.25)
        pendingResolvers.push(() => {
          onProgress?.(1)
          resolve()
        })
        signal?.addEventListener('abort', () => {
          reject(new DOMException('Aborted', 'AbortError'))
        })
      }),
  )
  const resolvePendingUploads = () => {
    pendingResolvers.splice(0).forEach((resolve) => resolve())
  }
  return { uploadToPresignedMock, resolvePendingUploads }
})

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
    if (path.startsWith(`/chat/conversations/${conversation.id}/media`)) {
      return { media: [] }
    }
    if (path === '/chat/attachments/upload-sessions' && options?.method === 'POST') {
      return {
        fileId: '01KCHATFILE00000000000000001',
        uploadUrl: 'https://upload.example/presigned',
      }
    }
    if (path.startsWith('/chat/attachments/') && path.endsWith('/complete') && options?.method === 'POST') {
      const payload = JSON.parse(options.body as string)
      return {
        id: '01KCHATATTACHMESSAGE00000000001',
        conversationId: payload.conversationId,
        body: payload.body,
        senderId: '01KCHATAUTH00000000000000001',
        createdAt: new Date().toISOString(),
        attachments: [{ id: '01KCHATATTACH000000000000001', name: 'photo.jpg', mimeType: 'image/jpeg', sizeBytes: 1024 }],
      }
    }
    throw new Error(`Unexpected API path: ${path}`)
  }),
  formatBytes: (bytes: number) => `${bytes} B`,
  getAccessToken: () => 'test-token',
  uploadToPresigned: uploadToPresignedMock,
  normalizePresignedUrl: (url: string) => url,
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

let mockRoute = { params: {} as Record<string, string>, query: {}, path: '/chat' }
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn<() => void>(), replace: vi.fn<() => void>() }),
  useRoute: () => mockRoute,
}))

describe('ChatView', () => {
  let pinia: ReturnType<typeof createPinia>

  async function mountInThread() {
    mockRoute = { params: { id: conversation.id }, query: {}, path: `/chat/${conversation.id}` }
    const wrapper = mount(ChatView, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })
    await flushPromises()
    return wrapper
  }

  function setFileInputFiles(input: HTMLInputElement, files: File[]) {
    Object.defineProperty(input, 'files', { configurable: true, value: files })
  }

  beforeEach(() => {
    uploadToPresignedMock.mockClear()
    mockRoute = { params: {}, query: {}, path: '/chat' }
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

  it('loads thread messages and hides skeleton when reloading directly on /chat/:id', async () => {
    mockRoute = { params: { id: conversation.id }, query: {}, path: `/chat/${conversation.id}` }

    const wrapper = mount(ChatView, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })

    await flushPromises()

    expect(wrapper.find('.sk-chat-thread').exists()).toBe(false)
    expect(wrapper.find('.thread-empty-state').exists()).toBe(false)
    expect(wrapper.find('.message-bubble:not(.outgoing)').exists()).toBe(true)
    expect(wrapper.find('.message-bubble:not(.outgoing)').text()).toContain('Hello from Recipient')

    wrapper.unmount()
  })

  it('keeps the composer enabled while an attachment upload is in progress', async () => {
    const wrapper = await mountInThread()

    const fileInput = wrapper.find<HTMLInputElement>('#chat-attachment')
    const file = new File(['pixels'], 'photo.jpg', { type: 'image/jpeg' })
    setFileInputFiles(fileInput.element, [file])
    await fileInput.trigger('change')
    await flushPromises()

    expect(uploadToPresignedMock).toHaveBeenCalledTimes(1)
    expect(wrapper.find('.pending-attachment-row').exists()).toBe(true)

    const textarea = wrapper.find<HTMLTextAreaElement>('.chat-composer textarea')
    const attachBtn = wrapper.find('.attach-btn')
    expect(textarea.attributes('disabled')).toBeUndefined()
    expect(attachBtn.attributes('disabled')).toBeUndefined()

    wrapper.unmount()
  })

  it('sends text while a background attachment upload is still running', async () => {
    const wrapper = await mountInThread()

    const fileInput = wrapper.find<HTMLInputElement>('#chat-attachment')
    const file = new File(['pixels'], 'photo.jpg', { type: 'image/jpeg' })
    setFileInputFiles(fileInput.element, [file])
    await fileInput.trigger('change')
    await flushPromises()

    const textarea = wrapper.find<HTMLTextAreaElement>('.chat-composer textarea')
    await textarea.setValue('hello during upload')
    await wrapper.find('.send-btn').trigger('click')
    await flushPromises()

    const { api } = await import('@/api/client')
    expect(api).toHaveBeenCalledWith(
      `/chat/conversations/${conversation.id}/messages`,
      expect.objectContaining({
        method: 'POST',
        body: expect.stringContaining('"body":"hello during upload"'),
      }),
    )

    resolvePendingUploads()
    await flushPromises()

    wrapper.unmount()
  })

  it('queues multiple attachments and uploads them one at a time', async () => {
    const wrapper = await mountInThread()
    const fileInput = wrapper.find<HTMLInputElement>('#chat-attachment')

    const first = new File(['one'], 'one.jpg', { type: 'image/jpeg' })
    setFileInputFiles(fileInput.element, [first])
    await fileInput.trigger('change')
    await flushPromises()

    const second = new File(['two'], 'two.jpg', { type: 'image/jpeg' })
    setFileInputFiles(fileInput.element, [second])
    await fileInput.trigger('change')
    await flushPromises()

    expect(wrapper.findAll('.pending-attachment-row')).toHaveLength(2)
    expect(uploadToPresignedMock).toHaveBeenCalledTimes(1)

    resolvePendingUploads()
    await flushPromises()
    expect(uploadToPresignedMock).toHaveBeenCalledTimes(2)

    wrapper.unmount()
  })

  it('exposes a single screen-reader upload progress region for attachment queue state', async () => {
    const wrapper = await mountInThread()

    const fileInput = wrapper.find<HTMLInputElement>('#chat-attachment')
    const file = new File(['pixels'], 'photo.jpg', { type: 'image/jpeg' })
    setFileInputFiles(fileInput.element, [file])
    await fileInput.trigger('change')
    await flushPromises()

    expect(wrapper.find('.attachment-queue').exists()).toBe(false)
    expect(wrapper.find('.upload-progress.sr-only').exists()).toBe(true)
    expect(wrapper.find('.message-bubble.has-pending-attachment[aria-live]').exists()).toBe(false)

    wrapper.unmount()
  })

  it('supports keyboard navigation for shared media tabs in the info sidebar', async () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn<(query: string) => MediaQueryList>((query: string) => ({
        matches: query.includes('min-width: 1024px'),
        media: query,
        onchange: null,
        addListener: vi.fn<() => void>(),
        removeListener: vi.fn<() => void>(),
        dispatchEvent: vi.fn<() => boolean>(),
        addEventListener: vi.fn<() => void>(),
        removeEventListener: vi.fn<() => void>(),
      })),
    )

    const wrapper = await mountInThread()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.chat-info-sidebar').exists()).toBe(true)

    const mediaSection = wrapper.findAll('.chat-info-sidebar .menu-section')[1]
    const fileMenuBtn = mediaSection.findAll('.menu-row-item')[1]
    expect(fileMenuBtn.exists()).toBe(true)
    await fileMenuBtn.trigger('click')
    await flushPromises()

    const filesTab = wrapper.find('#shared-media-tab-file')
    expect(filesTab.exists()).toBe(true)
    expect(filesTab.attributes('aria-controls')).toBe('shared-media-panel-file')

    await filesTab.trigger('keydown', { key: 'ArrowLeft' })
    await flushPromises()

    expect(wrapper.find('#shared-media-panel-media[role="tabpanel"]').exists()).toBe(true)
    expect(document.activeElement?.id).toBe('shared-media-tab-media')

    wrapper.unmount()
  })

  it('supports keyboard resizing for the desktop conversation rail splitter', async () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn<(query: string) => MediaQueryList>((query: string) => ({
        matches: query.includes('min-width: 1024px'),
        media: query,
        onchange: null,
        addListener: vi.fn<() => void>(),
        removeListener: vi.fn<() => void>(),
        dispatchEvent: vi.fn<() => boolean>(),
        addEventListener: vi.fn<() => void>(),
        removeEventListener: vi.fn<() => void>(),
      })),
    )

    const wrapper = await mountInThread()
    const resizer = wrapper.find('.chat-resizer.right-edge')
    expect(resizer.exists()).toBe(true)
    expect(resizer.attributes('role')).toBe('separator')
    expect(resizer.attributes('tabindex')).toBe('0')

    const initialWidth = Number((wrapper.find('.chat-app').element as HTMLElement).style.getPropertyValue('--rail-w').replace('px', ''))
    await resizer.trigger('keydown', { key: 'ArrowRight' })
    const widerWidth = Number((wrapper.find('.chat-app').element as HTMLElement).style.getPropertyValue('--rail-w').replace('px', ''))
    expect(widerWidth).toBeGreaterThan(initialWidth)

    await resizer.trigger('keydown', { key: 'Home' })
    const resetWidth = Number((wrapper.find('.chat-app').element as HTMLElement).style.getPropertyValue('--rail-w').replace('px', ''))
    expect(resetWidth).toBe(320)

    wrapper.unmount()
  })
})
