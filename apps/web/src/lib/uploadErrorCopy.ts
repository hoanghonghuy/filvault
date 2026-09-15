import type { Locale } from '@/lib/i18n'
import type { ApiErrorCopy } from '@/api/errors'

export interface UploadErrorCopy {
  fallback: string
  api: ApiErrorCopy
}

const copy: Record<Locale, UploadErrorCopy> = {
  vi: {
    fallback: 'Không thể tải tệp lên',
    api: {
      network: 'Không thể kết nối đến máy chủ. Vui lòng thử lại sau.',
      codes: {
        CONFLICT: 'Tên đã tồn tại ở vị trí này.',
        QUOTA_EXCEEDED: 'Dung lượng lưu trữ đã đầy. Hãy giải phóng dung lượng rồi thử lại.',
        FILE_TOO_LARGE: 'Tệp vượt quá giới hạn 100 MiB.',
        VALIDATION_ERROR: 'Dữ liệu không hợp lệ. Hãy kiểm tra loại tệp rồi thử lại.',
        INVALID_STATE: 'Không thể thực hiện thao tác này lúc này.',
        NOT_FOUND: 'Không tìm thấy mục.',
        UPLOAD_EXPIRED: 'Phiên tải lên đã hết hạn. Hãy tải lại tệp.',
      },
    },
  },
  en: {
    fallback: 'Upload failed',
    api: {},
  },
}

export function uploadErrorCopy(locale: Locale): UploadErrorCopy {
  return copy[locale]
}
