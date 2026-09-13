import { describe, expect, it } from 'vitest'
import { profileAvatarCopy } from '@/lib/profileAvatarCopy'

describe('profileAvatarCopy', () => {
  it('provides Vietnamese avatar actions and feedback', () => {
    const copy = profileAvatarCopy('vi')

    expect(copy.changeAvatar).toBe('Đổi ảnh đại diện')
    expect(copy.removeAvatar).toBe('Xóa ảnh đại diện')
    expect(copy.uploadSuccess).toBe('Đã cập nhật ảnh đại diện')
    expect(copy.removeSuccess).toBe('Đã xóa ảnh đại diện')
    expect(copy.tooLarge).toContain('25MB')
  })

  it('provides English avatar actions, validation and feedback', () => {
    const copy = profileAvatarCopy('en')

    expect(copy.changeAvatar).toBe('Change profile picture')
    expect(copy.removeAvatar).toBe('Remove profile picture')
    expect(copy.uploadFailed).toBe('Could not upload the image')
    expect(copy.removeFailed).toBe('Could not remove the profile picture')
    expect(copy.invalidType).toContain('JPG, PNG, WebP, HEIC')
    expect(copy.heicDecodeFailed).toContain('HEIC')
    expect(copy.tooLarge).toContain('25MB')
  })
})
