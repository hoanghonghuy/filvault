type AfterEachHook = (to: unknown, from: unknown) => void

type RouterLike = {
  afterEach: (hook: AfterEachHook) => (() => void) | void
  isReady: () => Promise<unknown>
}

const MODAL_SELECTOR = '[aria-modal="true"]'
const PAGE_TITLE_SELECTOR = '.header-title'

function isVisible(element: HTMLElement): boolean {
  let current: HTMLElement | null = element
  while (current) {
    if (current.hidden) return false
    const style = window.getComputedStyle(current)
    if (style.display === 'none' || style.visibility === 'hidden') return false
    current = current.parentElement
  }
  return true
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

  let disposed = false
  let frame = 0
  let removeHook: (() => void) | void

  void router.isReady().then(() => {
    if (disposed) return
    removeHook = router.afterEach(() => {
      if (frame) cancelAnimationFrame(frame)
      frame = requestAnimationFrame(() => {
        frame = 0
        focusPageTitle()
      })
    })
  })

  return () => {
    disposed = true
    if (frame) cancelAnimationFrame(frame)
    removeHook?.()
  }
}
