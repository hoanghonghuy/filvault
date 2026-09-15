/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import NotFoundView from './NotFoundView.vue'
import { setLocale } from '@/lib/i18n'

const authState = vi.hoisted(() => ({ authenticated: false }))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    get isAuthenticated() {
      return authState.authenticated
    },
  }),
}))

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', name: 'login', component: { template: '<div>Login</div>' } },
      { path: '/', name: 'home', component: { template: '<div>Home</div>' } },
      { path: '/files', name: 'files', component: { template: '<div>Files</div>' } },
      { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFoundView },
    ],
  })
}

describe('NotFoundView', () => {
  beforeEach(() => {
    authState.authenticated = false
    setLocale('en')
    window.history.replaceState({}, '', '/')
  })

  it('offers a privacy-safe sign-in recovery for guests', async () => {
    const router = makeRouter()
    await router.push('/missing/private-looking-path')
    await router.isReady()

    const wrapper = mount(NotFoundView, { global: { plugins: [router] } })

    expect(wrapper.text()).toContain('This path does not exist')
    expect(wrapper.text()).toContain('Sign in')
    expect(wrapper.text()).not.toContain('Open My Files')
    expect(wrapper.find('a[href="/login"]').exists()).toBe(true)

    wrapper.unmount()
  })

  it('offers Home and My Files recovery for authenticated users', async () => {
    authState.authenticated = true
    const router = makeRouter()
    await router.push('/missing')
    await router.isReady()

    const wrapper = mount(NotFoundView, { global: { plugins: [router] } })

    expect(wrapper.text()).toContain('Go home')
    expect(wrapper.text()).toContain('Open My Files')
    expect(wrapper.text()).not.toContain('Sign in')
    expect(wrapper.find('a[href="/"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/files"]').exists()).toBe(true)

    wrapper.unmount()
  })

  it('reacts to the active locale', async () => {
    const router = makeRouter()
    await router.push('/missing')
    await router.isReady()
    const wrapper = mount(NotFoundView, { global: { plugins: [router] } })

    setLocale('vi')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Đường dẫn này không tồn tại')
    expect(wrapper.text()).toContain('Đăng nhập')

    wrapper.unmount()
  })
})
