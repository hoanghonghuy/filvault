import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { chatApiErrorCopy } from '@/lib/chatApiErrorCopy'

describe('chatApiErrorCopy', () => {
  it('provides Vietnamese network and common API error copy', () => {
    const copy = chatApiErrorCopy('vi')

    expect(copy.apiError.network).toContain('Không thể kết nối')
    expect(copy.apiError.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(copy.apiError.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('preserves English shared formatter defaults', () => {
    expect(chatApiErrorCopy('en').apiError).toEqual({})
  })

  it('formats network failures with Vietnamese copy', () => {
    const { apiError } = chatApiErrorCopy('vi')

    expect(formatApiError(new TypeError('Failed to fetch'), 'Không thể tải tin nhắn', apiError)).toBe(
      apiError.network,
    )
  })

  it('formats mapped API codes with Vietnamese copy', () => {
    const { apiError } = chatApiErrorCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, 'Không thể tải tin nhắn', apiError)).toBe(
      'Không tìm thấy nội dung này hoặc nội dung đã bị xóa.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const { apiError } = chatApiErrorCopy('vi')

    expect(formatApiError(new Error('unexpected json'), 'Không thể gửi tin nhắn', apiError)).toBe(
      'Không thể gửi tin nhắn',
    )
  })
})
