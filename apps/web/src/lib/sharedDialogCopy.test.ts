import { describe, expect, it } from 'vitest'
import { sharedDialogCopy } from '@/lib/sharedDialogCopy'

describe('sharedDialogCopy', () => {
  it('provides Vietnamese shared dialog defaults', () => {
    expect(sharedDialogCopy('vi')).toEqual({
      confirm: 'Xác nhận',
      save: 'Lưu',
      cancel: 'Hủy',
    })
  })

  it('provides English shared dialog defaults', () => {
    expect(sharedDialogCopy('en')).toEqual({
      confirm: 'Confirm',
      save: 'Save',
      cancel: 'Cancel',
    })
  })
})
