import { describe, expect, it } from 'vitest'
import { photoMediaSheetCopy } from './photoMediaSheetCopy'

describe('photoMediaSheetCopy', () => {
  it('provides Vietnamese view action copy', () => {
    expect(photoMediaSheetCopy('vi')).toEqual({
      view: 'Xem',
    })
  })

  it('provides English view action copy', () => {
    expect(photoMediaSheetCopy('en')).toEqual({
      view: 'View',
    })
  })
})
