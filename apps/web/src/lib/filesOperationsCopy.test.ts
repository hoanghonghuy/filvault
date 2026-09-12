import { describe, expect, it } from 'vitest'
import { filesOperationsCopy } from './filesOperationsCopy'

describe('filesOperationsCopy', () => {
  it('localizes rename and move feedback', () => {
    const en = filesOperationsCopy('en')
    const vi = filesOperationsCopy('vi')

    expect(en.renameFileTitle).toBe('Rename file')
    expect(vi.renameFileTitle).toBe('Đổi tên tệp')
    expect(en.moveBatchTitle(2)).toBe('Move 2 items')
    expect(vi.moveBatchTitle(2)).toBe('Di chuyển 2 mục')
  })

  it('keeps English count grammar correct for one and many items', () => {
    const copy = filesOperationsCopy('en')
    expect(copy.moveBatchTitle(1)).toBe('Move 1 item')
    expect(copy.movedBatch(1)).toBe('Moved 1 item')
    expect(copy.trashBatchTitle(1)).toBe('Move 1 item to trash?')
    expect(copy.movedBatchToTrash(2)).toBe('Moved 2 items to trash')
  })

  it('keeps destructive trash recovery guidance explicit in both locales', () => {
    expect(filesOperationsCopy('en').trashFolderMessage).toContain('restore')
    expect(filesOperationsCopy('vi').trashFolderMessage).toContain('khôi phục')
    expect(filesOperationsCopy('en').trashBatchMessage).toContain('Trash')
    expect(filesOperationsCopy('vi').trashBatchMessage).toContain('Thùng rác')
  })
})
