import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView rename/move/trash localization wiring', () => {
  it('uses locale-reactive operation copy for prompts, picker titles and destructive feedback', () => {
    expect(source).toContain('filesOperationsCopy(locale.value)')
    expect(source).toContain('operationsCopy.value.renameFileTitle')
    expect(source).toContain('operationsCopy.value.moveBatchTitle(totalSelectedCount.value)')
    expect(source).toContain('operationsCopy.value.trashBatchTitle(count)')
    expect(source).toContain('operationsCopy.value.trashFolderMessage')
  })

  it('uses existing localized generic labels for file and folder action sheets', () => {
    expect(source).toContain('label: t.value.open')
    expect(source).toContain('label: t.value.rename')
    expect(source).toContain('label: t.value.move')
    expect(source).toContain('label: t.value.moveToTrash')
    expect(source).toContain('label: t.value.vaultMoveToVault')
  })

  it('does not retain the replaced English-only operation strings', () => {
    expect(source).not.toContain("title: 'Rename file'")
    expect(source).not.toContain("title: 'Rename folder'")
    expect(source).not.toContain("title: 'Move folder to trash?'")
    expect(source).not.toContain("message: 'You can restore this file from Trash later.'")
    expect(source).not.toContain("formatApiError(e, 'Move failed')")
    expect(source).not.toContain("formatApiError(e, 'Delete failed')")
  })
})
