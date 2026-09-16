import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { overviewCopy } from './overviewCopy'

describe('overviewCopy', () => {
  it('localizes Overview accessibility labels', () => {
    expect(overviewCopy('vi').clearSearchAria).toBe('Xóa tìm kiếm')
    expect(overviewCopy('vi').featureCategoriesAria).toBe('Danh mục tính năng')
    expect(overviewCopy('en').clearSearchAria).toBe('Clear search')
    expect(overviewCopy('en').featureCategoriesAria).toBe('Feature categories')
  })

  it('keeps Overview failure fallbacks locale-consistent', () => {
    const vi = overviewCopy('vi')
    const en = overviewCopy('en')

    expect(vi.loadOverviewFailed).toContain('Không thể')
    expect(vi.loadFilesFailed).toBe('Không thể tải tệp')
    expect(vi.loadPhotosFailed).toBe('Không thể tải ảnh')

    expect(en.loadOverviewFailed).toBe('Could not load overview. Try again later.')
    expect(en.loadFilesFailed).toBe('Could not load files')
    expect(en.loadPhotosFailed).toBe('Could not load photos')
  })

  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = overviewCopy('en').apiError
    const vietnamese = overviewCopy('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('Tên đã tồn tại')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
    expect(vietnamese.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('localizes network failures for Vietnamese Overview sessions', () => {
    const copy = overviewCopy('vi')

    expect(
      formatApiError(new TypeError('Failed to fetch'), copy.apiError.network ?? '', copy.apiError),
    ).toBe('Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.')
  })

  it('localizes common mapped API codes for Vietnamese Overview sessions', () => {
    const copy = overviewCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.apiError.network ?? '', copy.apiError)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = overviewCopy('vi')
    const fallback = 'Không thể tải tệp'

    expect(formatApiError(new Error('unexpected json'), fallback, copy.apiError)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Overview copy omits overrides', () => {
    const copy = overviewCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), 'Could not load files', copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), 'Could not load files', copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
