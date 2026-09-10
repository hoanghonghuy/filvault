/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { installCallModalFocusGuard } from './callModalFocusGuard'

const flush = async () => {
  await Promise.resolve()
  await Promise.resolve()
}

describe('installCallModalFocusGuard', () => {
  let cleanup: () => void

  beforeEach(() => {
    document.body.innerHTML = '<button id="trigger">Start call</button>'
    ;(document.querySelector('#trigger') as HTMLButtonElement).focus()
    cleanup = installCallModalFocusGuard()
  })

  afterEach(() => {
    cleanup()
    document.body.innerHTML = ''
  })

  it('prefers the incoming accept action and restores invoking focus on close', async () => {
    const overlay = document.createElement('div')
    overlay.className = 'call-overlay'
    overlay.setAttribute('role', 'dialog')
    overlay.setAttribute('aria-modal', 'true')
    overlay.innerHTML = `
      <button class="call-btn btn-decline">Decline</button>
      <button class="call-btn btn-accept">Accept</button>
    `
    document.body.append(overlay)

    await flush()
    expect(document.activeElement).toBe(overlay.querySelector('.btn-accept'))

    overlay.remove()
    await flush()
    expect(document.activeElement?.id).toBe('trigger')
  })

  it('wraps Tab and Shift+Tab inside the modal', async () => {
    const overlay = document.createElement('div')
    overlay.className = 'call-overlay'
    overlay.setAttribute('aria-modal', 'true')
    overlay.innerHTML = `
      <button class="header-action-btn">Fullscreen</button>
      <button class="control-btn">Microphone</button>
      <button class="control-btn btn-hangup">Hang up</button>
    `
    document.body.append(overlay)
    await flush()

    const buttons = overlay.querySelectorAll<HTMLButtonElement>('button')
    const first = buttons[0]
    const last = buttons[2]

    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(first)

    first.focus()
    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true }),
    )
    expect(document.activeElement).toBe(last)
  })

  it('suppresses Escape and avoids initially focusing a destructive-only outgoing action', async () => {
    const overlay = document.createElement('div')
    overlay.className = 'call-overlay'
    overlay.setAttribute('aria-modal', 'true')
    overlay.innerHTML = '<button class="call-btn btn-decline">Cancel call</button>'
    document.body.append(overlay)
    await flush()

    expect(document.activeElement).toBe(overlay)

    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true })
    document.dispatchEvent(event)
    expect(event.defaultPrevented).toBe(true)
    expect(overlay.isConnected).toBe(true)
  })
})
