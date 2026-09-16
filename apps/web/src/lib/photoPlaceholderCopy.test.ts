import { describe, expect, it } from 'vitest'
import { photoPlaceholderCopy } from './photoPlaceholderCopy'

describe('photoPlaceholderCopy', () => {
  it('provides Vietnamese photo copy', () => {
    expect(photoPlaceholderCopy('vi')).toEqual({ photo: 'Ảnh' })
  })

  it('provides English photo copy', () => {
    expect(photoPlaceholderCopy('en')).toEqual({ photo: 'Photo' })
  })
})
