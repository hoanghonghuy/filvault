import type { ApiErrorCopy } from '@/api/errors'
import type { Locale } from '@/lib/i18n'

export type AlbumRuntimeCopy = {
  apiError: ApiErrorCopy
  loadFailed: string
  addPhotos: string
  renameAlbum: string
  deleteAlbum: string
  renameFailed: string
  deleteFailed: string
  addItemFailed: string
  download: string
  removeCover: string
  setCover: string
  removeFromAlbum: string
  coverUpdated: string
  setCoverFailed: string
  coverReset: string
  removeCoverFailed: string
  removeConfirmation: (name: string) => string
  removeFailed: string
  viewFailed: string
  downloadFailed: string
}

const copy: Record<Locale, AlbumRuntimeCopy> = {
  vi: {
    apiError: {
      network: 'Không thể kết nối. Hãy kiểm tra mạng và thử lại.',
      codes: {
        unauthorized: 'Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.',
        forbidden: 'Bạn không có quyền thực hiện thao tác này.',
        not_found: 'Không tìm thấy nội dung này hoặc nội dung đã bị xóa.',
        conflict: 'Nội dung đã thay đổi. Hãy tải lại và thử lại.',
        rate_limited: 'Bạn thao tác quá nhanh. Vui lòng thử lại sau.',
      },
    },
    loadFailed: 'Không thể tải album',
    addPhotos: 'Thêm ảnh',
    renameAlbum: 'Đổi tên album',
    deleteAlbum: 'Xóa album',
    renameFailed: 'Không thể đổi tên album',
    deleteFailed: 'Không thể xóa album',
    addItemFailed: 'Không thể thêm vào album',
    download: 'Tải xuống',
    removeCover: 'Bỏ ảnh bìa',
    setCover: 'Đặt làm ảnh bìa',
    removeFromAlbum: 'Xóa khỏi album này',
    coverUpdated: 'Đã cập nhật ảnh bìa',
    setCoverFailed: 'Không thể đặt ảnh bìa',
    coverReset: 'Đã đặt ảnh bìa về tự động',
    removeCoverFailed: 'Không thể bỏ ảnh bìa',
    removeConfirmation: (name) => `"${name}" sẽ chỉ bị xóa khỏi album này, tệp gốc vẫn được giữ.`,
    removeFailed: 'Không thể xóa khỏi album',
    viewFailed: 'Không thể mở nội dung',
    downloadFailed: 'Không thể tải xuống',
  },
  en: {
    apiError: {},
    loadFailed: 'Failed to load album',
    addPhotos: 'Add photos',
    renameAlbum: 'Rename album',
    deleteAlbum: 'Delete album',
    renameFailed: 'Rename failed',
    deleteFailed: 'Delete failed',
    addItemFailed: 'Failed to add item',
    download: 'Download',
    removeCover: 'Remove cover',
    setCover: 'Set cover',
    removeFromAlbum: 'Remove from this album',
    coverUpdated: 'Cover updated',
    setCoverFailed: 'Failed to set cover',
    coverReset: 'Cover reset to automatic',
    removeCoverFailed: 'Failed to remove cover',
    removeConfirmation: (name) => `"${name}" will be removed from this album only. The original file will remain in your library.`,
    removeFailed: 'Remove failed',
    viewFailed: 'View failed',
    downloadFailed: 'Download failed',
  },
}

export function albumRuntimeCopy(locale: Locale): AlbumRuntimeCopy {
  return copy[locale]
}
