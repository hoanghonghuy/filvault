import { describe, expect, it } from 'vitest'
import { photoPlaceholderCopy } from './photoPlaceholderCopy'

describe('photoPlaceholderCopy', () => {
  it('provides Vietnamese media badge copy', () => {
    expect(photoPlaceholderCopy('vi')).toEqual({
      photo: 'Ảnh',
      video: 'Video',
    })
  })

  it('provides English media badge copy', () => {
    expect(photoPlaceholderCopy('en')).toEqual({
      photo: 'Photo',
      video: 'Video',
    })
  })
})
