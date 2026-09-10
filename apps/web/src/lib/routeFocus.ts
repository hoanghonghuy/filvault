type AfterEachHook = (to: unknown, from: unknown) => void

type RouterLike = {
  afterEach: (hook: AfterEachHook) => (() => void) | void
}

const MODAL_SELECTOR = '[aria-modal="true"]'
const PAGE_TITLE_SELECTOR = '.header-title'

function isVisible(element: HTMLElement): boolean {
  if (element.hidden) return false
  const style = window.getComputedStyle(element)
  return style.display !== 'none' && style.visibility !== 'hidden'
}

function hasOpenModal(): boolean {
  return Array.from(document.querySelectorAll<HTMLElement>(MODAL_SELECTOR)).some(isVisible)
}

function focusPageTitle() {
  if (hasOpenModal()) return

  const title = document.querySelector<HTMLElement>(PAGE_TITLE_SELECTOR)
  if (!title || !isVisible(title)) return

  const hadTabindex = title.hasAttribute('tabindex')
  const previousTabindex = title.getAttribute('tabindex')

  if (!hadTabindex) title.setAttribute('tabindex', '-1')
  title.focus({ preventScroll: true })

  if (!hadTabindex) {
    const cleanup = () => {
      if (previousTabindex === null) title.removeAttribute('tabindex')
      else title.setAttribute('tabindex', previousTabindex)
      title.removeEventListener('blur', cleanup)
    }
    title.addEventListener('blur', cleanup, { once: true })
  }
}

export function installRouteFocus(router: RouterLike): () => void {
  if (typeof document === 'undefined') return () => {}

  let initialized = false
  let frame = 0

  const removeHook = router.afterEach(() => {
    if (!initialized) {
      initialized = true
      return
    }

    if (frame) cancelAnimationFrame(frame)
    frame = requestAnimationFrame(() => {
      frame = 0
      focusPageTitle()
    })
  })

  return () => {
    if (frame) cancelAnimationFrame(frame)
    removeHook?.()
  }
}
