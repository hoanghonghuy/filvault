/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import SharedWithMeView from './SharedWithMeView.vue'
import { api } from '@/api/client'
import { setLocale } from '@/lib/i18n'

const uiSpies = vi.hoisted(() => ({
  showToast: vi.fn<(message: string, type?: 'success' | 'error' | 'info') => void>(),
  confirm: vi.fn<(options: { title: string; message: string; confirmLabel?: string; danger?: boolean }) => Promise<boolean>>(),
  openActionSheet: vi.fn<(title: string, actions: Array<{ id: string; label: string; icon?: string; danger?: boolean }>) => Promise<string | null>>(),
}))

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    api: vi.fn<typeof actual.api>(),
    formatBytes: (bytes: number) => `${bytes} B`,
  }
})

vi.mock('@/stores/ui', () => ({
  useUiStore: () => uiSpies,
}))

const mockedApi = vi.mocked(api)
const link = {
  token: 'token-1',
  fileId: 'file-1',
  fileName: 'report.pdf',
  expiresAt: null,
  createdAt: '2026-09-10T00:00:00Z',
}

function setupApi() {
  mockedApi.mockImplementation(async (path: string, options?: RequestInit) => {
    if (path === '/shares/with-me') return { shares: [] } as never
    if (path === '/share-links') return { links: [link] } as never
    if (path === '/files/file-1/share' && options?.method === 'DELETE') return {} as never
    throw new Error(`Unexpected API call: ${path}`)
  })
}

async function mountView() {
  const wrapper = mount(SharedWithMeView, {
    global: {
      stubs: {
        Icon: { template: '<span aria-hidden="true" />' },
        Transition: { template: '<slot />' },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('SharedWithMeView localization', () => {
  beforeEach(() => {
    setLocale('vi')
    mockedApi.mockReset()
    uiSpies.showToast.mockReset()
    uiSpies.confirm.mockReset()
    uiSpies.openActionSheet.mockReset()
    setupApi()
  })

  it('reactively updates share copy, expiry text and accessibility labels', async () => {
    const wrapper = await mountView()

    expect(wrapper.get('[role="tablist"]').attributes('aria-label')).toBe('Các chế độ xem chia sẻ')
    expect(wrapper.get('.share-item-sub').text()).toContain('Có hiệu lực vĩnh viễn')
    expect(wrapper.get('.share-action-btn').attributes('aria-label')).toBe('Thao tác chia sẻ')

    setLocale('en')
    await flushPromises()

    expect(wrapper.get('[role="tablist"]').attributes('aria-label')).toBe('Share views')
    expect(wrapper.get('.share-item-sub').text()).toContain('Active forever')
    expect(wrapper.get('.share-action-btn').attributes('aria-label')).toBe('Share actions')
  })

  it('localizes action-sheet and destructive revoke copy without changing semantics', async () => {
    uiSpies.openActionSheet.mockResolvedValue('revoke')
    uiSpies.confirm.mockResolvedValue(true)
    const wrapper = await mountView()

    await wrapper.get('.share-action-btn').trigger('click')
    await flushPromises()

    expect(uiSpies.openActionSheet).toHaveBeenCalledWith('report.pdf', [
      { id: 'copy', label: 'Sao chép liên kết', icon: 'copy' },
      { id: 'open', label: 'Mở liên kết', icon: 'external-link' },
      { id: 'revoke', label: 'Thu hồi liên kết', icon: 'trash', danger: true },
    ])
    expect(uiSpies.confirm).toHaveBeenCalledWith({
      title: 'Thu hồi liên kết?',
      message: '“report.pdf” sẽ không còn được chia sẻ công khai.',
      confirmLabel: 'Thu hồi liên kết',
      danger: true,
    })
    expect(mockedApi).toHaveBeenCalledWith('/files/file-1/share', { method: 'DELETE' })
    expect(uiSpies.showToast).toHaveBeenCalledWith('Đã thu hồi liên kết')
  })
})
