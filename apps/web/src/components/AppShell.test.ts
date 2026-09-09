/**
 * @vitest-environment jsdom
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AppShell from './AppShell.vue'
import { useAuthStore } from '@/stores/auth'
import {
  SHELL_BREAKPOINTS,
  SHELL_NAV,
  SHELL_NAV_WIDTH,
  isShellNavActive,
} from '@/lib/shellNav'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

vi.mock('@/stores/chat', () => ({
  useChatStore: () => ({
    connectEvents: vi.fn<() => void>(),
    stopEvents: vi.fn<() => void>(),
  }),
}))

const verifiedUser = {
  id: 'user-1',
  email: 'user@example.com',
  displayName: 'Very Long Display Name That Should Truncate In Chrome',
  emailVerified: true,
  storageUsed: 0,
  storageQuota: 1024,
  imageThumbnailsEnabled: true,
  videoThumbnailsEnabled: true,
  trashAutoDeleteEnabled: false,
  trashRetentionDays: 30,
  createdAt: '2026-08-28T12:00:00.000Z',
}

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'overview', component: { template: '<div>Overview</div>' } },
      { path: '/files/:folderId?', name: 'files', component: { template: '<div>Files</div>' } },
      { path: '/photos', name: 'photos', component: { template: '<div>Photos</div>' } },
      { path: '/shared', name: 'shared', component: { template: '<div>Shared</div>' } },
      { path: '/settings', name: 'settings', component: { template: '<div>Settings</div>' } },
      { path: '/profile', name: 'profile', component: { template: '<div>Profile</div>' } },
      { path: '/chat', name: 'chat', component: { template: '<div>Chat</div>' } },
    ],
  })
}

describe('shell breakpoint contract (AppShell.vue)', () => {
  const shell = readSrc('./AppShell.vue')
  const mainCss = readSrc('../assets/main.css')

  it('documents three shell form factors via CSS breakpoints', () => {
    expect(shell).toMatch(/@media \(min-width: 768px\) and \(max-width: 1023\.98px\)/)
    expect(shell).toMatch(/@media \(min-width: 1024px\)/)
    expect(mainCss).toContain('--bp-tablet: 768px')
    expect(mainCss).toContain('--bp-desktop: 1024px')
    expect(mainCss).toContain('--nav-rail-w: 80px')
    expect(mainCss).toContain('--nav-sidebar-w: 220px')
  })

  it('uses a compact rail width on tablet and expanded sidebar on desktop', () => {
    const tabletBlock =
      shell.match(/@media \(min-width: 768px\) and \(max-width: 1023\.98px\)[\s\S]*?(?=@media)/)?.[0] ?? ''
    const desktopBlock = shell.match(/@media \(min-width: 1024px\)[\s\S]*?(?=@media \(min-width: 768px\))/s)?.[0] ?? ''
    expect(tabletBlock).toContain('width: var(--nav-rail-w)')
    expect(desktopBlock).toContain('width: var(--nav-sidebar-w)')
  })

  it('keeps mobile bottom nav and safe-area padding', () => {
    expect(shell).toContain('.bottom-nav')
    expect(shell).toMatch(/padding-bottom:\s*calc\(var\(--bottom-nav-h\) \+ env\(safe-area-inset-bottom\)\)/)
    expect(shell).toMatch(/env\(safe-area-inset-bottom\)/)
  })

  it('hides bottom nav from tablet upward while keeping page header on tablet', () => {
    const tabletBlock =
      shell.match(/@media \(min-width: 768px\) and \(max-width: 1023\.98px\)[\s\S]*?(?=@media)/)?.[0] ?? ''
    const desktopBlock = shell.match(/@media \(min-width: 1024px\)[\s\S]*?(?=@media \(min-width: 768px\))/s)?.[0] ?? ''
    expect(tabletBlock).toContain('.bottom-nav')
    expect(tabletBlock).toMatch(/\.bottom-nav\s*\{[^}]*display:\s*none/)
    expect(tabletBlock).not.toMatch(/\.header\s*\{[^}]*display:\s*none/)
    expect(desktopBlock).toMatch(/\.header\s*\{[^}]*display:\s*none/)
  })

  it('exposes accessible labels on shell navigation links', () => {
    expect(shell).toContain(':aria-label="navLabels[item.to]"')
    expect(shell).toContain(':title="navLabels[item.to]"')
  })
})

describe('isShellNavActive', () => {
  it('matches home only on the overview route', () => {
    expect(isShellNavActive('/', '/')).toBe(true)
    expect(isShellNavActive('/files', '/')).toBe(false)
  })

  it('matches nested routes for non-home destinations', () => {
    expect(isShellNavActive('/files', '/files')).toBe(true)
    expect(isShellNavActive('/files/folder/abc', '/files')).toBe(true)
    expect(isShellNavActive('/photos/albums/01', '/photos')).toBe(true)
    expect(isShellNavActive('/settings', '/files')).toBe(false)
  })
})

describe('SHELL_BREAKPOINTS', () => {
  it('aligns with DESIGN.md shell ranges', () => {
    expect(SHELL_BREAKPOINTS.tabletMin).toBe(768)
    expect(SHELL_BREAKPOINTS.desktopMin).toBe(1024)
    expect(SHELL_NAV_WIDTH.rail).toBeGreaterThanOrEqual(72)
    expect(SHELL_NAV_WIDTH.rail).toBeLessThanOrEqual(88)
    expect(SHELL_NAV_WIDTH.sidebar).toBe(220)
  })
})

describe('AppShell navigation state', () => {
  let pinia: ReturnType<typeof createPinia>
  let router: ReturnType<typeof createTestRouter>

  beforeEach(async () => {
    pinia = createPinia()
    setActivePinia(pinia)
    router = createTestRouter('/files')
    const auth = useAuthStore(pinia)
    auth.user = verifiedUser
    await router.push('/files')
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  async function mountShell() {
    const wrapper = mount(AppShell, {
      global: {
        plugins: [pinia, router],
        stubs: {
          StorageBar: true,
          ToastHost: true,
          GlobalConfirm: true,
          GlobalPrompt: true,
          GlobalActionSheet: true,
          CallModal: true,
        },
      },
    })
    await flushPromises()
    return wrapper
  }

  it('renders every core shell destination in bottom and side navigation', async () => {
    const wrapper = await mountShell()
    for (const item of SHELL_NAV) {
      const bottom = wrapper.findAll(`.bottom-link[href="${item.to}"]`)
      const side = wrapper.findAll(`.side-link[href="${item.to}"]`)
      expect(bottom.length, `bottom nav ${item.to}`).toBe(1)
      expect(side.length, `side nav ${item.to}`).toBe(1)
    }
    wrapper.unmount()
  })

  it('marks the current route active in shared navigation chrome', async () => {
    const wrapper = await mountShell()
    const activeBottom = wrapper.find('.bottom-link.router-link-active')
    const activeSide = wrapper.find('.side-link.router-link-active')
    expect(activeBottom.attributes('href')).toBe('/files')
    expect(activeSide.attributes('href')).toBe('/files')
    wrapper.unmount()
  })

  it('keeps nested file routes mapped to the Files destination', async () => {
    await router.push('/files/folder-abc')
    const wrapper = await mountShell()
    expect(wrapper.find('.bottom-link[href="/files"]').classes()).toContain('router-link-active')
    expect(wrapper.find('.side-link[href="/files"]').classes()).toContain('router-link-active')
    wrapper.unmount()
  })

  it('does not mark Home active on non-overview routes', async () => {
    await router.push('/settings')
    const wrapper = await mountShell()
    const homeBottom = wrapper.find('.bottom-link[href="/"]')
    const homeSide = wrapper.find('.side-link[href="/"]')
    expect(homeBottom.classes()).not.toContain('router-link-active')
    expect(homeSide.classes()).not.toContain('router-link-active')
    expect(wrapper.find('.bottom-link.router-link-active').attributes('href')).toBe('/settings')
    wrapper.unmount()
  })
})
