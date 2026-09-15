import { describe, expect, it } from 'vitest'
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
})
