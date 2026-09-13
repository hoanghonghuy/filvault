export interface TrashPurgeTarget {
  type: 'files' | 'folders'
  id: string
}

export interface TrashPurgeResult {
  total: number
  deleted: number
  failed: number
}

export async function runTrashPurge(
  targets: readonly TrashPurgeTarget[],
  deleteTarget: (target: TrashPurgeTarget) => Promise<unknown>,
  onProgress?: (completed: number, total: number) => void,
): Promise<TrashPurgeResult> {
  let deleted = 0
  let failed = 0

  for (const target of targets) {
    try {
      await deleteTarget(target)
      deleted += 1
    } catch {
      failed += 1
    } finally {
      onProgress?.(deleted + failed, targets.length)
    }
  }

  return {
    total: targets.length,
    deleted,
    failed,
  }
}
