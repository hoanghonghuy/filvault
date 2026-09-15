import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { VaultFile, VaultSession, VaultStatus } from '@/api/types'

const VAULT_TOKEN_KEY = 'filvault.vaultToken'

export const useVaultStore = defineStore('vault', () => {
  const status = ref<VaultStatus | null>(null)
  const vaultToken = ref<string | null>(
    typeof sessionStorage !== 'undefined' ? sessionStorage.getItem(VAULT_TOKEN_KEY) : null,
  )
  const files = ref<VaultFile[]>([])
  const loading = ref(false)
  const error = ref('')
  let fileHydrationSequence = 0
  let statusRequestSequence = 0
  let sessionMutationSequence = 0

  const isInitialized = computed(() => status.value?.initialized ?? false)
  const isUnlocked = computed(() => Boolean(vaultToken.value && status.value?.unlocked))

  function invalidateSessionRequests() {
    fileHydrationSequence += 1
    statusRequestSequence += 1
    sessionMutationSequence += 1
  }

  function setToken(token: string) {
    invalidateSessionRequests()
    vaultToken.value = token
    if (typeof sessionStorage !== 'undefined') {
      sessionStorage.setItem(VAULT_TOKEN_KEY, token)
    }
  }

  function clearToken() {
    invalidateSessionRequests()
    vaultToken.value = null
    loading.value = false
    error.value = ''
    if (typeof sessionStorage !== 'undefined') {
      sessionStorage.removeItem(VAULT_TOKEN_KEY)
    }
  }

  function getHeaders(): HeadersInit {
    const headers: Record<string, string> = {}
    if (vaultToken.value) {
      headers['X-Vault-Token'] = vaultToken.value
    }
    return headers
  }

  async function fetchStatus(): Promise<VaultStatus> {
    const requestToken = vaultToken.value
    const requestSequence = ++statusRequestSequence
    try {
      const res = await api<VaultStatus>('/vault/status', {
        headers: getHeaders(),
      })
      if (requestSequence !== statusRequestSequence || requestToken !== vaultToken.value) {
        return status.value ?? res
      }
      status.value = res
      if (!res.unlocked && vaultToken.value) {
        clearToken()
      }
      return res
    } catch (e) {
      if (requestSequence === statusRequestSequence && requestToken === vaultToken.value) {
        status.value = { initialized: false, unlocked: false }
      }
      throw e
    }
  }

  async function setup(pin: string): Promise<VaultSession> {
    const requestSequence = ++sessionMutationSequence
    loading.value = true
    error.value = ''
    try {
      const session = await api<VaultSession>('/vault/setup', {
        method: 'POST',
        body: JSON.stringify({ pin }),
      })
      if (requestSequence !== sessionMutationSequence) return session
      setToken(session.token)
      status.value = { initialized: true, unlocked: true }
      loading.value = false
      return session
    } finally {
      if (requestSequence === sessionMutationSequence) loading.value = false
    }
  }

  async function unlock(pin: string): Promise<VaultSession> {
    const requestSequence = ++sessionMutationSequence
    loading.value = true
    error.value = ''
    try {
      const session = await api<VaultSession>('/vault/unlock', {
        method: 'POST',
        body: JSON.stringify({ pin }),
      })
      if (requestSequence !== sessionMutationSequence) return session
      setToken(session.token)
      status.value = { initialized: true, unlocked: true }
      loading.value = false
      return session
    } finally {
      if (requestSequence === sessionMutationSequence) loading.value = false
    }
  }

  function lock(): void {
    clearToken()
    files.value = []
    if (status.value) {
      status.value.unlocked = false
    }
  }

  async function changePin(currentPin: string, newPin: string): Promise<void> {
    loading.value = true
    error.value = ''
    try {
      await api('/vault/change-pin', {
        method: 'POST',
        body: JSON.stringify({ currentPin, newPin }),
        headers: getHeaders(),
      })
    } finally {
      loading.value = false
    }
  }

  async function resetPin(accountPassword: string, newPin: string): Promise<VaultSession> {
    const requestSequence = ++sessionMutationSequence
    loading.value = true
    error.value = ''
    try {
      const session = await api<VaultSession>('/vault/reset', {
        method: 'POST',
        body: JSON.stringify({ accountPassword, newPin }),
      })
      if (requestSequence !== sessionMutationSequence) return session
      setToken(session.token)
      status.value = { initialized: true, unlocked: true }
      loading.value = false
      return session
    } finally {
      if (requestSequence === sessionMutationSequence) loading.value = false
    }
  }

  async function loadFiles(): Promise<VaultFile[]> {
    const requestToken = vaultToken.value
    if (!requestToken) {
      files.value = []
      return []
    }
    const requestSequence = ++fileHydrationSequence
    loading.value = true
    error.value = ''
    try {
      const res = await api<{ files: VaultFile[] }>('/vault/files', {
        headers: getHeaders(),
      })
      if (requestSequence !== fileHydrationSequence || requestToken !== vaultToken.value || !isUnlocked.value) {
        return files.value
      }
      files.value = res.files ?? []
      return files.value
    } catch (e) {
      if (requestSequence !== fileHydrationSequence || requestToken !== vaultToken.value || !isUnlocked.value) {
        return files.value
      }
      error.value = 'Failed to load vault files'
      throw e
    } finally {
      if (requestSequence === fileHydrationSequence && requestToken === vaultToken.value) {
        loading.value = false
      }
    }
  }

  async function moveToVault(fileIds: string[]): Promise<void> {
    if (fileIds.length === 0) return
    await api('/vault/items', {
      method: 'POST',
      body: JSON.stringify({ fileIds }),
      headers: getHeaders(),
    })
    if (isUnlocked.value) {
      await loadFiles()
    }
  }

  async function removeFromVault(fileIds: string[]): Promise<void> {
    if (fileIds.length === 0) return
    await api('/vault/items/remove', {
      method: 'POST',
      body: JSON.stringify({ fileIds }),
      headers: getHeaders(),
    })
    await loadFiles()
  }

  return {
    status,
    vaultToken,
    files,
    loading,
    error,
    isInitialized,
    isUnlocked,
    fetchStatus,
    setup,
    unlock,
    lock,
    changePin,
    resetPin,
    loadFiles,
    moveToVault,
    removeFromVault,
  }
})