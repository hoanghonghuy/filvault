import type { ApiErrorCopy } from '@/api/errors'
import type { Locale } from '@/lib/i18n'

export type FolderPickerCopy = {
  apiError: ApiErrorCopy
  browseFolders: string
  root: string
  parentFolder: string
  loadFailed: string
  empty: string
}

const copy: Record<Locale, FolderPickerCopy> = {
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
    browseFolders: 'Duyệt thư mục',
    root: 'Gốc',
    parentFolder: '.. Thư mục cha',
    loadFailed: 'Không thể tải thư mục',
    empty: 'Không có thư mục con.',
  },
  en: {
    apiError: {},
    browseFolders: 'Browse folders',
    root: 'Root',
    parentFolder: '.. Parent folder',
    loadFailed: 'Failed to load folders',
    empty: 'No subfolders here.',
  },
}

export function folderPickerCopy(locale: Locale): FolderPickerCopy {
  return copy[locale]
}
