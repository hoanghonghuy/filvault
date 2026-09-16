import type { ApiErrorCopy } from '@/api/errors'

export type OverviewLocale = 'vi' | 'en'

export type OverviewCopy = {
  apiError: ApiErrorCopy
  clearSearchAria: string
  featureCategoriesAria: string
  loadOverviewFailed: string
  loadFilesFailed: string
  loadPhotosFailed: string
}

const COPY: Record<OverviewLocale, OverviewCopy> = {
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
    clearSearchAria: 'Xóa tìm kiếm',
    featureCategoriesAria: 'Danh mục tính năng',
    loadOverviewFailed: 'Không thể tải trang tổng quan. Vui lòng thử lại sau.',
    loadFilesFailed: 'Không thể tải tệp',
    loadPhotosFailed: 'Không thể tải ảnh',
  },
  en: {
    apiError: {},
    clearSearchAria: 'Clear search',
    featureCategoriesAria: 'Feature categories',
    loadOverviewFailed: 'Could not load overview. Try again later.',
    loadFilesFailed: 'Could not load files',
    loadPhotosFailed: 'Could not load photos',
  },
}

export function overviewCopy(locale: OverviewLocale): OverviewCopy {
  return COPY[locale]
}
