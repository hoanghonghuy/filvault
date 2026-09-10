const CALL_MODAL_SELECTOR = '.call-overlay[aria-modal="true"]'
const FOCUSABLE_SELECTOR = [
  'button:not([disabled])',
  'a[href]',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

function getFocusable(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter((element) => {
    const style = window.getComputedStyle(element)
    return style.visibility !== 'hidden' && style.display !== 'none'
  })
}

function focusOverlayFallback(overlay: HTMLElement) {
  if (!overlay.hasAttribute('tabindex')) overlay.setAttribute('tabindex', '-1')
  overlay.focus()
}

function focusInitialControl(overlay: HTMLElement) {
  const preferred =
    overlay.querySelector<HTMLElement>('.btn-accept:not([disabled])') ??
    overlay.querySelector<HTMLElement>('.header-action-btn:not([disabled])') ??
    overlay.querySelector<HTMLElement>('.control-btn:not(.btn-hangup):not([disabled])')

  if (preferred) {
    preferred.focus()
    return
  }

  // Outgoing/ended states may expose only a destructive action (cancel/hang up).
  // Focus the dialog container instead so opening the modal never preselects a
  // destructive control.
  focusOverlayFallback(overlay)
}

export function installCallModalFocusGuard(): () => void {
  if (typeof document === 'undefined' || typeof MutationObserver === 'undefined') return () => {}

  let activeOverlay: HTMLElement | null = null
  let previousFocus: HTMLElement | null = null

  const activate = (overlay: HTMLElement) => {
    if (activeOverlay === overlay) {
      if (!overlay.contains(document.activeElement)) focusInitialControl(overlay)
      return
    }

    activeOverlay = overlay
    previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
    queueMicrotask(() => {
      if (activeOverlay === overlay && overlay.isConnected) focusInitialControl(overlay)
    })
  }

  const deactivate = () => {
    const restoreTarget = previousFocus
    activeOverlay = null
    previousFocus = null
    if (restoreTarget?.isConnected) queueMicrotask(() => restoreTarget.focus())
  }

  const syncOverlay = () => {
    const overlay = document.querySelector<HTMLElement>(CALL_MODAL_SELECTOR)
    if (overlay) activate(overlay)
    else if (activeOverlay) deactivate()
  }

  const onKeydown = (event: KeyboardEvent) => {
    const overlay = activeOverlay
    if (!overlay || !overlay.isConnected) return

    if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      event.stopImmediatePropagation()
      return
    }

    if (event.key !== 'Tab') return

    const focusable = getFocusable(overlay)
    if (focusable.length === 0) {
      event.preventDefault()
      focusOverlayFallback(overlay)
      return
    }

    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    const current = document.activeElement

    if (!overlay.contains(current)) {
      event.preventDefault()
      ;(event.shiftKey ? last : first).focus()
      return
    }

    if (event.shiftKey && (current === first || current === overlay)) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && current === last) {
      event.preventDefault()
      first.focus()
    }
  }

  const observer = new MutationObserver(syncOverlay)
  observer.observe(document.body, { childList: true, subtree: true })
  document.addEventListener('keydown', onKeydown, true)
  syncOverlay()

  return () => {
    observer.disconnect()
    document.removeEventListener('keydown', onKeydown, true)
    if (activeOverlay) deactivate()
  }
}
