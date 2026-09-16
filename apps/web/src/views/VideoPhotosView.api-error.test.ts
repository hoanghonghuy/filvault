import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { photosRuntimeCopy } from '@/lib/photosCopy'

const videoErrorCopy = {
  vi: {
    load: 'Không thể tải video',
    loadMore: 'Không thể tải thêm video',
    view: 'Không thể mở video',
    download: 'Không thể tải video xuống',
    favorite: 'Không thể cập nhật mục yêu thích',
  },
  en: {
    load: 'Failed to load videos',
    loadMore: 'Failed to load more videos',
    view: 'View failed',
    download: 'Download failed',
    favorite: 'Failed to update favorite',
  },
} as const

describe('VideoPhotosView API error localization', () => {
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

  it('localizes network failures for Vietnamese Video Photos sessions', () => {
    const apiError = photosRuntimeCopy('vi').apiError
    const fallback = videoErrorCopy.vi.load

    expect(formatApiError(new TypeError('Failed to fetch'), fallback, apiError)).toBe(
      'Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.',
    )
  })

  it('localizes common mapped API codes for Vietnamese Video Photos sessions', () => {
    const apiError = photosRuntimeCopy('vi').apiError
    const fallback = videoErrorCopy.vi.view
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, fallback, apiError)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const apiError = photosRuntimeCopy('vi').apiError
    const fallback = videoErrorCopy.vi.download

    expect(formatApiError(new Error('unexpected json'), fallback, apiError)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Video Photos copy omits overrides', () => {
    const apiError = photosRuntimeCopy('en').apiError
    const fallback = videoErrorCopy.en.load

    expect(formatApiError(new TypeError('Failed to fetch'), fallback, apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), fallback, apiError)).toBe(
      'Item not found.',
    )
  })
})
