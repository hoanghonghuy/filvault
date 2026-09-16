import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { photosRuntimeCopy } from './photosCopy'

describe('photosRuntimeCopy', () => {
  it('provides Vietnamese runtime, action and accessibility copy', () => {
    const copy = photosRuntimeCopy('vi')

    expect(copy.loadPhotosFailed).toBe('Không thể tải ảnh')
    expect(copy.openAlbum).toBe('Mở')
    expect(copy.deleteAlbum).toBe('Xóa album')
    expect(copy.addFavorite('ảnh.jpg')).toBe('Đã thêm "ảnh.jpg" vào mục yêu thích')
    expect(copy.removeFavorite('ảnh.jpg')).toBe('Đã bỏ "ảnh.jpg" khỏi mục yêu thích')
    expect(copy.pullToRefreshPhotos).toBe('Kéo để làm mới ảnh')
    expect(copy.photosViewsAria).toBe('Các chế độ xem ảnh')
  })

  it('preserves the existing English semantics', () => {
    const copy = photosRuntimeCopy('en')

    expect(copy.loadPhotosFailed).toBe('Failed to load photos')
    expect(copy.openAlbum).toBe('Open')
    expect(copy.renameAlbum).toBe('Rename')
    expect(copy.deleteAlbum).toBe('Delete album')
    expect(copy.addFavorite('photo.jpg')).toBe('Added "photo.jpg" to favorites')
    expect(copy.removeFavorite('photo.jpg')).toBe('Removed "photo.jpg" from favorites')
    expect(copy.pullToRefreshPhotos).toBe('Pull to refresh photos')
    expect(copy.photosViewsAria).toBe('Photos views')
  })

  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = photosRuntimeCopy('en').apiError
    const vietnamese = photosRuntimeCopy('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('Tên đã tồn tại')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
    expect(vietnamese.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('localizes network failures for Vietnamese Photos sessions', () => {
    const copy = photosRuntimeCopy('vi')

    expect(
      formatApiError(new TypeError('Failed to fetch'), copy.apiError.network ?? '', copy.apiError),
    ).toBe('Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.')
  })

  it('localizes common mapped API codes for Vietnamese Photos sessions', () => {
    const copy = photosRuntimeCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.apiError.network ?? '', copy.apiError)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = photosRuntimeCopy('vi')
    const fallback = 'Không thể tải ảnh'

    expect(formatApiError(new Error('unexpected json'), fallback, copy.apiError)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Photos copy omits overrides', () => {
    const copy = photosRuntimeCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), 'Failed to load photos', copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), 'Failed to load photos', copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
