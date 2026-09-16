import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { chatRuntimeCopy } from './chatCopy'

describe('chatRuntimeCopy', () => {
  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = chatRuntimeCopy('en').apiError
    const vietnamese = chatRuntimeCopy('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('thay đổi')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
  })

  it('localizes network failures for Vietnamese Chat sessions', () => {
    const copy = chatRuntimeCopy('vi')

    expect(
      formatApiError(new TypeError('Failed to fetch'), copy.apiError.network ?? '', copy.apiError),
    ).toBe('Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.')
  })

  it('localizes common mapped API codes for Vietnamese Chat sessions', () => {
    const copy = chatRuntimeCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.apiError.network ?? '', copy.apiError)).toBe(
      'Không tìm thấy nội dung này hoặc nội dung đã bị xóa.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = chatRuntimeCopy('vi')
    const fallback = 'Không thể tải cuộc trò chuyện'

    expect(formatApiError(new Error('unexpected json'), fallback, copy.apiError)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Chat copy omits overrides', () => {
    const copy = chatRuntimeCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), 'Failed to load chats', copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), 'Failed to load chats', copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
