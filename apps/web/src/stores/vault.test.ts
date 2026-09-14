/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useVaultStore } from '@/stores/vault'

const apiMock = vi.hoisted(() => vi.fn<(path: string, options?: unknown) => Promise<unknown>>())

vi.mock('@/api/client', () => ({
  api: apiMock,
}))

describe('useVaultStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    sessionStorage.clear()
    apiMock.mockReset()
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

  it.each([
    {
      name: 'setup',
      path: '/vault/setup',
      run: (vault: ReturnType<typeof useVaultStore>) => vault.setup('1234'),
    },
    {
      name: 'unlock',
      path: '/vault/unlock',
      run: (vault: ReturnType<typeof useVaultStore>) => vault.unlock('1234'),
    },
    {
      name: 'reset',
      path: '/vault/reset',
      run: (vault: ReturnType<typeof useVaultStore>) => vault.resetPin('account-password', '1234'),
    },
  ])('$name establishes an unlocked session without implicitly hydrating files', async ({ path, run }) => {
    apiMock.mockResolvedValue({ token: 'vault-token' })
    const vault = useVaultStore()

    await run(vault)

    expect(apiMock).toHaveBeenCalledTimes(1)
    expect(apiMock).toHaveBeenCalledWith(path, expect.any(Object))
    expect(apiMock).not.toHaveBeenCalledWith('/vault/files', expect.anything())
    expect(vault.vaultToken).toBe('vault-token')
    expect(vault.isUnlocked).toBe(true)
  })
})