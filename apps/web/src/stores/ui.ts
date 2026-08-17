import { ref } from 'vue'
import { defineStore } from 'pinia'

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
  const toastMessage = ref<string | null>(null)
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
  let actionSheetResolve: ((id: string | null) => void) | null = null

  function showToast(message: string, durationMs = 3000) {
    toastMessage.value = message
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      toastMessage.value = null
      toastTimer = null
    }, durationMs)
  }

  function confirm(options: ConfirmOptions): Promise<boolean> {
    return new Promise((resolve) => {
      confirmResolve = resolve
      confirmState.value = {
        open: true,
        confirmLabel: 'Confirm',
        cancelLabel: 'Cancel',
        ...options,
      }
    })
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
    return new Promise((resolve) => {
      actionSheetResolve = resolve
      actionSheetState.value = { open: true, title, items }
    })
  }

  function resolveActionSheet(id: string | null) {
    actionSheetState.value = { ...actionSheetState.value, open: false }
    actionSheetResolve?.(id)
    actionSheetResolve = null
  }

  return {
    toastMessage,
    confirmState,
    promptState,
    actionSheetState,
    showToast,
    confirm,
    resolveConfirm,
    prompt,
    resolvePrompt,
    openActionSheet,
    resolveActionSheet,
  }
})
