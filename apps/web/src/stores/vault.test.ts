/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useVaultStore } from '@/stores/vault'
import type { VaultFile } from '@/api/types'

const apiMock = vi.hoisted(() => vi.fn<(path: string, options?: unknown) => Promise<unknown>>())

vi.mock('@/api/client', () => ({
  api: apiMock,
}))

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function secretFile(id: string): VaultFile {
  return {
    id,
    name: `${id}.txt`,
    mimeType: 'text/plain',
    sizeBytes: 100,
    createdAt: '',
    updatedAt: '',
  }
}

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
    vault.files = [secretFile('1')]
    vault.status = { initialized: true, unlocked: true }

    expect(vault.isUnlocked).toBe(true)

    vault.lock()

    expect(vault.vaultToken).toBeNull()
    expect(vault.files).toEqual([])
    expect(vault.isUnlocked).toBe(false)
  })

  it('loads files for the current unlocked Vault session', async () => {
    apiMock.mockResolvedValue({ files: [secretFile('current-secret')] })
    const vault = useVaultStore()
    vault.vaultToken = 'current-token'
    vault.status = { initialized: true, unlocked: true }

    await expect(vault.loadFiles()).resolves.toEqual([secretFile('current-secret')])

    expect(vault.files).toEqual([secretFile('current-secret')])
    expect(vault.loading).toBe(false)
    expect(vault.error).toBe('')
  })

  it('keeps lock authoritative when an older file hydration resolves late', async () => {
    const pendingFiles = deferred<{ files: VaultFile[] }>()
    apiMock.mockImplementation((path) => {
      if (path === '/vault/files') return pendingFiles.promise
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })
    const vault = useVaultStore()
    vault.vaultToken = 'old-token'
    vault.status = { initialized: true, unlocked: true }

    const hydration = vault.loadFiles()
    expect(vault.loading).toBe(true)

    vault.lock()
    pendingFiles.resolve({ files: [secretFile('stale-secret')] })
    await hydration

    expect(vault.vaultToken).toBeNull()
    expect(vault.files).toEqual([])
    expect(vault.loading).toBe(false)
    expect(vault.error).toBe('')
    expect(vault.isUnlocked).toBe(false)
  })

  it('keeps a newer Vault session hydration authoritative over an older failure and finally', async () => {
    const oldFiles = deferred<{ files: VaultFile[] }>()
    const newFiles = deferred<{ files: VaultFile[] }>()
    apiMock.mockImplementation((path, options) => {
      if (path === '/vault/files') {
        const token = (options as { headers?: Record<string, string> } | undefined)?.headers?.['X-Vault-Token']
        return token === 'old-token' ? oldFiles.promise : newFiles.promise
      }
      if (path === '/vault/unlock') return Promise.resolve({ token: 'new-token' })
      return Promise.reject(new Error(`Unexpected path: ${path}`))
    })
    const vault = useVaultStore()
    vault.vaultToken = 'old-token'
    vault.status = { initialized: true, unlocked: true }

    const oldHydration = vault.loadFiles()
    await vault.unlock('1234')
    const newHydration = vault.loadFiles()

    oldFiles.reject(new Error('stale old-session failure'))
    await expect(oldHydration).resolves.toEqual([])

    expect(vault.vaultToken).toBe('new-token')
    expect(vault.files).toEqual([])
    expect(vault.loading).toBe(true)
    expect(vault.error).toBe('')

    newFiles.resolve({ files: [secretFile('new-secret')] })
    await newHydration

    expect(vault.files).toEqual([secretFile('new-secret')])
    expect(vault.loading).toBe(false)
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