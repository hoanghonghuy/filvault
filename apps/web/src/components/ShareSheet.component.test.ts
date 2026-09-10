/**
 * @vitest-environment jsdom
 */
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ShareSheet from './ShareSheet.vue'

const existingLink = {
  url: '/s/public-token',
  expiresAt: null,
  createdAt: '2026-09-10T00:00:00.000Z',
}

function mountSheet(options: {
  existing?: typeof existingLink | null
  onCreate?: (ttl: unknown) => unknown
  onCopy?: (url: unknown) => unknown
  onRevoke?: () => unknown
} = {}) {
  return mount(ShareSheet, {
    props: {
      open: true,
      name: 'Report.pdf',
      existing: options.existing ?? null,
    },
    attrs: {
      onCreate: options.onCreate,
      onCopy: options.onCopy,
      onRevoke: options.onRevoke,
    },
    attachTo: document.body,
    global: {
      stubs: { Teleport: true },
    },
  })
}

function deferred() {
  let resolve!: () => void
  const promise = new Promise<void>((done) => {
    resolve = done
  })
  return { promise, resolve }
}

afterEach(() => {
  document.body.innerHTML = ''
})

describe('ShareSheet owner operation lifecycle', () => {
  it('tracks the real create promise and suppresses duplicate creates', async () => {
    const pending = deferred()
    const onCreate = vi.fn(() => pending.promise)
    const wrapper = mountSheet({ onCreate })
    const create = wrapper.get('.btn.ink')

    void create.trigger('click')
    await flushPromises()

    expect(onCreate).toHaveBeenCalledTimes(1)
    expect(onCreate).toHaveBeenCalledWith(null)
    expect(create.text()).toBe('Creating…')
    expect(create.attributes('aria-busy')).toBe('true')
    expect(create.attributes('disabled')).toBeDefined()

    void create.trigger('click')
    await flushPromises()
    expect(onCreate).toHaveBeenCalledTimes(1)

    pending.resolve()
    await flushPromises()
    expect(create.text()).toBe('Create link')
    expect(create.attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('always releases create busy state when a listener rejects', async () => {
    const onCreate = vi.fn(() => Promise.reject(new Error('handled by parent')))
    const wrapper = mountSheet({ onCreate })
    const create = wrapper.get('.btn.ink')

    await create.trigger('click')
    await flushPromises()

    expect(onCreate).toHaveBeenCalledTimes(1)
    expect(create.text()).toBe('Create link')
    expect(create.attributes('aria-busy')).toBeUndefined()
    expect(create.attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('serializes copy and revoke actions against their real handlers', async () => {
    const copyPending = deferred()
    const revokePending = deferred()
    const onCopy = vi.fn(() => copyPending.promise)
    const onRevoke = vi.fn(() => revokePending.promise)
    const wrapper = mountSheet({ existing: existingLink, onCopy, onRevoke })

    const copy = wrapper.get('button[aria-label="Copy link"]')
    void copy.trigger('click')
    await flushPromises()
    expect(onCopy).toHaveBeenCalledTimes(1)
    expect(onCopy).toHaveBeenCalledWith(existingLink.url)
    expect(copy.attributes('aria-label')).toBe('Copying link…')
    expect(copy.attributes('aria-busy')).toBe('true')

    void copy.trigger('click')
    await flushPromises()
    expect(onCopy).toHaveBeenCalledTimes(1)

    copyPending.resolve()
    await flushPromises()

    const revoke = wrapper.get('.btn.danger')
    void revoke.trigger('click')
    await flushPromises()
    expect(onRevoke).toHaveBeenCalledTimes(1)
    expect(revoke.text()).toBe('Revoking…')
    expect(revoke.attributes('aria-busy')).toBe('true')

    void revoke.trigger('click')
    await flushPromises()
    expect(onRevoke).toHaveBeenCalledTimes(1)

    revokePending.resolve()
    await flushPromises()
    expect(revoke.text()).toBe('Revoke link')
    expect(revoke.attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('preserves the TTL roving-radio keyboard contract', async () => {
    const wrapper = mountSheet({ onCreate: vi.fn() })
    const radios = wrapper.findAll<HTMLButtonElement>('[role="radio"]')

    expect(radios[0]?.attributes('aria-checked')).toBe('true')
    await radios[0]?.trigger('keydown', { key: 'ArrowRight' })

    expect(radios[1]?.attributes('aria-checked')).toBe('true')
    expect(radios[1]?.attributes('tabindex')).toBe('0')
    expect(document.activeElement).toBe(radios[1]?.element)
    wrapper.unmount()
  })
})
