import { ref } from 'vue'
import { defineStore } from 'pinia'
import { useI18n } from '@/lib/i18n'
import {
  isLegacyLogoutConfirmation,
  logoutConfirmationOptions,
} from '@/lib/logoutConfirmation'

export type ConfirmOptions = {
  title: string
  message?: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

export type PromptOptions = {
  title: string
  label?: string
  initialValue?: string
  confirmLabel?: string
  cancelLabel?: string
}

export type ActionSheetItem = {
  id: string
  label: string
  icon?: string
  danger?: boolean
}

type ConfirmState = ConfirmOptions & { open: boolean }
type PromptState = PromptOptions & { open: boolean; value: string }
type ActionSheetState = {
  open: boolean
  title?: string
  items: ActionSheetItem[]
}

export const useUiStore = defineStore('ui', () => {
  const { locale } = useI18n()
  const toastMessage = ref<string | null>(null)
  const toastType = ref<'success' | 'error' | 'info'>('info')
  let toastTimer: ReturnType<typeof setTimeout> | null = null

  const confirmState = ref<ConfirmState>({
    open: false,
    title: '',
    confirmLabel: 'Confirm',
    cancelLabel: 'Cancel',
  })
  let confirmResolve: ((value: boolean) => void) | null = null

  const promptState = ref<PromptState>({
    open: false,
    title: '',
    label: '',
    value: '',
    confirmLabel: 'Save',
    cancelLabel: 'Cancel',
  })
  let promptResolve: ((value: string | null) => void) | null = null

  const actionSheetState = ref<ActionSheetState>({
    open: false,
    items: [],
  })
  // Resolver for the sheet currently on screen.
  let actionSheetActiveResolve: ((id: string | null) => void) | null = null
  // Selection is held here until the sheet's leave transition has finished,
  // so every caller resumes only when the screen is clear for the next dialog.
  let actionSheetClosing: { id: string | null; resolve: (value: string | null) => void } | null = null

  function showToast(message: string, type: 'success' | 'error' | 'info' = 'info', durationMs = 3000) {
    toastMessage.value = message
    toastType.value = type
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      toastMessage.value = null
      toastTimer = null
    }, durationMs)
  }

  function confirm(options: ConfirmOptions): Promise<boolean> {
    return new Promise((resolve) => {
      confirmResolve = resolve
      const normalizedOptions = isLegacyLogoutConfirmation(options)
        ? logoutConfirmationOptions(locale.value)
        : options
      confirmState.value = {
        open: true,
        confirmLabel: 'Confirm',
        cancelLabel: 'Cancel',
        ...normalizedOptions,
      }
    })
  }

  function confirmLogout(): Promise<boolean> {
    return confirm(logoutConfirmationOptions(locale.value))
  }

  function resolveConfirm(value: boolean) {
    confirmState.value = { ...confirmState.value, open: false }
    confirmResolve?.(value)
    confirmResolve = null
  }

  function prompt(options: PromptOptions): Promise<string | null> {
    return new Promise((resolve) => {
      promptResolve = resolve
      promptState.value = {
        open: true,
        title: options.title,
        label: options.label ?? '',
        value: options.initialValue ?? '',
        confirmLabel: options.confirmLabel ?? 'Save',
        cancelLabel: options.cancelLabel ?? 'Cancel',
      }
    })
  }

  function resolvePrompt(value: string | null) {
    promptState.value = { ...promptState.value, open: false }
    promptResolve?.(value)
    promptResolve = null
  }

  function openActionSheet(title: string | undefined, items: ActionSheetItem[]): Promise<string | null> {
    // Opening over a visible sheet swaps content in place: no leave transition
    // fires, so the replaced caller settles immediately with null.
    if (actionSheetActiveResolve !== null) {
      const replaced = actionSheetActiveResolve
      actionSheetActiveResolve = null
      replaced(null)
    }
    // Re-opening mid-leave cancels that leave (after-leave never fires),
    // so settle the closing caller here as well to avoid a hang.
    if (actionSheetClosing !== null) {
      const closing = actionSheetClosing
      actionSheetClosing = null
      closing.resolve(closing.id)
    }
    return new Promise((resolve) => {
      actionSheetActiveResolve = resolve
      actionSheetState.value = { open: true, title, items }
    })
  }

  function resolveActionSheet(id: string | null) {
    if (actionSheetActiveResolve === null) return
    const resolve = actionSheetActiveResolve
    actionSheetActiveResolve = null
    actionSheetState.value = { ...actionSheetState.value, open: false }
    // Hold the result until the leave transition reports completion.
    actionSheetClosing = { id: id ?? null, resolve }
  }

  function notifyActionSheetAfterLeave() {
    if (actionSheetClosing === null) return
    const closing = actionSheetClosing
    actionSheetClosing = null
    closing.resolve(closing.id)
  }

  return {
    toastMessage,
    toastType,
    confirmState,
    promptState,
    actionSheetState,
    showToast,
    confirm,
    confirmLogout,
    resolveConfirm,
    prompt,
    resolvePrompt,
    openActionSheet,
    resolveActionSheet,
    notifyActionSheetAfterLeave,
  }
})
