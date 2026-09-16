import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { folderPickerCopy } from '@/lib/folderPickerCopy'

describe('FolderPickerSheet API error localization', () => {
  it('provides Vietnamese shared API error copy while preserving English defaults', () => {
    const english = folderPickerCopy('en').apiError
    const vietnamese = folderPickerCopy('vi').apiError

    expect(english).toEqual({})
    expect(vietnamese.network).toContain('kết nối')
    expect(vietnamese.codes?.UNAUTHORIZED).toContain('đăng nhập')
    expect(vietnamese.codes?.FORBIDDEN).toContain('quyền')
    expect(vietnamese.codes?.NOT_FOUND).toContain('Không tìm thấy')
    expect(vietnamese.codes?.CONFLICT).toContain('Tên đã tồn tại')
    expect(vietnamese.codes?.RATE_LIMITED).toContain('thử lại')
    expect(vietnamese.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
  })

  it('localizes network failures for Vietnamese Folder Picker sessions', () => {
    const copy = folderPickerCopy('vi')

    expect(formatApiError(new TypeError('Failed to fetch'), copy.loadFailed, copy.apiError)).toBe(
      'Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.',
    )
  })

  it('localizes common mapped API codes for Vietnamese Folder Picker sessions', () => {
    const copy = folderPickerCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.loadFailed, copy.apiError)).toBe(
      'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    )
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = folderPickerCopy('vi')

    expect(formatApiError(new Error('unexpected json'), copy.loadFailed, copy.apiError)).toBe(
      copy.loadFailed,
    )
  })

  it('keeps English shared formatter defaults when Folder Picker copy omits overrides', () => {
    const copy = folderPickerCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), copy.loadFailed, copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), copy.loadFailed, copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
