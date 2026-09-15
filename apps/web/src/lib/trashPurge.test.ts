import { describe, expect, it, vi } from 'vitest'
import { runTrashPurge, type TrashPurgeTarget } from './trashPurge'

const targets: TrashPurgeTarget[] = [
  { type: 'folders', id: 'folder-1' },
  { type: 'files', id: 'file-1' },
  { type: 'files', id: 'file-2' },
]

describe('runTrashPurge', () => {
  it('continues after an individual delete failure and reports partial truth', async () => {
    const deleteTarget = vi.fn<(target: TrashPurgeTarget) => Promise<void>>(async (target) => {
      if (target.id === 'file-1') throw new Error('network')
    })
    const onProgress = vi.fn<(completed: number, total: number) => void>()

    const result = await runTrashPurge(targets, deleteTarget, onProgress)

    expect(deleteTarget).toHaveBeenCalledTimes(3)
    expect(deleteTarget).toHaveBeenNthCalledWith(3, targets[2])
    expect(result).toEqual({ total: 3, deleted: 2, failed: 1 })
    expect(onProgress).toHaveBeenLastCalledWith(3, 3)
  })

  it('reports full success without fabricating failures', async () => {
    const deleteTarget = vi.fn<(target: TrashPurgeTarget) => Promise<void>>().mockResolvedValue(undefined)

    await expect(runTrashPurge(targets, deleteTarget)).resolves.toEqual({
      total: 3,
      deleted: 3,
      failed: 0,
    })
  })
})
