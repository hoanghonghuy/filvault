/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { installRouteFocus } from './routeFocus'

type Hook = (to: unknown, from: unknown) => void

function makeRouter() {
  let hook: Hook | null = null
  return {
    router: {
      afterEach(next: Hook) {
        hook = next
        return () => {
          hook = null
        }
      },
    },
    navigate() {
      hook?.({}, {})
    },
  }
}

function flushFrame() {
  return new Promise<void>((resolve) => requestAnimationFrame(() => resolve()))
}

describe('installRouteFocus', () => {
  beforeEach(() => {
    document.body.innerHTML = '<button id="nav">Files</button><h1 class="header-title">Overview</h1>'
    vi.spyOn(window, 'requestAnimationFrame').mockImplementation((callback: FrameRequestCallback) => {
      callback(0)
      return 1
    })
    vi.spyOn(window, 'cancelAnimationFrame').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('does not steal focus on initial navigation but focuses page title afterward', async () => {
    const nav = document.querySelector<HTMLButtonElement>('#nav')!
    nav.focus()
    const { router, navigate } = makeRouter()
    const cleanup = installRouteFocus(router)

    navigate()
    expect(document.activeElement).toBe(nav)

    navigate()
    await flushFrame()
    expect(document.activeElement).toBe(document.querySelector('.header-title'))
    cleanup()
  })

  it('does not steal focus from an open aria-modal dialog', async () => {
    const modal = document.createElement('div')
    modal.setAttribute('aria-modal', 'true')
    modal.innerHTML = '<button id="modal-action">Confirm</button>'
    document.body.append(modal)
    const modalAction = document.querySelector<HTMLButtonElement>('#modal-action')!
    modalAction.focus()

    const { router, navigate } = makeRouter()
    const cleanup = installRouteFocus(router)
    navigate()
    navigate()
    await flushFrame()

    expect(document.activeElement).toBe(modalAction)
    cleanup()
  })

  it('ignores hidden aria-modal elements when deciding whether to focus the page title', async () => {
    const modal = document.createElement('div')
    modal.setAttribute('aria-modal', 'true')
    modal.hidden = true
    document.body.append(modal)

    const { router, navigate } = makeRouter()
    const cleanup = installRouteFocus(router)
    navigate()
    navigate()
    await flushFrame()

    expect(document.activeElement).toBe(document.querySelector('.header-title'))
    cleanup()
  })

  it('does not focus a hidden page title', async () => {
    const nav = document.querySelector<HTMLButtonElement>('#nav')!
    const title = document.querySelector<HTMLElement>('.header-title')!
    title.style.display = 'none'
    nav.focus()

    const { router, navigate } = makeRouter()
    const cleanup = installRouteFocus(router)
    navigate()
    navigate()
    await flushFrame()

    expect(document.activeElement).toBe(nav)
    cleanup()
  })

  it('does not focus a page title inside a hidden ancestor', async () => {
    const nav = document.querySelector<HTMLButtonElement>('#nav')!
    const title = document.querySelector<HTMLElement>('.header-title')!
    const wrapper = document.createElement('div')
    title.before(wrapper)
    wrapper.append(title)
    wrapper.style.display = 'none'
    nav.focus()

    const { router, navigate } = makeRouter()
    const cleanup = installRouteFocus(router)
    navigate()
    navigate()
    await flushFrame()

    expect(document.activeElement).toBe(nav)
    cleanup()
  })

  it('uses tabindex -1 only for programmatic focus and removes it after blur', async () => {
    const title = document.querySelector<HTMLElement>('.header-title')!
    const { router, navigate } = makeRouter()
    const cleanup = installRouteFocus(router)
    navigate()
    navigate()
    await flushFrame()

    expect(title.getAttribute('tabindex')).toBe('-1')
    title.blur()
    expect(title.hasAttribute('tabindex')).toBe(false)
    cleanup()
  })
})
