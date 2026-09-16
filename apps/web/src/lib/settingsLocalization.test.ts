import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'

import {
  formatSettingsDate,
  formatSettingsRelativeTime,
  getSettingsText,
  settingsActivityLabel,
} from './settingsLocalization'

describe('settings localization', () => {
  it('localizes destructive share copy in Vietnamese and English', () => {
    expect(getSettingsText('vi').revokeConfirmTitle).toBe('Thu hồi liên kết?')
    expect(getSettingsText('vi').revokeConfirmMessage('photo.jpg')).toContain('photo.jpg')
    expect(getSettingsText('vi').revokeConfirmMessage('photo.jpg')).toContain('không còn được chia sẻ công khai')
    expect(getSettingsText('vi').revokeConfirmMessage()).toContain('Mục này')

    expect(getSettingsText('en').revokeConfirmTitle).toBe('Revoke link?')
    expect(getSettingsText('en').revokeConfirmMessage('photo.jpg')).toContain('will no longer be shared publicly')
    expect(getSettingsText('en').revokeConfirmMessage()).toContain('This item')
  })

  it('localizes activity labels', () => {
    expect(settingsActivityLabel('vi', 'file.uploaded')).toBe('Đã tải lên')
    expect(settingsActivityLabel('en', 'file.uploaded')).toBe('Uploaded')
  })

  it('preserves second-level precision in relative time', () => {
    const now = Date.parse('2026-09-11T03:00:00.000Z')
    const thirtySecondsAgo = '2026-09-11T02:59:30.000Z'

    expect(formatSettingsRelativeTime('en', thirtySecondsAgo, now)).toContain('30')
    expect(formatSettingsRelativeTime('en', thirtySecondsAgo, now).toLowerCase()).toContain('ago')
    expect(formatSettingsRelativeTime('vi', thirtySecondsAgo, now)).toContain('30')
    expect(formatSettingsRelativeTime('vi', thirtySecondsAgo, now).toLowerCase()).toContain('trước')
  })

  it('formats relative time with the active locale', () => {
    const now = Date.parse('2026-09-11T03:00:00.000Z')
    const fiveMinutesAgo = '2026-09-11T02:55:00.000Z'

    expect(formatSettingsRelativeTime('en', fiveMinutesAgo, now)).toContain('5')
    expect(formatSettingsRelativeTime('en', fiveMinutesAgo, now).toLowerCase()).toContain('ago')
    expect(formatSettingsRelativeTime('vi', fiveMinutesAgo, now)).toContain('5')
    expect(formatSettingsRelativeTime('vi', fiveMinutesAgo, now).toLowerCase()).toContain('trước')
  })

  it('formats older dates using the active locale', () => {
    const iso = '2026-01-02T10:00:00.000Z'
    expect(formatSettingsDate('en', iso)).not.toBe(formatSettingsDate('vi', iso))
  })

  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = getSettingsText('en').apiError
    const vietnamese = getSettingsText('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('Tên đã tồn tại')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
    expect(vietnamese.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('localizes network failures for Vietnamese Settings sessions', () => {
    const copy = getSettingsText('vi').apiError
    const fallback = getSettingsText('vi').revokeFailed

    expect(formatApiError(new TypeError('Failed to fetch'), fallback, copy)).toBe(
      'Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.',
    )
  })

  it('localizes common mapped API codes for Vietnamese Settings sessions', () => {
    const copy = getSettingsText('vi').apiError
    const fallback = getSettingsText('vi').revokeFailed
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, fallback, copy)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = getSettingsText('vi').apiError
    const fallback = getSettingsText('vi').revokeFailed

    expect(formatApiError(new Error('unexpected json'), fallback, copy)).toBe(fallback)
  })

  it('keeps English shared formatter defaults when Settings copy omits overrides', () => {
    const copy = getSettingsText('en').apiError
    const fallback = getSettingsText('en').revokeFailed

    expect(formatApiError(new TypeError('Failed to fetch'), fallback, copy)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), fallback, copy)).toBe(
      'Item not found.',
    )
  })
})
