/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useVaultStore } from '@/stores/vault'

describe('useVaultStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    sessionStorage.clear()
    vi.restoreAllMocks()
  })

  it('starts with not initialized and locked', () => {
    const vault = useVaultStore()
    expect(vault.isInitialized).toBe(false)
    expect(vault.isUnlocked).toBe(false)
    expect(vault.vaultToken).toBeNull()
    expect(vault.files).toEqual([])
  })

  it('lock clears token and files', () => {
    const vault = useVaultStore()
    vault.vaultToken = 'test-token'
    vault.files = [{ id: '1', name: 'secret.txt', mimeType: 'text/plain', sizeBytes: 100, createdAt: '', updatedAt: '' }]
    vault.status = { initialized: true, unlocked: true }

    expect(vault.isUnlocked).toBe(true)

    vault.lock()

    expect(vault.vaultToken).toBeNull()
    expect(vault.files).toEqual([])
    expect(vault.isUnlocked).toBe(false)
  })
})
