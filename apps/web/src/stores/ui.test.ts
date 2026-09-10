import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { setLocale } from '@/lib/i18n'
import { useUiStore } from '@/stores/ui'

describe('useUiStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setLocale('vi')
  })

  it('builds logout confirmation from the active locale without a reload', async () => {
    const ui = useUiStore()

    const viConfirmation = ui.confirmLogout()
    expect(ui.confirmState.title).toBe('Đăng xuất?')
    expect(ui.confirmState.message).toContain('đăng nhập lại')
    expect(ui.confirmState.confirmLabel).toBe('Đăng xuất')
    expect(ui.confirmState.cancelLabel).toBe('Hủy')
    expect(ui.confirmState.danger).toBe(true)
    ui.resolveConfirm(false)
    await expect(viConfirmation).resolves.toBe(false)

    setLocale('en')
    const enConfirmation = ui.confirmLogout()
    expect(ui.confirmState.title).toBe('Log out?')
    expect(ui.confirmState.message).toContain('sign in again')
    expect(ui.confirmState.confirmLabel).toBe('Log out')
    expect(ui.confirmState.cancelLabel).toBe('Cancel')
    ui.resolveConfirm(true)
    await expect(enConfirmation).resolves.toBe(true)
  })

  it('normalizes the historical Settings logout call to the canonical localized contract', async () => {
    setLocale('vi')
    const ui = useUiStore()
    const confirmation = ui.confirm({
      title: 'Log out?',
      message: 'You will need to sign in again to access your files.',
      confirmLabel: 'Log out',
    })

    expect(ui.confirmState.title).toBe('Đăng xuất?')
    expect(ui.confirmState.confirmLabel).toBe('Đăng xuất')
    expect(ui.confirmState.cancelLabel).toBe('Hủy')
    expect(ui.confirmState.danger).toBe(true)
    ui.resolveConfirm(false)
    await expect(confirmation).resolves.toBe(false)
  })

  it('settles the previous sheet with null when a new one opens over it', async () => {
    const ui = useUiStore()
    const first = ui.openActionSheet('first', [{ id: 'a', label: 'A' }])

    ui.openActionSheet('second', [{ id: 'b', label: 'B' }])
    // Content swaps in place: no leave transition runs, so settle at once.
    await expect(first).resolves.toBeNull()
    expect(ui.actionSheetState.title).toBe('second')
    expect(ui.actionSheetState.open).toBe(true)
  })

  it('settles the chosen action only after the sheet has left the screen', async () => {
    const ui = useUiStore()
    const chosen = ui.openActionSheet('t', [{ id: 'x', label: 'X' }])

    let settled = false
    void chosen.then(() => {
      settled = true
    })

    ui.resolveActionSheet('x')
    await new Promise((resolve) => setTimeout(resolve, 0))
    // Selection happened, but the leave transition is still running.
    expect(settled).toBe(false)

    ui.notifyActionSheetAfterLeave()
    await expect(chosen).resolves.toBe('x')
    expect(settled).toBe(true)
  })

  it('settles a closing sheet immediately when a new one opens mid-leave', async () => {
    const ui = useUiStore()
    const chosen = ui.openActionSheet('first', [{ id: 'a', label: 'A' }])
    ui.resolveActionSheet(null)

    // Re-open while the previous leave is still animating: after-leave will
    // never fire for the cancelled transition, so settle without hanging.
    const second = ui.openActionSheet('second', [{ id: 'b', label: 'B' }])
    await expect(chosen).resolves.toBeNull()
    expect(ui.actionSheetState.title).toBe('second')

    ui.resolveActionSheet('b')
    ui.notifyActionSheetAfterLeave()
    await expect(second).resolves.toBe('b')
  })

  it('keeps a superseded sheet pending until its leave finishes', async () => {
    const ui = useUiStore()
    const first = ui.openActionSheet('first', [{ id: 'a', label: 'A' }])
    const second = ui.openActionSheet('second', [{ id: 'b', label: 'B' }])
    await expect(first).resolves.toBeNull()

    let settled = false
    void second.then(() => {
      settled = true
    })
    ui.resolveActionSheet('b')
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(settled).toBe(false)

    ui.notifyActionSheetAfterLeave()
    await expect(second).resolves.toBe('b')
    expect(settled).toBe(true)
  })
})
