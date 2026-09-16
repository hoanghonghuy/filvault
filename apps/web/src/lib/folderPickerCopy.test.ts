import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { folderPickerCopy } from './folderPickerCopy'

const viApiError = {
  network: 'Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.',
  codes: {
    UNAUTHORIZED: 'Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.',
    FORBIDDEN: 'Bạn không có quyền thực hiện thao tác này.',
    NOT_FOUND: 'Không tìm thấy mục này hoặc mục đã bị thay đổi.',
    CONFLICT: 'Tên đã tồn tại ở vị trí này.',
    RATE_LIMITED: 'Bạn thao tác quá nhanh. Vui lòng thử lại sau.',
    QUOTA_EXCEEDED: 'Dung lượng lưu trữ đã đầy. Hãy giải phóng dung lượng rồi thử lại.',
    FILE_TOO_LARGE: 'Tệp vượt quá giới hạn 100 MiB.',
    VALIDATION_ERROR: 'Dữ liệu không hợp lệ. Hãy kiểm tra loại tệp rồi thử lại.',
    INVALID_STATE: 'Không thể thực hiện thao tác này lúc này.',
    UPLOAD_EXPIRED: 'Phiên tải lên đã hết hạn. Hãy tải lại tệp.',
  },
}

describe('folderPickerCopy', () => {
  it('provides Vietnamese folder navigation copy', () => {
    expect(folderPickerCopy('vi')).toEqual({
      apiError: viApiError,
      browseFolders: 'Duyệt thư mục',
      root: 'Gốc',
      parentFolder: '.. Thư mục cha',
      loadFailed: 'Không thể tải thư mục',
      empty: 'Không có thư mục con.',
    })
  })

  it('provides English folder navigation copy', () => {
    expect(folderPickerCopy('en')).toEqual({
      apiError: {},
      browseFolders: 'Browse folders',
      root: 'Root',
      parentFolder: '.. Parent folder',
      loadFailed: 'Failed to load folders',
      empty: 'No subfolders here.',
    })
  })

  it('localizes network failures for Vietnamese Folder Picker sessions', () => {
    const copy = folderPickerCopy('vi')

    expect(formatApiError(new TypeError('Failed to fetch'), copy.loadFailed, copy.apiError)).toBe(
      viApiError.network,
    )
  })

  it('localizes common mapped API codes for Vietnamese Folder Picker sessions', () => {
    const copy = folderPickerCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.loadFailed, copy.apiError)).toBe(viApiError.codes.NOT_FOUND)
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
