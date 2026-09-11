import { describe, expect, it } from 'vitest'
import { chatAttachmentCopy } from './chatAttachmentCopy'

describe('chatAttachmentCopy', () => {
  it('localizes inline image action labels', () => {
    expect(chatAttachmentCopy('vi').viewImageAria('photo.heic')).toBe('Xem photo.heic')
    expect(chatAttachmentCopy('en').viewImageAria('photo.heic')).toBe('View photo.heic')
  })
})
