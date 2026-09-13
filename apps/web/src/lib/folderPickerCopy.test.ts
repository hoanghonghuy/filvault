import { describe, expect, it } from 'vitest'
import { folderPickerCopy } from './folderPickerCopy'

describe('folderPickerCopy', () => {
  it('provides Vietnamese folder navigation copy', () => {
    expect(folderPickerCopy('vi')).toEqual({
      browseFolders: 'Duyệt thư mục',
      root: 'Gốc',
      parentFolder: '.. Thư mục cha',
      loadFailed: 'Không thể tải thư mục',
      empty: 'Không có thư mục con.',
    })
  })

  it('provides English folder navigation copy', () => {
    expect(folderPickerCopy('en')).toEqual({
      browseFolders: 'Browse folders',
      root: 'Root',
      parentFolder: '.. Parent folder',
      loadFailed: 'Failed to load folders',
      empty: 'No subfolders here.',
    })
  })
})
