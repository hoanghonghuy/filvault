/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { installSharedTabKeyboard } from './sharedTabKeyboard'

const flush = async () => {
  await Promise.resolve()
  await Promise.resolve()
}

function mountSharedTabs() {
  document.body.innerHTML = `
    <div class="shared-page">
      <div role="tablist" aria-label="Shares views">
        <button role="tab" aria-selected="true">My shares</button>
        <button role="tab" aria-selected="false">Shared with me</button>
      </div>
    </div>
    <div class="other-page">
      <div role="tablist">
        <button role="tab" aria-selected="true">Other A</button>
        <button role="tab" aria-selected="false">Other B</button>
      </div>
    </div>
  `
}

describe('installSharedTabKeyboard', () => {
  let cleanup: () => void

  beforeEach(() => {
    mountSharedTabs()
    cleanup = installSharedTabKeyboard()
  })

  afterEach(() => {
    cleanup()
    document.body.innerHTML = ''
  })

  it('keeps one Shared tab stop aligned with aria-selected', () => {
    const tabs = document.querySelectorAll<HTMLButtonElement>('.shared-page [role="tab"]')
    expect(tabs.item(0).tabIndex).toBe(0)
    expect(tabs.item(1).tabIndex).toBe(-1)

    tabs.item(0).setAttribute('aria-selected', 'false')
    tabs.item(1).setAttribute('aria-selected', 'true')
    return flush().then(() => {
      expect(tabs.item(0).tabIndex).toBe(-1)
      expect(tabs.item(1).tabIndex).toBe(0)
    })
  })

  it('wraps arrows and activates through the existing click behavior', async () => {
    const tabs = document.querySelectorAll<HTMLButtonElement>('.shared-page [role="tab"]')
    tabs.item(0).addEventListener('click', () => {
      tabs.item(0).setAttribute('aria-selected', 'true')
      tabs.item(1).setAttribute('aria-selected', 'false')
    })
    tabs.item(1).addEventListener('click', () => {
      tabs.item(0).setAttribute('aria-selected', 'false')
      tabs.item(1).setAttribute('aria-selected', 'true')
    })

    tabs.item(0).focus()
    tabs.item(0).dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowLeft', bubbles: true, cancelable: true }))
    await flush()
    expect(document.activeElement).toBe(tabs.item(1))
    expect(tabs.item(1).getAttribute('aria-selected')).toBe('true')

    tabs.item(1).dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true, cancelable: true }))
    await flush()
    expect(document.activeElement).toBe(tabs.item(0))
  })

  it('supports Home and End', async () => {
    const tabs = document.querySelectorAll<HTMLButtonElement>('.shared-page [role="tab"]')
    tabs.item(0).focus()
    tabs.item(0).dispatchEvent(new KeyboardEvent('keydown', { key: 'End', bubbles: true, cancelable: true }))
    await flush()
    expect(document.activeElement).toBe(tabs.item(1))

    tabs.item(1).dispatchEvent(new KeyboardEvent('keydown', { key: 'Home', bubbles: true, cancelable: true }))
    await flush()
    expect(document.activeElement).toBe(tabs.item(0))
  })

  it('does not alter tablists outside Shared', () => {
    const otherTabs = document.querySelectorAll<HTMLButtonElement>('.other-page [role="tab"]')
    expect(otherTabs.item(0).tabIndex).toBe(0)
    expect(otherTabs.item(1).tabIndex).toBe(0)
  })
})
