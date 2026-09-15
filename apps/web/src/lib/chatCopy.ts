import type { ApiErrorCopy } from '@/api/errors'
import type { Locale } from '@/lib/i18n'

export type ChatRuntimeCopy = {
  apiError: ApiErrorCopy
}

const COPY: Record<Locale, ChatRuntimeCopy> = {
  vi: {
    apiError: {
      network: 'Không thể kết nối đến máy chủ. Hãy kiểm tra kết nối mạng và thử lại.',
      codes: {
        UNAUTHORIZED: 'Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.',
        FORBIDDEN: 'Bạn không có quyền thực hiện thao tác này.',
        NOT_FOUND: 'Không tìm thấy nội dung này hoặc nội dung đã bị xóa.',
        CONFLICT: 'Nội dung đã thay đổi. Hãy tải lại và thử lại.',
        RATE_LIMITED: 'Bạn thao tác quá nhanh. Vui lòng thử lại sau.',
        VALIDATION_ERROR: 'Dữ liệu không hợp lệ. Vui lòng kiểm tra và thử lại.',
        INVALID_STATE: 'Không thể thực hiện thao tác này lúc này.',
        UPLOAD_EXPIRED: 'Phiên tải lên đã hết hạn. Vui lòng tải lại.',
        QUOTA_EXCEEDED: 'Dung lượng lưu trữ đã đầy. Hãy giải phóng dung lượng rồi thử lại.',
        FILE_TOO_LARGE: 'Tệp vượt quá giới hạn 100 MiB.',
      },
    },
  },
  en: {
    apiError: {},
  },
}

export function chatRuntimeCopy(locale: Locale): ChatRuntimeCopy {
  return COPY[locale]
}
