const SHARED_TABLIST_SELECTOR = '.shared-page [role="tablist"]'
const TAB_SELECTOR = '[role="tab"]'

function tabsOf(tablist: HTMLElement): HTMLElement[] {
  return Array.from(tablist.querySelectorAll<HTMLElement>(TAB_SELECTOR)).filter(
    (tab) => !tab.hasAttribute('disabled'),
  )
}

function syncTabStops(tablist: HTMLElement) {
  const tabs = tabsOf(tablist)
  if (tabs.length === 0) return
  const selected = tabs.find((tab) => tab.getAttribute('aria-selected') === 'true') ?? tabs[0]
  for (const tab of tabs) tab.tabIndex = tab === selected ? 0 : -1
}

function activate(tabs: HTMLElement[], index: number) {
  const tab = tabs[index]
  if (!tab) return
  tab.focus()
  tab.click()
}

export function installSharedTabKeyboard(): () => void {
  if (typeof document === 'undefined' || typeof MutationObserver === 'undefined') return () => {}

  const syncAll = () => {
    document.querySelectorAll<HTMLElement>(SHARED_TABLIST_SELECTOR).forEach(syncTabStops)
  }

  const onKeydown = (event: KeyboardEvent) => {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    const target = event.target
    if (!(target instanceof HTMLElement) || target.getAttribute('role') !== 'tab') return
    const tablist = target.closest<HTMLElement>(SHARED_TABLIST_SELECTOR)
    if (!tablist) return

    const tabs = tabsOf(tablist)
    const current = tabs.indexOf(target)
    if (current < 0 || tabs.length === 0) return

    let next = current
    if (event.key === 'ArrowRight') next = (current + 1) % tabs.length
    if (event.key === 'ArrowLeft') next = (current - 1 + tabs.length) % tabs.length
    if (event.key === 'Home') next = 0
    if (event.key === 'End') next = tabs.length - 1

    event.preventDefault()
    activate(tabs, next)
    queueMicrotask(() => syncTabStops(tablist))
  }

  const observer = new MutationObserver(syncAll)
  observer.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ['aria-selected', 'disabled'],
  })
  document.addEventListener('keydown', onKeydown)
  syncAll()

  return () => {
    observer.disconnect()
    document.removeEventListener('keydown', onKeydown)
  }
}
