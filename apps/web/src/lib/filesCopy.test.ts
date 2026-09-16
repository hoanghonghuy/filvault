import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { filesRuntimeCopy } from './filesCopy'

describe('filesRuntimeCopy', () => {
  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = filesRuntimeCopy('en').apiError
    const vietnamese = filesRuntimeCopy('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('Tên đã tồn tại')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
    expect(vietnamese.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('localizes network failures for Vietnamese Files sessions', () => {
    const copy = filesRuntimeCopy('vi')

    expect(
      formatApiError(new TypeError('Failed to fetch'), copy.apiError.network ?? '', copy.apiError),
    ).toBe('Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.')
  })

  it('localizes common mapped API codes for Vietnamese Files sessions', () => {
    const copy = filesRuntimeCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.apiError.network ?? '', copy.apiError)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = filesRuntimeCopy('vi')
    const fallback = 'Không thể tải tệp'

    expect(formatApiError(new Error('unexpected json'), fallback, copy.apiError)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Files copy omits overrides', () => {
    const copy = filesRuntimeCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), 'Failed to load files', copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), 'Failed to load files', copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
