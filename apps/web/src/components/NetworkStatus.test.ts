/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import NetworkStatus from './NetworkStatus.vue'
import { setLocale } from '@/lib/i18n'

function setOnline(value: boolean) {
  Object.defineProperty(window.navigator, 'onLine', {
    configurable: true,
    value,
  })
}

describe('NetworkStatus', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setLocale('en')
    setOnline(true)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('stays hidden during normal connected use', () => {
    const wrapper = mount(NetworkStatus)
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows a persistent localized offline state when mounted offline', async () => {
    setOnline(false)
    setLocale('vi')
    const wrapper = mount(NetworkStatus)
    await wrapper.vm.$nextTick()

    const status = wrapper.get('[role="status"]')
    expect(status.text()).toContain('Bạn đang ngoại tuyến')
    expect(status.text()).toContain('Các thao tác cần mạng')

    vi.advanceTimersByTime(10_000)
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('announces recovery after offline and dismisses it automatically', async () => {
    const wrapper = mount(NetworkStatus)

    window.dispatchEvent(new Event('offline'))
    await wrapper.vm.$nextTick()
    expect(wrapper.get('[role="status"]').text()).toContain('You’re offline')

    window.dispatchEvent(new Event('online'))
    await wrapper.vm.$nextTick()
    expect(wrapper.get('[role="status"]').text()).toContain('Back online')

    vi.advanceTimersByTime(3000)
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('removes listeners and recovery timer on unmount', async () => {
    const removeSpy = vi.spyOn(window, 'removeEventListener')
    const wrapper = mount(NetworkStatus)

    window.dispatchEvent(new Event('offline'))
    window.dispatchEvent(new Event('online'))
    await wrapper.vm.$nextTick()
    wrapper.unmount()

    expect(removeSpy).toHaveBeenCalledWith('offline', expect.any(Function))
    expect(removeSpy).toHaveBeenCalledWith('online', expect.any(Function))
    vi.runOnlyPendingTimers()
  })
})
