/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ChatView from './ChatView.vue'
import { useAuthStore } from '@/stores/auth'
import { setLocale } from '@/lib/i18n'

const STORAGE_KEY_WALLPAPERS = 'filvault.chat.wallpapers'

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
        throw new Error('Unexpected POST in wallpaper i18n test')
      }
      return { messages, hasMore: false }
    }
    if (path.startsWith(`/chat/conversations/${conversation.id}/media`)) {
      return { media: [] }
    }
    throw new Error(`Unexpected API path: ${path}`)
  }),
  formatBytes: (bytes: number) => `${bytes} B`,
  getAccessToken: () => 'test-token',
  uploadToPresigned: vi.fn<() => Promise<void>>(),
  normalizePresignedUrl: (url: string) => url,
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

let mockRoute = { params: { id: conversation.id } as Record<string, string>, query: {}, path: `/chat/${conversation.id}` }
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn<() => void>(), replace: vi.fn<() => void>() }),
  useRoute: () => mockRoute,
}))

function enableDesktopLayout() {
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
}

function findWallpaperBadge(wrapper: ReturnType<typeof mount>) {
  const sidebar = wrapper.find('.chat-info-sidebar')
  const rows = sidebar.findAll('.menu-row-item')
  const wallpaperRow = rows.find((row) => {
    const text = row.text()
    return text.includes('Hình nền đoạn chat') || text.includes('Chat wallpaper')
  })
  expect(wallpaperRow).toBeTruthy()
  return wallpaperRow!.find('.menu-badge')
}

describe('ChatView wallpaper label localization', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    setLocale('vi')
    localStorage.clear()
    mockRoute = { params: { id: conversation.id }, query: {}, path: `/chat/${conversation.id}` }
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
    enableDesktopLayout()
    vi.stubGlobal(
      'fetch',
      vi.fn<() => Promise<Response>>().mockRejectedValue(new Error('test stream is not connected')),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    localStorage.clear()
  })

  async function mountThreadWithInfoOpen() {
    const wrapper = mount(ChatView, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })
    await flushPromises()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.chat-info-sidebar').exists()).toBe(true)
    return wrapper
  }

  it('reactively updates the default wallpaper menu badge from Vietnamese to English without remounting', async () => {
    const wrapper = await mountThreadWithInfoOpen()
    const badge = findWallpaperBadge(wrapper)

    expect(badge.text()).toBe('Mặc định')

    setLocale('en')
    await flushPromises()

    expect(findWallpaperBadge(wrapper).text()).toBe('Default')

    wrapper.unmount()
  })

  it('reactively updates the custom image wallpaper menu badge from Vietnamese to English without remounting', async () => {
    localStorage.setItem(
      STORAGE_KEY_WALLPAPERS,
      JSON.stringify({ [conversation.id]: 'https://cdn.example.com/custom-wallpaper.jpg' }),
    )

    const wrapper = await mountThreadWithInfoOpen()
    const badge = findWallpaperBadge(wrapper)

    expect(badge.text()).toBe('Ảnh tùy chỉnh')

    setLocale('en')
    await flushPromises()

    expect(findWallpaperBadge(wrapper).text()).toBe('Custom image')

    wrapper.unmount()
  })
})
