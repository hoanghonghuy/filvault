import type { ApiErrorCopy } from '@/api/errors'
import type { Locale } from '@/lib/i18n'

export type MediaPickerCopy = {
  apiError: ApiErrorCopy
  title: string
  hint: string
  loadFailed: string
  loadMoreFailed: string
  emptyTitle: string
  emptyDescription: string
}

const copy: Record<Locale, MediaPickerCopy> = {
  vi: {
    apiError: {
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
    },
    title: 'Thêm vào album',
    hint: 'Chạm vào ảnh hoặc video để thêm.',
    loadFailed: 'Không thể tải ảnh và video',
    loadMoreFailed: 'Không thể tải thêm',
    emptyTitle: 'Không có ảnh hoặc video khả dụng',
    emptyDescription: 'Hãy tải ảnh hoặc video lên Tệp của tôi trước.',
  },
  en: {
    apiError: {},
    title: 'Add to album',
    hint: 'Tap a photo or video to add it.',
    loadFailed: 'Failed to load photos',
    loadMoreFailed: 'Failed to load more',
    emptyTitle: 'No media available',
    emptyDescription: 'Upload photos or videos in My Files first.',
  },
}

export function mediaPickerCopy(locale: Locale): MediaPickerCopy {
  return copy[locale]
}
