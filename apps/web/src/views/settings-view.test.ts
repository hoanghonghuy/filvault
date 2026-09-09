/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import SettingsView from './SettingsView.vue'
import { useAuthStore } from '@/stores/auth'
import { setLocale } from '@/lib/i18n'
import type { User } from '@/api/types'

const baseUser: User = {
  id: '01KSETTINGSUSER00000000000001',
  email: 'user@example.com',
  displayName: 'User',
  emailVerified: true,
  storageUsed: 0,
  storageQuota: 1024,
  imageThumbnailsEnabled: true,
  videoThumbnailsEnabled: true,
  trashAutoDeleteEnabled: false,
  trashRetentionDays: 30,
  createdAt: '2026-01-01T00:00:00.000Z',
}

const apiMock = vi.fn<(path: string, options?: RequestInit) => Promise<unknown>>()

vi.mock('@/api/client', () => ({
  api: (path: string, options?: RequestInit) => apiMock(path, options),
  formatBytes: (bytes: number) => `${bytes} B`,
}))

vi.mock('@/api/errors', () => ({
  formatApiError: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('vue-router', () => ({
  RouterLink: {
    name: 'RouterLink',
    props: ['to'],
    template: '<a><slot /></a>',
  },
  onBeforeRouteLeave: vi.fn<() => void>(),
}))

vi.mock('@/lib/theme', () => ({
  THEMES: [
    {
      id: 'default',
      nameKey: 'themeDefault',
      nameDefault: 'Default',
      swatchGradient: 'linear-gradient(#000, #fff)',
    },
  ],
  useTheme: () => ({
    currentColorTheme: { value: 'default' },
    applyColorTheme: vi.fn<(id: string) => void>(),
  }),
}))

vi.mock('@/components/AppIcon.vue', () => ({
  default: { name: 'AppIcon', template: '<span />' },
}))

function mountSettings() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const auth = useAuthStore(pinia)
  auth.user = { ...baseUser }
  return mount(SettingsView, {
    global: {
      plugins: [pinia],
    },
  })
}

function findTrashSaveButton(wrapper: ReturnType<typeof mount>) {
  return wrapper
    .findAll('button')
    .find((button) => button.text().includes('Save trash settings') || button.text().includes('Lưu cài đặt thùng rác'))
}

function findPreviewSaveButton(wrapper: ReturnType<typeof mount>) {
  return wrapper
    .findAll('button')
    .find((button) => button.text().includes('Save preview settings') || button.text().includes('Lưu cài đặt xem trước'))
}

function findCheckboxByLabel(wrapper: ReturnType<typeof mount>, label: string) {
  return wrapper
    .findAll('input[type="checkbox"]')
    .find((input) => input.element.closest('label')?.textContent?.includes(label))
}

describe('SettingsView explicit settings', () => {
  beforeEach(() => {
    setLocale('en')
    apiMock.mockReset()
    apiMock.mockImplementation(async (path: string) => {
      if (path === '/share-links') return { links: [] }
      if (path.startsWith('/activity')) return { events: [] }
      throw new Error(`Unexpected API path: ${path}`)
    })
    document.documentElement.dataset.theme = 'light'
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('disables section save buttons until that section is dirty', async () => {
    const wrapper = mountSettings()
    await flushPromises()

    expect(findTrashSaveButton(wrapper)?.attributes('disabled')).toBeDefined()
    expect(findPreviewSaveButton(wrapper)?.attributes('disabled')).toBeDefined()

    await wrapper.get('input[type="number"]').setValue(14)
    expect(findTrashSaveButton(wrapper)?.attributes('disabled')).toBeUndefined()
    expect(findPreviewSaveButton(wrapper)?.attributes('disabled')).toBeDefined()
  })

  it('saves only preview fields and keeps pending trash edits unsaved', async () => {
    apiMock.mockImplementation(async (path: string, options?: RequestInit) => {
      if (path === '/share-links') return { links: [] }
      if (path.startsWith('/activity')) return { events: [] }
      if (path === '/users/me' && options?.method === 'PATCH') {
        const body = JSON.parse(options.body as string)
        return {
          ...baseUser,
          ...body,
        }
      }
      throw new Error(`Unexpected API path: ${path}`)
    })

    const wrapper = mountSettings()
    await flushPromises()

    await wrapper.get('input[type="number"]').setValue(14)
    const previewToggle = findCheckboxByLabel(wrapper, 'Show image thumbnails')
    expect(previewToggle).toBeTruthy()
    await previewToggle!.setValue(false)

    await findPreviewSaveButton(wrapper)!.trigger('click')
    await flushPromises()

    const patchCalls = apiMock.mock.calls.filter(([path, options]) => path === '/users/me' && options?.method === 'PATCH')
    expect(patchCalls).toHaveLength(1)
    const patchBody = JSON.parse(patchCalls[0]![1]!.body as string)
    expect(patchBody).toEqual({
      imageThumbnailsEnabled: false,
      videoThumbnailsEnabled: true,
    })
    expect(patchBody).not.toHaveProperty('trashRetentionDays')
    expect(patchBody).not.toHaveProperty('trashAutoDeleteEnabled')

    expect(findTrashSaveButton(wrapper)?.attributes('disabled')).toBeUndefined()
    expect(wrapper.get('input[type="number"]').element).toHaveProperty('value', '14')
  })

  it('keeps edited values and shows section-scoped error on failed save', async () => {
    apiMock.mockImplementation(async (path: string, options?: RequestInit) => {
      if (path === '/share-links') return { links: [] }
      if (path.startsWith('/activity')) return { events: [] }
      if (path === '/users/me' && options?.method === 'PATCH') {
        throw new Error('network')
      }
      throw new Error(`Unexpected API path: ${path}`)
    })

    const wrapper = mountSettings()
    await flushPromises()

    await wrapper.get('input[type="number"]').setValue(21)
    await findTrashSaveButton(wrapper)!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Could not save trash settings')
    expect(wrapper.get('input[type="number"]').element).toHaveProperty('value', '21')
    expect(findTrashSaveButton(wrapper)?.attributes('disabled')).toBeUndefined()
  })

  it('refreshes canonical auth state for the saved section without overwriting unrelated unsaved values', async () => {
    apiMock.mockImplementation(async (path: string, options?: RequestInit) => {
      if (path === '/share-links') return { links: [] }
      if (path.startsWith('/activity')) return { events: [] }
      if (path === '/users/me' && options?.method === 'PATCH') {
        const body = JSON.parse(options.body as string)
        return {
          ...baseUser,
          trashAutoDeleteEnabled: true,
          trashRetentionDays: 7,
          ...body,
        }
      }
      throw new Error(`Unexpected API path: ${path}`)
    })

    const pinia = createPinia()
    setActivePinia(pinia)
    const auth = useAuthStore(pinia)
    auth.user = { ...baseUser }

    const wrapper = mount(SettingsView, { global: { plugins: [pinia] } })
    await flushPromises()

    const previewToggle = findCheckboxByLabel(wrapper, 'Show image thumbnails')
    await previewToggle!.setValue(false)

    const trashToggle = findCheckboxByLabel(wrapper, 'Auto-delete trash')
    await trashToggle!.setValue(true)
    await wrapper.get('input[type="number"]').setValue(7)
    await findTrashSaveButton(wrapper)!.trigger('click')
    await flushPromises()

    expect(auth.user?.trashAutoDeleteEnabled).toBe(true)
    expect(auth.user?.trashRetentionDays).toBe(7)
    expect(auth.user?.imageThumbnailsEnabled).toBe(false)
    expect(findPreviewSaveButton(wrapper)?.attributes('disabled')).toBeUndefined()
  })
})
