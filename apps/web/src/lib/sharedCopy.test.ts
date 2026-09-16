import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { formatShareExpiry, formatSharedDate, sharedCopy } from './sharedCopy'

describe('sharedCopy', () => {
  it('keeps destructive and action copy locale-consistent', () => {
    const vi = sharedCopy('vi')
    const en = sharedCopy('en')

    expect(vi.revokeTitle).toBe('Thu hồi liên kết?')
    expect(vi.revokeMessage('report.pdf')).toContain('report.pdf')
    expect(vi.revokeConfirm).toBe('Thu hồi liên kết')
    expect(vi.copyLink).toBe('Sao chép liên kết')

    expect(en.revokeTitle).toBe('Revoke link?')
    expect(en.revokeMessage('report.pdf')).toContain('report.pdf')
    expect(en.revokeConfirm).toBe('Revoke link')
    expect(en.copyLink).toBe('Copy link')
  })

  it('localizes load, failure, success and accessibility copy', () => {
    const vi = sharedCopy('vi')
    const en = sharedCopy('en')

    expect(vi.loadIncomingFailed).toContain('Không thể')
    expect(vi.linkCopied).toBe('Đã sao chép liên kết')
    expect(vi.sharedFolderContentsAria).toBe('Nội dung thư mục được chia sẻ')

    expect(en.loadIncomingFailed).toBe('Could not load shared items')
    expect(en.linkCopied).toBe('Link copied')
    expect(en.sharedFolderContentsAria).toBe('Shared folder contents')
  })

  it('formats dates and expiry labels for the active locale', () => {
    const future = '2030-04-15T12:00:00.000Z'
    const past = '2020-04-15T12:00:00.000Z'
    const now = new Date('2026-09-11T12:00:00.000Z').getTime()

    expect(formatSharedDate(future, 'vi')).not.toBe(formatSharedDate(future, 'en'))
    expect(formatShareExpiry(null, 'vi', now)).toBe('Có hiệu lực vĩnh viễn')
    expect(formatShareExpiry(null, 'en', now)).toBe('Active forever')
    expect(formatShareExpiry(past, 'vi', now)).toBe('Đã hết hạn')
    expect(formatShareExpiry(past, 'en', now)).toBe('Expired')
    expect(formatShareExpiry(future, 'vi', now)).toContain('Hết hạn')
    expect(formatShareExpiry(future, 'en', now)).toContain('Expires')
  })

  it('returns an empty label for invalid expiry dates instead of leaking invalid text', () => {
    expect(formatShareExpiry('not-a-date', 'vi')).toBe('')
    expect(formatShareExpiry('not-a-date', 'en')).toBe('')
  })

  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = sharedCopy('en').apiError
    const vietnamese = sharedCopy('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('Tên đã tồn tại')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
    expect(vietnamese.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('localizes network failures for Vietnamese Shared sessions', () => {
    const copy = sharedCopy('vi')

    expect(
      formatApiError(new TypeError('Failed to fetch'), copy.loadIncomingFailed, copy.apiError),
    ).toBe('Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.')
  })

  it('localizes common mapped API codes for Vietnamese Shared sessions', () => {
    const copy = sharedCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.loadIncomingFailed, copy.apiError)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = sharedCopy('vi')
    const fallback = 'Không thể tải các mục được chia sẻ'

    expect(formatApiError(new Error('unexpected json'), fallback, copy.apiError)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Shared copy omits overrides', () => {
    const copy = sharedCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), copy.loadIncomingFailed, copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), copy.loadIncomingFailed, copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
