import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { profileAvatarCopy } from '@/lib/profileAvatarCopy'

describe('profileAvatarCopy', () => {
  it('provides Vietnamese avatar actions, feedback and API error copy', () => {
    const copy = profileAvatarCopy('vi')

    expect(copy.changeAvatar).toBe('Đổi ảnh đại diện')
    expect(copy.removeAvatar).toBe('Xóa ảnh đại diện')
    expect(copy.uploadSuccess).toBe('Đã cập nhật ảnh đại diện')
    expect(copy.removeSuccess).toBe('Đã xóa ảnh đại diện')
    expect(copy.saveFailed).toBe('Không thể lưu hồ sơ')
    expect(copy.apiError.network).toContain('Không thể kết nối')
    expect(copy.tooLarge).toContain('25MB')
  })

  it('keeps English shared API defaults', () => {
    const copy = profileAvatarCopy('en')

    expect(copy.changeAvatar).toBe('Change profile picture')
    expect(copy.removeAvatar).toBe('Remove profile picture')
    expect(copy.uploadFailed).toBe('Could not upload the image')
    expect(copy.removeFailed).toBe('Could not remove the profile picture')
    expect(copy.saveFailed).toBe('Save failed')
    expect(copy.apiError).toEqual({})
    expect(copy.invalidType).toContain('JPG, PNG, WebP, HEIC')
    expect(copy.heicDecodeFailed).toContain('HEIC')
    expect(copy.tooLarge).toContain('25MB')
  })

  it('formats common API codes with Vietnamese copy', () => {
    const copy = profileAvatarCopy('vi')
    const error = new ApiError('FORBIDDEN', 'Forbidden', 403)

    expect(formatApiError(error, copy.saveFailed, copy.apiError)).toBe(
      'Bạn không có quyền thực hiện thao tác này.',
    )
  })
})
