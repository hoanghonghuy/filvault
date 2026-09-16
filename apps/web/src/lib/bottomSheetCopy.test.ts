import { describe, expect, it } from 'vitest'
import { bottomSheetCopy } from './bottomSheetCopy'

describe('bottomSheetCopy', () => {
  it('provides Vietnamese untitled dialog aria copy', () => {
    expect(bottomSheetCopy('vi')).toEqual({
      untitledDialogAria: 'Hộp thoại',
    })
  })

  it('provides English untitled dialog aria copy', () => {
    expect(bottomSheetCopy('en')).toEqual({
      untitledDialogAria: 'Dialog',
    })
  })
})
