import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { mediaPickerCopy } from './mediaPickerCopy'

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

describe('mediaPickerCopy', () => {
  it('provides the full Vietnamese album media-picker contract', () => {
    expect(mediaPickerCopy('vi')).toEqual({
      apiError: viApiError,
      title: 'Thêm vào album',
      hint: 'Chạm vào ảnh hoặc video để thêm.',
      loadFailed: 'Không thể tải ảnh và video',
      loadMoreFailed: 'Không thể tải thêm',
      emptyTitle: 'Không có ảnh hoặc video khả dụng',
      emptyDescription: 'Hãy tải ảnh hoặc video lên Tệp của tôi trước.',
    })
  })

  it('provides the full English album media-picker contract', () => {
    expect(mediaPickerCopy('en')).toEqual({
      apiError: {},
      title: 'Add to album',
      hint: 'Tap a photo or video to add it.',
      loadFailed: 'Failed to load photos',
      loadMoreFailed: 'Failed to load more',
      emptyTitle: 'No media available',
      emptyDescription: 'Upload photos or videos in My Files first.',
    })
  })

  it('localizes network failures for Vietnamese Media Picker sessions', () => {
    const copy = mediaPickerCopy('vi')

    expect(formatApiError(new TypeError('Failed to fetch'), copy.loadFailed, copy.apiError)).toBe(
      viApiError.network,
    )
  })

  it('localizes common mapped API codes for Vietnamese Media Picker sessions', () => {
    const copy = mediaPickerCopy('vi')
    const error = new ApiError('NOT_FOUND', 'not found', 404)

    expect(formatApiError(error, copy.loadFailed, copy.apiError)).toBe(viApiError.codes.NOT_FOUND)
  })

  it('keeps localized per-action fallback for unmapped errors', () => {
    const copy = mediaPickerCopy('vi')

    expect(formatApiError(new Error('unexpected json'), copy.loadMoreFailed, copy.apiError)).toBe(
      copy.loadMoreFailed,
    )
  })

  it('keeps English shared formatter defaults when Media Picker copy omits overrides', () => {
    const copy = mediaPickerCopy('en')

    expect(formatApiError(new TypeError('Failed to fetch'), copy.loadFailed, copy.apiError)).toBe(
      "Can't reach the server. Try again in a moment.",
    )
    expect(formatApiError(new ApiError('NOT_FOUND', 'not found', 404), copy.loadFailed, copy.apiError)).toBe(
      'Item not found.',
    )
  })
})
